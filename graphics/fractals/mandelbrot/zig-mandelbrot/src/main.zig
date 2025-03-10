const std = @import("std");
const jok = @import("jok");
const j2d = jok.j2d;
const zm = @import("zm");
const Complex = @import("std").math.complex.Complex;

var batchpool: j2d.BatchPool(64, false) = undefined;
var texture: jok.Texture = undefined;
var pixels: jok.Texture.PixelData = undefined;

const mandelbrotMin = zm.Vec2{ -2.5, -1.0 };
const mandelbrotMax = zm.Vec2{ 1.0, 1.0 };

const zoomPercent = 0.1;
const zoomMultiplier = 1.0 - zoomPercent;
var userScale = zm.Mat3 {
    .data = .{
        mandelbrotMax[0] - mandelbrotMin[0], 0.0, mandelbrotMin[0],
        0.0, mandelbrotMax[1] - mandelbrotMin[1], mandelbrotMin[1],
        0.0, 0.0, 1.0,
    },
};

fn projectMat3(mat: zm.Mat3, point: zm.Vec2) zm.Vec2 {
    return zm.Vec2{
        mat.data[0] * point[0] + mat.data[1] * point[1] + mat.data[2],
        mat.data[3] * point[0] + mat.data[4] * point[1] + mat.data[5],
    };
}

fn mat3ScaledXY(mat: zm.Mat3, around: zm.Vec2, scale: zm.Vec2) zm.Mat3 {
    const preX = mat.data[2] - around[0];
    const preY = mat.data[5] - around[1];

    return zm.Mat3{
        .data = .{
            mat.data[0] * scale[0], mat.data[1] * scale[0], preX * scale[0] + around[0],
            mat.data[3] * scale[1], mat.data[4] * scale[1], preY * scale[1] + around[1],
            0.0, 0.0, 1.0,
        },
    };
}

pub fn init(ctx: jok.Context) !void {
    batchpool = try @TypeOf(batchpool).init(ctx);
    texture = try ctx.renderer().createTexture(.{ .width = 640, .height = 480 }, null, .{
        .access = .streaming,
    });
    pixels = try texture.createPixelData(ctx.allocator(), null);

    try render(ctx);
}

fn getImageSize(ctx: jok.Context) [2]usize {
    const csz = ctx.cfg().jok_canvas_size orelse jok.Size{ .width = 640, .height = 480 };
    return .{
        csz.width,
        csz.height,
    };
}

fn render(ctx: jok.Context) !void {
    const imgSize = getImageSize(ctx);

    for (0..imgSize[1]) |row| {
        const yScale = @as(f64, @as(f64, @floatFromInt(row))) / @as(f64, @floatFromInt(imgSize[1]));

        for (0..imgSize[0]) |col| {
            const xScale = @as(f64, @as(f64, @floatFromInt(col))) / @as(f64, @floatFromInt(imgSize[0]));
            const point = projectMat3(userScale, zm.Vec2{xScale, yScale});

            const escapeCount = escapeTime(Complex(f64){ .re = point[0], .im = point[1] }, 255);
            const intensity: u8 = @truncate(255 - (escapeCount orelse 255));

            pixels.setPixel(@truncate(col), @truncate(row), jok.Color.rgb(intensity, intensity, intensity));
        }
    }

    try texture.update(pixels);
}

fn calculateZoom(ctx: jok.Context, mousePixel: jok.Point, zoomScale: zm.Vec2) zm.Mat3 {
    const imgSize = getImageSize(ctx);

    const mouseScale = zm.Vec2{
        @as(f64, @as(f64, mousePixel.x)) / @as(f64, @floatFromInt(imgSize[0])),
        @as(f64, @as(f64, mousePixel.y)) / @as(f64, @floatFromInt(imgSize[1])),
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
                scaleFactor = std.math.pow(f64, 1.0 + zoomPercent, -@as(f64, @floatFromInt(-mouse_event.delta_y)));
            }

            if (scaleFactor) |s| {
                const mouseState = jok.io.getMouseState();
                userScale = calculateZoom(ctx, mouseState.pos, zm.vec.scale(zm.Vec2{1.0, 1.0}, s));

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
    try ctx.renderer().drawTexture(texture, null, null);
}

pub fn quit(ctx: jok.Context) void {
    // your deinit code
    _ = ctx;

    batchpool.deinit();
    pixels.destroy();
    texture.destroy();
}


// ref: https://www.rdiachenko.com/posts/zig/exploring-ziglang-with-mandelbrot-set/
fn escapeTime(c: Complex(f64), limit: usize) ?usize {
    var z = Complex(f64){ .re = 0.0, .im = 0.0 };
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
    const c = Complex(f64){ .re = 1.0, .im = 1.0 };
    const result = escapeTime(c, limit);
    try std.testing.expect(result != null);
}

test "expect point stays within the Mandelbrot set" {
    const limit = 1000;
    const c = Complex(f64){ .re = 0.0, .im = 0.0 };
    const result = escapeTime(c, limit);
    try std.testing.expect(result == null);
}
