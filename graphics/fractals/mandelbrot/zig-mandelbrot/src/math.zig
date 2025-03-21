const std = @import("std");
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

pub fn RectangleBase(comptime intType: type) type {
    const BorderIterator = struct {
        const Self = @This();
        const Vec2 = @Vector(2, intType);

        x: intType,
        y: intType,
        width: intType,
        height: intType,

        c_x: intType,
        c_y: intType,
        direction: u8 = 0, // 0: unstarted, 1: right, 2: down, 3: left, 4: up

        pub fn next(self: *Self) ?Vec2 {
            if (self.width == 0 or self.height == 0) {
                return null; // no points to iterate
            }

            if (self.direction == 0) {
                self.c_x = self.x;
                self.c_y = self.y;
                self.direction = 1; // start moving to the right

                return Vec2{ self.c_x, self.c_y };
            }

            if (self.direction == 1) { // move right
                if (self.c_x < self.x + self.width - 1) {
                    self.c_x += 1;
                    return Vec2{ self.c_x, self.c_y };
                } else {
                    self.direction = 2; // move down
                }
            }

            if (self.height == 1) {
                // A special case for single row rectangles
                self.direction = 0;
                return null;
            }

            if (self.direction == 2) { // move down
                if (self.c_y < self.y + self.height - 1) {
                    self.c_y += 1;
                    return Vec2{ self.c_x, self.c_y };
                } else {
                    self.direction = 3; // move left
                }
            }

            if (self.width == 1) {
                // A special case for single column rectangles
                self.direction = 0;
                return null;
            }

            if (self.direction == 3) { // move left
                if (self.c_x > self.x) {
                    self.c_x -= 1;
                    return Vec2{ self.c_x, self.c_y };
                } else {
                    self.direction = 4; // move up
                }
            }

            if (self.direction == 4) { // move up
                if (self.c_y > self.y + 1) {
                    self.c_y -= 1;
                    return Vec2{ self.c_x, self.c_y };
                } else {
                    self.direction = 0; // reset to unstarted
                }
            }

            return null;
        }
    };

    return struct {
        const Self = @This();

        x: intType = 0,
        y: intType = 0,
        width: intType = 0,
        height: intType = 0,

        /// Returns the area of the rectangle.
        pub fn area(self: Self) intType {
            return self.width * self.height;
        }

        /// Split the rectangle into two equal halves,
        /// splitting vertically if the width is greater than or equal to the height,
        /// otherwise splitting horizontally.
        pub fn split(self: Self) struct { Self, Self } {
            if (self.width >= self.height) {
                return self.splitVertical();
            } else {
                return self.splitHorizontal();
            }
        }

        /// Splits the rectangle into two equal halves horizontally.
        /// If the rectangle has an odd height, the top half will be one unit larger.
        pub fn splitHorizontal(self: Self) struct { Self, Self } {
            const half_height = self.height / 2;
            const top = Self{
                .x = self.x,
                .y = self.y,
                .width = self.width,
                .height = self.height - half_height,
            };
            const bottom = Self{
                .x = self.x,
                .y = top.y + top.height,
                .width = self.width,
                .height = half_height,
            };
            return .{ top, bottom };
        }

        /// Splits the rectangle into two equal halves vertically.
        /// If the rectangle has an odd width, the left half will be one unit larger.
        pub fn splitVertical(self: Self) struct { Self, Self } {
            const half_width = self.width / 2;
            const left = Self{
                .x = self.x,
                .y = self.y,
                .width = self.width - half_width,
                .height = self.height,
            };
            const right = Self{
                .x = left.x + left.width,
                .y = self.y,
                .width = half_width,
                .height = self.height,
            };
            return .{ left, right };
        }

        /// Shrinks the rectangle by a given amount on all sides.
        /// Returns a new rectangle with the reduced dimensions.
        /// If the shrink amount is greater than half the width or height, the rectangle will collapse to zero size.
        pub fn shrink(self: Self, amount: intType) Self {
            const amount_2 = amount * 2;
            const new_width = if (self.width > amount_2) self.width - amount_2 else 0;
            const new_height = if (self.height > amount_2) self.height - amount_2 else 0;
            return Self{
                .x = self.x + amount,
                .y = self.y + amount,
                .width = new_width,
                .height = new_height,
            };
        }

        /// Iterate over the coordinates on rectangle's borders, from top-left going clockwise.
        /// Returns an iterator that yields the coordinates as tuples of (x, y).
        pub fn iterateBorders(self: Self) BorderIterator {
            return BorderIterator{
                .x = self.x,
                .y = self.y,
                .width = self.width,
                .height = self.height,
                .c_x = self.x,
                .c_y = self.y,
            };
        }
    };
}

pub const u32Rectangle = RectangleBase(u32);

test {
    _ = @import("math_test.zig");
}
