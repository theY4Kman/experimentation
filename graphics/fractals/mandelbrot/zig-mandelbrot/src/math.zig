const math = @import("std").math;
const zm = @import("zm");

pub fn MathTypes(comptime FloatType: type) type {
    return struct {
        const Self = @This();

        pub const Float = FloatType;
        pub const Complex = math.complex.Complex(FloatType);
        pub const Vec2 = @Vector(2, FloatType);
        pub const Mat3 = zm.matrix.Mat3Base(FloatType);
    };
}
