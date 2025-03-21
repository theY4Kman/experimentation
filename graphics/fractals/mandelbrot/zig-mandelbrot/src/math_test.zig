const std = @import("std");
const ourMath = @import("math.zig");
const u32Rectangle = ourMath.u32Rectangle;

test "Rectangle.area()" {
    const rect = u32Rectangle{ .width = 10, .height = 20 };
    try std.testing.expectEqual(rect.area(), 200);
}

test "Rectangle.splitHorizontal()" {
    {
        const rect = u32Rectangle{ .width = 10, .height = 10 };
        const top, const bottom = rect.splitHorizontal();
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 0, .y = 0, .width = 10, .height = 5 }, top);
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 0, .y = 5, .width = 10, .height = 5 }, bottom);
    }
    {
        const rect = u32Rectangle{ .width = 10, .height = 9 };
        const top, const bottom = rect.splitHorizontal();
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 0, .y = 0, .width = 10, .height = 5 }, top);
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 0, .y = 5, .width = 10, .height = 4 }, bottom);
    }
}

test "Rectangle.splitVertical()" {
    {
        const rect = u32Rectangle{ .width = 10, .height = 10 };
        const left, const right = rect.splitVertical();
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 0, .y = 0, .width = 5, .height = 10 }, left);
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 5, .y = 0, .width = 5, .height = 10 }, right);
    }
    {
        const rect = u32Rectangle{ .width = 9, .height = 10 };
        const left, const right = rect.splitVertical();
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 0, .y = 0, .width = 5, .height = 10 }, left);
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 5, .y = 0, .width = 4, .height = 10 }, right);
    }
}

test "Rectangle.shrink()" {
    {
        const rect = u32Rectangle{ .x = 2, .y = 2, .width = 10, .height = 10 };
        const shrunk = rect.shrink(2);
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 4, .y = 4, .width = 6, .height = 6 }, shrunk);
    }
    {
        const rect = u32Rectangle{ .x = 2, .y = 2, .width = 10, .height = 10 };
        const shrunk = rect.shrink(5);
        try std.testing.expectEqualDeep(u32Rectangle{ .x = 7, .y = 7, .width = 0, .height = 0 }, shrunk);
    }
}

const u32Vec2 = @Vector(2, u32);

fn testIterateBorders(
    rect: u32Rectangle,
    expected_points: []const u32Vec2,
) !void {
    var points = std.ArrayList(u32Vec2).init(std.testing.allocator);
    defer points.deinit();

    var iter = rect.iterateBorders();
    while (iter.next()) |point| {
        try points.append(point);
    }

    const pointsSlice = try points.toOwnedSlice();
    defer std.testing.allocator.free(pointsSlice);

    try std.testing.expectEqualSlices(u32Vec2, expected_points, pointsSlice);
}

test "Rectangle.iterateBorders()" {
    {
        try testIterateBorders(
            u32Rectangle{ .x = 0, .y = 0, .width = 0, .height = 0 },
            &.{},
        );
    }
    {
        try testIterateBorders(
            u32Rectangle{ .x = 0, .y = 0, .width = 1, .height = 1 },
            &.{
                u32Vec2{ 0, 0 },
            },
        );
    }
    {
        try testIterateBorders(
            u32Rectangle{ .x = 0, .y = 0, .width = 5, .height = 4 },
            &.{
                u32Vec2{ 0, 0 }, u32Vec2{ 1, 0 }, u32Vec2{ 2, 0 }, u32Vec2{ 3, 0 }, u32Vec2{ 4, 0 },
                u32Vec2{ 4, 1 }, u32Vec2{ 4, 2 }, u32Vec2{ 4, 3 },
                u32Vec2{ 3, 3 }, u32Vec2{ 2, 3 }, u32Vec2{ 1, 3 }, u32Vec2{ 0, 3 },
                u32Vec2{ 0, 2 }, u32Vec2{ 0, 1 }
            },
        );
    }
    {
        try testIterateBorders(
            u32Rectangle{ .x = 0, .y = 0, .width = 3, .height = 2 },
            &.{
                u32Vec2{ 0, 0 }, u32Vec2{ 1, 0 }, u32Vec2{ 2, 0 },
                u32Vec2{ 2, 1 }, u32Vec2{ 1, 1 }, u32Vec2{ 0, 1 }
            },
        );
    }
}
