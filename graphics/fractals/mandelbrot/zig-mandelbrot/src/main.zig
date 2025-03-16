const std = @import("std");
const testing = std.testing;
const jok = @import("jok");
const sdl = jok.sdl;
const j2d = jok.j2d;
const zm = @import("zm");
const spice = @import("spice");
const mt = @import("math.zig").MathTypes(f64);
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

var window_size: jok.Size = .{ .width = 0, .height = 0 };

const mandelbrotMin = mt.Vec2{ -2.5, -1.0 };
const mandelbrotMax = mt.Vec2{ 1.0, 1.0 };

const MAX_ITERATIONS = 1024;
const RENDER_WORKER_BATCH_SIZE: u32 = 1<<11;

const baseColorGrad = ColorGradient(usize).init(
    &.{
        // NB(zk): original colours from go-mandelbrot
        // .{@intFromFloat(0.0000 * MAX_ITERATIONS), "#000635"},
        // .{@intFromFloat(0.0078 * MAX_ITERATIONS), "#13399a"},
        // .{@intFromFloat(0.0392 * MAX_ITERATIONS), "#1360d4"},
        // .{@intFromFloat(0.1176 * MAX_ITERATIONS), "#ffffff"},
        // .{@intFromFloat(0.3921 * MAX_ITERATIONS), "#ff7f0e"},
        // .{@intFromFloat(0.7843 * MAX_ITERATIONS), "#653608"},
        // .{MAX_ITERATIONS, "#000000"},

        .{@intFromFloat(0.0000 * MAX_ITERATIONS), "#000764"},
        .{@intFromFloat(0.1600 * MAX_ITERATIONS), "#206bcb"},
        .{@intFromFloat(0.4200 * MAX_ITERATIONS), "#edffff"},
        .{@intFromFloat(0.6425 * MAX_ITERATIONS), "#ffaa00"},
        .{@intFromFloat(0.8575 * MAX_ITERATIONS), "#000200"},
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

    render_thread = try std.Thread.spawn(.{}, renderWorker, .{ ctx });

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

const RenderWorkerParams = struct {
    /// Index of pixel to start rendering from
    from: u32,

    /// Index of pixel to stop rendering at
    to: u32,

    /// Width of the image
    width: u32,

    /// Height of the image
    height: u32,

    /// Transform matrix (converts ratios between 0.0 and 1.0 to coords in the Mandelbrot set)
    scale: mt.Mat3,

    inline fn size(self: RenderWorkerParams) u32 {
        return self.to - self.from;
    }

    inline fn middle(self: RenderWorkerParams) u32 {
        return (self.from + self.to) / 2;
    }

    fn splitHigh(self: *RenderWorkerParams) ?RenderWorkerParams {
        if (self.from + RENDER_WORKER_BATCH_SIZE >= self.to) {
            return null;
        }

        const origTo = self.to;
        self.to = self.middle();

        return RenderWorkerParams{
            .from = self.to,
            .to = origTo,
            .width = self.width,
            .height = self.height,
            .scale = self.scale,
        };
    }

    fn splitLowBatch(self: *RenderWorkerParams) ?RenderWorkerParams {
        if (self.from + RENDER_WORKER_BATCH_SIZE >= self.to) {
            return null;
        }

        const origFrom = self.from;
        self.from += RENDER_WORKER_BATCH_SIZE;

        return RenderWorkerParams{
            .from = origFrom,
            .to = self.from,
            .width = self.width,
            .height = self.height,
            .scale = self.scale,
        };
    }

    inline fn projectIndex(self: RenderWorkerParams, index: u32) mt.Complex {
        const rowCol = self.indexToRowCol(index);
        return self.projectRowCol(rowCol);
    }

    inline fn projectRowCol(self: RenderWorkerParams, rowCol: RowCol) mt.Complex {
        const xScale = @as(mt.Float, @as(mt.Float, @floatFromInt(rowCol.col))) / @as(mt.Float, @floatFromInt(self.width));
        const yScale = @as(mt.Float, @as(mt.Float, @floatFromInt(rowCol.row))) / @as(mt.Float, @floatFromInt(self.height));
        const vec = projectMat3(self.scale, mt.Vec2{ xScale, yScale });
        return mt.Complex{ .re = vec[0], .im = vec[1] };
    }

    inline fn indexToRowCol(self: RenderWorkerParams, index: u32) RowCol {
        return RowCol{
            .row = index / self.width,
            .col = index % self.width,
        };
    }
};

fn renderWorker(ctx: jok.Context) !void {
    _ = ctx;

    thread_pool = spice.ThreadPool.init(allocator);
    thread_pool.start(.{
        .background_worker_count = 16,
    });

    while (!render_thread_canceled) {
        render_requested.wait();
        render_requested.reset();

        if (render_thread_canceled) {
            return;
        }

        doRenderParallel();
    }
}

fn doRenderParallel() void {
    pixels_mutex.lock();
    defer pixels_mutex.unlock();

    var params = RenderWorkerParams{
        .from = 0,
        .to = window_size.width * window_size.height,
        .width = window_size.width,
        .height = window_size.height,
        .scale = userScale,
    };

    thread_pool.call(void, doRenderParallelWorker, &params);
}

const RenderWorkerFuture = spice.Future(*RenderWorkerParams, void);


fn doRenderParallelWorker(t: *spice.Task, params: *RenderWorkerParams) void {
    const origTo = params.to;

    while (params.from < params.to) {
        var highFut = RenderWorkerFuture.init();
        var highParams: RenderWorkerParams = undefined;

        var hasBatch1 = false;
        var batch1Fut: RenderWorkerFuture = RenderWorkerFuture.init();
        var batch1Params: RenderWorkerParams = undefined;

        if (params.splitHigh()) |next| {
            highParams = next;
            highFut.fork(t, doRenderParallelWorker, &highParams);
        }

        if (params.splitLowBatch()) |batch| {
            hasBatch1 = true;
            batch1Params = batch;
            batch1Fut.fork(t, doRenderParallelWorker, &batch1Params);
        }

        const endIndex = @min(params.to, params.from + RENDER_WORKER_BATCH_SIZE);
        while (params.from < endIndex) : (params.from += 1) {
            const rowCol = params.indexToRowCol(params.from);
            const point = params.projectRowCol(rowCol);
            const escapeCount = t.call(?usize, calculateEscapeCount, point) orelse MAX_ITERATIONS;

            const smoothed = std.math.log2(std.math.log2(point.squaredMagnitude() + 1.0) / 2.0);
            const colorIdxf = std.math.sqrt(@as(mt.Float, @floatFromInt(escapeCount)) + 10.0 - smoothed) * @as(mt.Float, @floatFromInt(MAX_ITERATIONS));
            const colorIdx = @as(usize, @intFromFloat(colorIdxf)) % MAX_ITERATIONS;
            pixels.?.setPixel(rowCol.col, rowCol.row, colorGrad.at(colorIdx));
        }

        if (hasBatch1 and (batch1Fut.tryJoin(t) == null)) {
            t.call(void, doRenderParallelWorker, &batch1Params);
        }

        if (highFut.tryJoin(t)) |_| {
            t.call(void, doRenderParallelWorker, params);
            return;
        } else {
            params.to = origTo;
        }
    }
}

inline fn cast(comptime T: type, value: anytype) T {
    return @intCast(value);
}

fn calculateEscapeCount(t: *spice.Task, point: mt.Complex) ?usize {
    _ = t;
    return escapeTime(point, MAX_ITERATIONS);
}

fn calculateZoom(mousePixel: jok.Point, zoomScale: mt.Vec2) mt.Mat3 {
    const mouseScale = mt.Vec2{
        @as(mt.Float, @as(mt.Float, mousePixel.x)) / @as(mt.Float, @floatFromInt(window_size.width)),
        @as(mt.Float, @as(mt.Float, mousePixel.y)) / @as(mt.Float, @floatFromInt(window_size.height)),
    };
    const mousePoint = projectMat3(userScale, mouseScale);

    return mat3ScaledXY(userScale, mousePoint, zoomScale);
}

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

                //XXX///////////////////////////////////////////////////////////////////////////////////////////
                //XXX///////////////////////////////////////////////////////////////////////////////////////////
                std.debug.print("{d} — ({d:6.5},{d:6.5}) to ({d:6.5},{d:6.5})\n", .{ totalZoom, userScale.data[2], userScale.data[5], userScale.data[0] + userScale.data[2], userScale.data[4] + userScale.data[5] });
                //XXX///////////////////////////////////////////////////////////////////////////////////////////

                try render(ctx);
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

pub fn update(ctx: jok.Context) !void {
    // your game state updating code
    _ = ctx;
}

pub fn draw(ctx: jok.Context) !void {
    if (texture) |t| {
        if (pixels) |p| {
            try t.update(p);
        }
        try ctx.renderer().drawTexture(t, null, null);
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
        if (z.squaredMagnitude() > 4.0) {
            return i;
        }
        z = z.mul(z).add(c);
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
