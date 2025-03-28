const std = @import("std");
const testing = std.testing;
const jok = @import("jok");
const sdl = jok.sdl;
const j2d = jok.j2d;
const zm = @import("zm");
const spice = @import("spice");

const ourMath = @import("math.zig");
const mt = ourMath.MathTypes(f64);
const u32Rectangle = ourMath.u32Rectangle;
const ColorGradient = @import("color.zig").ColorGradient;

var allocator: std.mem.Allocator = undefined;

var thread_pool: spice.ThreadPool = undefined;
var render_requested = std.Thread.ResetEvent{};
var render_thread: ?std.Thread = null;
var render_thread_canceled = false;

var batchpool: j2d.BatchPool(64, false) = undefined;
var texture: ?jok.Texture = null;
var pixels: ?jok.Texture.PixelData = null;
var pixels_mutex = std.Thread.Mutex{};
var pixels_changed = true;

var window_size: jok.Size = .{ .width = 0, .height = 0 };

const mandelbrotMin = mt.Vec2{ -2.5, -1.0 };
const mandelbrotMax = mt.Vec2{ 1.0, 1.0 };

const MAX_ITERATIONS = 1024;
const RENDER_WORKER_RECT_BATCH_AREA: u32 = 1024;
const RENDER_WORKER_RECT_BATCH_SHRINK_STEP: u32 = 2;
const RENDER_WORKER_RECT_BATCH_MAX_SPLITS: u32 = 5;

const baseColorGrad = ColorGradient(usize).init(
    &.{
        .{ @intFromFloat(0.0000 * MAX_ITERATIONS), "#000635" },
        .{ @intFromFloat(0.0002 * MAX_ITERATIONS), "#000764" },
        .{ @intFromFloat(0.0078 * MAX_ITERATIONS), "#13399a" },
        .{ @intFromFloat(0.0392 * MAX_ITERATIONS), "#1360d4" },
        .{ @intFromFloat(0.1600 * MAX_ITERATIONS), "#206bcb" },
        .{ @intFromFloat(0.4200 * MAX_ITERATIONS), "#edffff" },
        .{ @intFromFloat(0.6425 * MAX_ITERATIONS), "#ffaa00" },
        .{ @intFromFloat(0.8575 * MAX_ITERATIONS), "#000200" },
    },
) catch ColorGradient(u8).GREYSCALE;
var colorGrad = baseColorGrad.compute(2048);

const zoomPercent = 0.1;
const zoomMultiplier = 1.0 - zoomPercent;
const baseScale = mt.Mat3{
    .data = .{
        mandelbrotMax[0] - mandelbrotMin[0], 0.0,                                 mandelbrotMin[0],
        0.0,                                 mandelbrotMax[1] - mandelbrotMin[1], mandelbrotMin[1],
        0.0,                                 0.0,                                 1.0,
    },
};
var userScale = baseScale;
var totalZoom: i32 = 0;

fn projectMat3(mat: mt.Mat3, point: mt.Vec2) mt.Vec2 {
    return mt.Vec2{
        mat.data[0] * point[0] + mat.data[1] * point[1] + mat.data[2],
        mat.data[3] * point[0] + mat.data[4] * point[1] + mat.data[5],
    };
}

fn mat3ScaledXY(mat: mt.Mat3, around: mt.Vec2, scale: mt.Vec2) mt.Mat3 {
    const preX = mat.data[2] - around[0];
    const preY = mat.data[5] - around[1];

    return mt.Mat3{
        .data = .{
            mat.data[0] * scale[0], mat.data[1] * scale[0], preX * scale[0] + around[0],
            mat.data[3] * scale[1], mat.data[4] * scale[1], preY * scale[1] + around[1],
            0.0,                    0.0,                    1.0,
        },
    };
}

pub fn init(ctx: jok.Context) !void {
    allocator = ctx.allocator();

    const window = ctx.window();
    window.setTitle("Interactive Mandelbrot Set");
    window.setSize(jok.Size{ .width = 1024, .height = 768 });
    window.setResizable(true);
    window_size = ctx.window().getSize();

    render_thread = try std.Thread.spawn(.{}, renderWorker, .{ctx});

    batchpool = try @TypeOf(batchpool).init(ctx);
    try resizeTexture(ctx);

    try render(ctx);
}

fn resizeTexture(ctx: jok.Context) !void {
    pixels_mutex.lock();
    defer pixels_mutex.unlock();

    const prev_pixels = pixels;
    const prev_texture = texture;

    texture = try ctx.renderer().createTexture(.{ .width = @truncate(window_size.width), .height = @truncate(window_size.height) }, null, .{
        .access = .streaming,
    });
    pixels = try texture.?.createPixelData(ctx.allocator(), null);

    if (prev_pixels) |p| {
        p.destroy();
    }
    if (prev_texture) |t| {
        t.destroy();
    }
}

fn render(ctx: jok.Context) !void {
    _ = ctx;

    render_requested.set();
}

const RowCol = struct {
    row: u32,
    col: u32,
};

const RectArrayList = std.ArrayList(u32Rectangle);

const RenderWorkerRectParams = struct {
    /// Rectangles the worker is managing
    rects: RectArrayList,

    /// Width of the image
    width: u32,

    /// Height of the image
    height: u32,

    /// Transform matrix (converts ratios between 0.0 and 1.0 to coords in the Mandelbrot set)
    scale: mt.Mat3,

    pub inline fn allocator(self: RenderWorkerRectParams) *const std.mem.Allocator {
        return &self.rects.allocator;
    }

    pub inline fn numRects(self: RenderWorkerRectParams) u64 {
        return self.rects.items.len;
    }

    pub inline fn isEmpty(self: RenderWorkerRectParams) bool {
        return self.numRects() == 0;
    }

    pub fn deinit(self: RenderWorkerRectParams) void {
        self.rects.deinit();
    }

    pub fn forkChild(self: *RenderWorkerRectParams, rect: u32Rectangle) !RenderWorkerRectParams {
        var rects = try RectArrayList.initCapacity(self.rects.allocator, 1);
        rects.appendAssumeCapacity(rect);
        return RenderWorkerRectParams{
            .rects = rects,
            .width = self.width,
            .height = self.height,
            .scale = self.scale,
        };
    }

    pub fn forkChildren(self: *RenderWorkerRectParams, rects: []const u32Rectangle) !RenderWorkerRectParams {
        var rectsList = try RectArrayList.initCapacity(self.rects.allocator, rects.len);
        rectsList.appendSliceAssumeCapacity(rects);
        return RenderWorkerRectParams{
            .rects = rectsList,
            .width = self.width,
            .height = self.height,
            .scale = self.scale,
        };
    }

    fn _rectIsLessThan(_: void, lhs: u32Rectangle, rhs: u32Rectangle) bool {
        return lhs.area() < rhs.area();
    }

    pub fn popBatch(self: *RenderWorkerRectParams) ![]u32Rectangle {
        if (self.isEmpty()) {
            return &.{};
        }

        var batch = batch: {
            var rectsList: RectArrayList = undefined;

            if (self.numRects() <= 3) {
                rectsList = self.rects;
                self.rects = RectArrayList.init(rectsList.allocator);

                var i: usize = 0;
                while (i < rectsList.items.len) {
                    if (rectsList.items[i].area() <= 0) {
                        _ = rectsList.swapRemove(i);
                    } else {
                        i += 1;
                    }
                }
            } else {
                rectsList = try RectArrayList.initCapacity(self.allocator().*, 3);

                while (self.rects.items.len > 0 and rectsList.items.len < 3) {
                    const rect = self.rects.swapRemove(0);
                    if (rect.area() > 0) {
                        rectsList.appendAssumeCapacity(rect);
                    }
                }
            }

            break :batch rectsList;
        };

        if (batch.items.len == 0) {
            batch.deinit();
            return &.{};
        }

        try batch.ensureTotalCapacityPrecise(3);

        std.mem.sortUnstable(u32Rectangle, batch.items, {}, _rectIsLessThan);

        if (batch.items.len == 1 and batch.items[0].area() > RENDER_WORKER_RECT_BATCH_AREA) {
            batch.items.len = 2;
            batch.items[0], batch.items[1] = batch.items[0].split();
        }

        if (batch.items.len < 3) {
            if (batch.items.len == 2 and batch.items[1].area() > RENDER_WORKER_RECT_BATCH_AREA) {
                batch.items.len = 3;
                batch.items[1], batch.items[2] = batch.items[1].split();
            } else if (batch.items[0].area() > RENDER_WORKER_RECT_BATCH_AREA) {
                batch.items.len = 3;
                batch.items[0], batch.items[2] = batch.items[0].split();
            }
        }

        return try batch.toOwnedSlice();
    }
};

fn renderWorker(ctx: jok.Context) !void {
    _ = ctx;

    thread_pool = spice.ThreadPool.init(allocator);
    thread_pool.start(.{
        .background_worker_count = 24,
        .heartbeat_interval = 10 * std.time.ns_per_us,
    });

    while (!render_thread_canceled) {
        render_requested.wait();
        render_requested.reset();

        if (render_thread_canceled) {
            return;
        }

        doRenderParallel() catch |err| {
            // handle error
            std.debug.print("Render error: {}\n", .{err});
            continue;
        };
    }
}

fn doRenderParallel() !void {
    pixels_mutex.lock();
    defer pixels_mutex.unlock();

    const buffer: []u8 = @ptrCast(try allocator.alloc(u32Rectangle, (window_size.width * window_size.height / RENDER_WORKER_RECT_BATCH_AREA) * 15));
    defer allocator.free(buffer);

    var fba = std.heap.FixedBufferAllocator.init(buffer);
    const workerAllocator = fba.threadSafeAllocator();

    var params = RenderWorkerRectParams{
        .rects = RectArrayList.init(workerAllocator),
        .width = window_size.width,
        .height = window_size.height,
        .scale = userScale,
    };
    try params.rects.append(u32Rectangle{
        .x = 0,
        .y = 0,
        .width = window_size.width,
        .height = window_size.height,
    });

    thread_pool.call(void, doRenderParallelWorker, &params);
}

const RenderWorkerRectFuture = spice.Future(*RenderWorkerRectParams, void);

fn doRenderParallelWorker(t: *spice.Task, params: *RenderWorkerRectParams) void {
    while (!params.isEmpty()) {
        const batch = params.popBatch() catch |err| {
            std.debug.print("Error popping batch: {}\n", .{err});
            break;
        };
        if (batch.len == 0) {
            break;
        }
        defer params.allocator().free(batch);

        var smallRect: u32Rectangle = batch[0];

        var hasBigBatch = false;
        var bigParams: RenderWorkerRectParams = undefined;
        var bigRect: ?u32Rectangle = null;
        var bigFut = RenderWorkerRectFuture.init();

        var hasMedBatch = false;
        var medRect: ?u32Rectangle = null;
        var medParams: RenderWorkerRectParams = undefined;
        var medFut = RenderWorkerRectFuture.init();

        if (batch.len > 1) {
            smallRect = batch[1];
            medRect = batch[0];
            hasMedBatch = true;
        }

        if (batch.len > 2) {
            smallRect = batch[2];
            medRect = batch[1];
            bigRect = batch[0];
            hasBigBatch = true;
        }

        if (bigRect) |rect| {
            bigParams = params.forkChild(rect) catch |err| {
                std.debug.print("Error forking bigParams: {}\n", .{err});
                return;
            };
            bigFut.fork(t, doRenderParallelWorker, &bigParams);
        }

        if (medRect) |rect| {
            medParams = params.forkChild(rect) catch |err| {
                std.debug.print("Error forking medParams: {}\n", .{err});
                return;
            };
            medFut.fork(t, doRenderParallelWorker, &medParams);
        }

        if (smallRect.area() > RENDER_WORKER_RECT_BATCH_AREA) {
            if (!fillRect(
                params.width,
                params.height,
                params.scale,
                &smallRect,
                RENDER_WORKER_RECT_BATCH_SHRINK_STEP,
            )) {
                const chunkA, const chunkB = smallRect.split();
                var smallParams = params.forkChildren(&.{ chunkA, chunkB }) catch |err| {
                    std.debug.print("Error forking smallParams: {}\n", .{err});
                    return;
                };
                t.call(void, doRenderParallelWorker, &smallParams);
            }
        } else {
            t.tick();
            _ = renderRect(
                params.width,
                params.height,
                params.scale,
                &smallRect,
                RENDER_WORKER_RECT_BATCH_MAX_SPLITS,
            );
            t.tick();
        }

        if (hasMedBatch) {
            if (medFut.join(t) == null) {
                if (medParams.rects.items[0].area() > RENDER_WORKER_RECT_BATCH_AREA) {
                    const rect = medParams.rects.pop().?;
                    const chunkA, const chunkB = rect.split();
                    medParams.rects.appendSlice(&.{ chunkA, chunkB }) catch |err| {
                        std.debug.print("Error splitting medParams: {}\n", .{err});
                        std.debug.assert(medParams.rects.capacity >= 1);
                        medParams.rects.append(rect) catch unreachable;
                    };
                }
                t.call(void, doRenderParallelWorker, &medParams);
            }
        }

        if (bigFut.tryJoin(t)) |_| {
            t.call(void, doRenderParallelWorker, params);
            return;
        } else {
            if (bigRect) |rect| {
                if (rect.area() > RENDER_WORKER_RECT_BATCH_AREA) {
                    const chunkA, const chunkB = rect.split();
                    params.rects.appendSlice(&.{ chunkA, chunkB }) catch |err| {
                        std.debug.print("Error splitting bigParams: {}\n", .{err});
                        std.debug.assert(bigParams.rects.capacity >= 1);
                        params.rects.append(rect) catch unreachable;
                    };
                } else {
                    params.rects.append(rect) catch |err| {
                        std.debug.print("Error appending bigRect: {}\n", .{err});
                    };
                }
                bigParams.deinit();
            }
        }
    }

    params.deinit();
}

fn renderRect(width: u32, height: u32, scale: mt.Mat3, rect: *u32Rectangle, max_splits: u32) void {
    if (rect.area() <= 0) {
        return;
    }

    if (max_splits == 0) {
        _ = fillRect(
            width,
            height,
            scale,
            rect,
            null,
        );
        return;
    }

    if (fillRect(
        width,
        height,
        scale,
        rect,
        RENDER_WORKER_RECT_BATCH_SHRINK_STEP,
    )) {
        return;
    }

    var chunkA, var chunkB = rect.split();
    renderRect(
        width,
        height,
        scale,
        &chunkA,
        max_splits - 1,
    );
    renderRect(
        width,
        height,
        scale,
        &chunkB,
        max_splits - 1,
    );
}

/// Fill a rectangle with the Mandelbrot set, using the given scale and max_shrink
/// Returns true if the rectangle was completely filled with a single color.
fn fillRect(width: u32, height: u32, scale: mt.Mat3, rect: *u32Rectangle, max_shrink: ?u32) bool {
    var i: u32 = 0;
    while (rect.area() > 0) : (i += 1) {
        if (max_shrink) |m| {
            if (i >= m) {
                break;
            }
        }

        var initialEscapeCount: ?usize = null;
        var areBordersHomogeneous = true;

        var borderIter = rect.iterateBorders();
        while (borderIter.next()) |rowCol| {
            const point = pointFromRowCol(rowCol, width, height, scale);
            const escapeCount = escapeTime(point, MAX_ITERATIONS) orelse MAX_ITERATIONS;

            if (initialEscapeCount == null) {
                initialEscapeCount = escapeCount;
            } else if (initialEscapeCount != escapeCount) {
                areBordersHomogeneous = false;
            }

            const color = pointColor(point, escapeCount);
            std.debug.assert(rowCol[0] < width and rowCol[1] < height);
            pixels.?.setPixel(rowCol[0], rowCol[1], color);
            pixels_changed = true;
        }

        rect.* = rect.shrink(1);

        if (areBordersHomogeneous) {
            const endY = rect.y + rect.height;
            const endX = rect.x + rect.width;

            var y: u32 = rect.y;
            while (y < endY) : (y += 1) {
                var x: u32 = rect.x;
                while (x < endX) : (x += 1) {
                    const rowCol = @Vector(2, u32){ x, y };
                    const point = pointFromRowCol(rowCol, width, height, scale);
                    const color = pointColor(point, initialEscapeCount.?);
                    std.debug.assert(rowCol[0] < width and rowCol[1] < height);
                    pixels.?.setPixel(rowCol[0], rowCol[1], color);
                    pixels_changed = true;
                }
            }
            return true;
        }
    }

    return rect.area() <= 0; // return true if we filled the rectangle completely
}

inline fn pointFromRowCol(rowCol: @Vector(2, u32), width: u32, height: u32, scale: mt.Mat3) mt.Complex {
    const xScale = @as(mt.Float, @as(mt.Float, @floatFromInt(rowCol[0]))) / @as(mt.Float, @floatFromInt(width));
    const yScale = @as(mt.Float, @as(mt.Float, @floatFromInt(rowCol[1]))) / @as(mt.Float, @floatFromInt(height));
    const vec = projectMat3(scale, mt.Vec2{ xScale, yScale });
    return mt.Complex{ .re = vec[0], .im = vec[1] };
}

inline fn pointColor(point: mt.Complex, escapeCount: usize) jok.Color {
    const smoothed = std.math.log2(std.math.log2(point.squaredMagnitude() + 1.0) / 2.0);
    const colorIdxf = std.math.sqrt(@as(mt.Float, @floatFromInt(escapeCount)) + 10.0 - smoothed) * @as(mt.Float, @floatFromInt(MAX_ITERATIONS));
    const colorIdx = @as(usize, @intFromFloat(colorIdxf)) % MAX_ITERATIONS;
    return colorGrad.at(colorIdx);
}

fn calculateZoom(mousePixel: jok.Point, zoomScale: mt.Vec2) mt.Mat3 {
    const mouseScale = mt.Vec2{
        @as(mt.Float, @as(mt.Float, mousePixel.x)) / @as(mt.Float, @floatFromInt(window_size.width)),
        @as(mt.Float, @as(mt.Float, mousePixel.y)) / @as(mt.Float, @floatFromInt(window_size.height)),
    };
    const mousePoint = projectMat3(userScale, mouseScale);

    return mat3ScaledXY(userScale, mousePoint, zoomScale);
}

fn calculatePan(startPixel: jok.Point, endPixel: jok.Point) mt.Mat3 {
    const d_x = @as(mt.Float, startPixel.x - endPixel.x);
    const d_y = @as(mt.Float, startPixel.y - endPixel.y);

    if (d_x == 0 and d_y == 0) {
        return userScale; // no movement
    }

    const pixelSize = mt.Vec2{
        userScale.data[0] / @as(mt.Float, @floatFromInt(window_size.width)),
        userScale.data[4] / @as(mt.Float, @floatFromInt(window_size.height)),
    };

    const delta = mt.Vec2{
        d_x * pixelSize[0],
        d_y * pixelSize[1],
    };

    var newScale = userScale;
    newScale.data[2] += delta[0];
    newScale.data[5] += delta[1];
    return newScale;
}

fn calculateCenterZoom(zoomScale: mt.Vec2) mt.Mat3 {
    const mouseScale = mt.Vec2{ 0.5, 0.5 };
    const centerPoint = projectMat3(userScale, mouseScale);
    return mat3ScaledXY(userScale, centerPoint, zoomScale);
}

var isDragging = false;
var dragStartPoint: jok.Point = undefined;

pub fn event(ctx: jok.Context, e: jok.Event) !void {
    switch (e) {
        .mouse_wheel => |mouse_event| {
            var scaleFactor: ?f64 = null;
            if (mouse_event.delta_y > 0) {
                scaleFactor = std.math.pow(f64, zoomMultiplier, @as(f64, @floatFromInt(mouse_event.delta_y)));
            } else if (mouse_event.delta_y < 0) {
                scaleFactor = std.math.pow(f64, 1.0 + zoomPercent, @as(f64, @floatFromInt(-mouse_event.delta_y)));
            }

            if (scaleFactor) |s| {
                const mouseState = jok.io.getMouseState();
                userScale = calculateZoom(mouseState.pos, zm.vec.scale(mt.Vec2{ 1.0, 1.0 }, s));
                totalZoom += mouse_event.delta_y;

                try onScaleChanged(ctx);
            }
        },
        .mouse_button_down => |mouse_event| {
            if (mouse_event.button == .left) {
                isDragging = true;
                dragStartPoint = mouse_event.pos;
            }
        },
        .mouse_button_up => |mouse_event| {
            if (isDragging and mouse_event.button == .left) {
                isDragging = false;
            }
        },
        .mouse_motion => |mouse_event| {
            if (isDragging) {
                userScale = calculatePan(dragStartPoint, mouse_event.pos);
                dragStartPoint = mouse_event.pos; // update the start point for the next motion
                try onScaleChanged(ctx);
            }
        },
        .key_up => |key_event| {
            switch (key_event.keycode) {
                .r => {
                    userScale = baseScale;
                    totalZoom = 0;
                    try onScaleChanged(ctx);
                },
                .plus, .equals => {
                    userScale = calculateCenterZoom(zm.vec.scale(mt.Vec2{ 1.0, 1.0 }, zoomMultiplier));
                    totalZoom += 1;
                    try onScaleChanged(ctx);
                },
                .minus => {
                    userScale = calculateCenterZoom(zm.vec.scale(mt.Vec2{ 1.0, 1.0 }, 1.0 + zoomPercent));
                    totalZoom -= 1;
                    try onScaleChanged(ctx);
                },
                .f5 => {
                    // force re-render
                    try render(ctx);
                },
                else => {},
            }
        },
        .window => |window_event| {
            if (window_event.type == .resized) {
                window_size = ctx.window().getSize();
                try resizeTexture(ctx);
                try render(ctx);
            }
        },
        else => {},
    }
}

fn onScaleChanged(ctx: jok.Context) !void {
    std.debug.print("{d} — ({d:6.7},{d:6.7}) to ({d:6.7},{d:6.7})\n", .{ totalZoom, userScale.data[2], userScale.data[5], userScale.data[0] + userScale.data[2], userScale.data[4] + userScale.data[5] });
    try render(ctx);
}

pub fn update(ctx: jok.Context) !void {
    // your game state updating code
    _ = ctx;
}

pub fn draw(ctx: jok.Context) !void {
    if (pixels_changed) {
        if (texture) |t| {
            if (pixels) |p| {
                try t.update(p);
                pixels_changed = false;
            }
            try ctx.renderer().drawTexture(t, null, null);
        }
    }
}

pub fn quit(ctx: jok.Context) void {
    // your deinit code
    _ = ctx;

    if (render_thread) |t| {
        render_thread_canceled = true;
        render_requested.set();
        t.join();
    }

    thread_pool.deinit();

    if (pixels) |p| {
        p.destroy();
    }
    if (texture) |t| {
        t.destroy();
    }

    batchpool.deinit();
}

// ref: https://www.rdiachenko.com/posts/zig/exploring-ziglang-with-mandelbrot-set/
fn escapeTime(c: mt.Complex, limit: usize) ?usize {
    var z = c;
    for (0..limit) |i| {
        const zRe2 = z.re * z.re;
        const zIm2 = z.im * z.im;
        if ((zRe2 + zIm2) > 4.0) {
            return i;
        }
        z = mt.Complex{
            .re = zRe2 - zIm2 + c.re,
            .im = 2.0 * z.re * z.im + c.im,
        };
    }
    return null;
}

test "expect point escapes the Mandelbrot set" {
    const limit = 1000;
    const c = mt.Complex{ .re = 1.0, .im = 1.0 };
    const result = escapeTime(c, limit);
    try std.testing.expect(result != null);
}

test "expect point stays within the Mandelbrot set" {
    const limit = 1000;
    const c = mt.Complex{ .re = 0.0, .im = 0.0 };
    const result = escapeTime(c, limit);
    try std.testing.expect(result == null);
}

test {
    testing.refAllDecls(@This());
}
