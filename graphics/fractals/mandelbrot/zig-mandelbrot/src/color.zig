const std = @import("std");
const testing = std.testing;
const jok = @import("jok");


/// Utility used by Hxx color-spaces for interpolating between two angles in [0,360].
/// NB(zk): stolen from go-colorful
fn interp_angle(a: f32, b: f32, t: f32) f32 {
    // Based on the answer here: http://stackoverflow.com/a/14498790/2366315
	// With potential proof that it works here: http://math.stackexchange.com/a/2144499
    const delta = @mod(@mod(b - a, 360.0) + 540.0, 360.0) - 180.0;
    return @mod(a + t * delta + 360.0, 360.0);
}


pub fn ColorGradient(comptime Domain: type) type {
    return struct {
        pub fn init(comptime domain_colors: []const struct { Domain, []const u8 }) !_ColorGradient(Domain, domain_colors.len) {
            return try _ColorGradient(Domain, domain_colors.len).init(domain_colors);
        }

        pub const GREYSCALE = _ColorGradient(Domain, 2){
            .colorsHSL = &.{
                jok.Color.rgb(0, 0, 0).toHSL(),
                jok.Color.rgb(255, 255, 255).toHSL(),
            },
            .domain = &.{ 0, 255 },
        };
    };
}


fn _ColorGradient(comptime Domain: type, comptime N: usize) type {
    return struct {
        const Self = @This();

        /// Color steps
        colorsHSL: [N][4]f32,

        /// Values within the domain describing where each color step is in the gradient
        domain: [N]Domain,

        pub fn init(comptime domain_colors: []const struct { Domain, []const u8 }) !Self {
            @setEvalBranchQuota(8000);

            var self = Self{
                .colorsHSL = undefined,
                .domain = undefined,
            };

            for (domain_colors, 0..) |domainColor, i| {
                const domain, const color = domainColor;
                self.colorsHSL[i] = (try jok.Color.parse(color)).toHSL();
                self.domain[i] = domain;
            }

            return self;
        }

        pub fn at(self: Self, value: Domain) jok.Color {
            if (self.colorsHSL.len == 1) {
                return self.colors[0];
            }

            var lowHSL = self.colorsHSL[0];
            var lowDomain = self.domain[0];

            if (value <= lowDomain) {
                return jok.Color.fromHSL(lowHSL);
            }

            var highHSL = lowHSL;
            var highDomain = lowDomain;

            for (self.colorsHSL, self.domain) |colorHSL, domain| {
                lowHSL = highHSL;
                lowDomain = highDomain;

                highHSL = colorHSL;
                highDomain = domain;

                if (value < domain) {
                    break;
                } else if (value == domain) {
                    return jok.Color.fromHSL(colorHSL);
                }
            } else {
                return jok.Color.fromHSL(highHSL);
            }

            const range = highDomain - lowDomain;
            const ratio = @as(f32, @floatFromInt(value - lowDomain)) / @as(f32, @floatFromInt(range));

            const blendedHSL = [_]f32{
                interp_angle(lowHSL[0], highHSL[0], ratio),
                lowHSL[1] + ratio * (highHSL[1] - lowHSL[1]),
                lowHSL[2] + ratio * (highHSL[2] - lowHSL[2]),
                lowHSL[3] + ratio * (highHSL[3] - lowHSL[3]),
            };

            return jok.Color.fromHSL(blendedHSL);
        }

        pub fn compute(self: Self, comptime num_slots: usize) _ComputedColorGradient(Domain, num_slots, N) {
            return _ComputedColorGradient(Domain, num_slots, self.colorsHSL.len).init(self);
        }
    };
}
//
// fn computeGradient(
//     comptime Domain: type,
//     colorsHSL: []const[4]f32,
//     domains: []const Domain,
//     computed: []jok.Color,
// ) void {
//     //XXX///////////////////////////////////////////////////////////////////////////////////////////
//     //XXX///////////////////////////////////////////////////////////////////////////////////////////
//     @compileLog("colorsHSL len: ", colorsHSL.len);
//     @compileLog("domains len: ", domains.len);
//     @compileLog("computed len: ", computed.len);
//     //XXX///////////////////////////////////////////////////////////////////////////////////////////
//     //XXX///////////////////////////////////////////////////////////////////////////////////////////
//
//     if (colorsHSL.len == 1) {
//         const color = jok.Color.fromHSL(colorsHSL[0]);
//         for (0..computed.len) |i| {
//             computed[i] = color;
//         }
//     } else {
//         const min_domain = domains[0];
//         const max_domain = domains[domains.len - 1];
//
//         var i = 0;
//         var stepIdx = 0;
//         const increment = (max_domain - min_domain) / (
//             switch (@typeInfo(Domain)) {
//                 .float => @as(Domain, @floatFromInt(computed.len)),
//                 .int => computed.len,
//                 else => {
//                     @compileError("Unsupported domain type");
//                 }
//             }
//         );
//         var acc = min_domain;
//
//         //XXX///////////////////////////////////////////////////////////////////////////////////////////
//         @compileLog("min_domain: ", min_domain);
//         @compileLog("max_domain: ", max_domain);
//         @compileLog("increment: ", increment);
//         //XXX///////////////////////////////////////////////////////////////////////////////////////////
//
//         var lowHSL = colorsHSL[0];
//         var lowDomain = domains[0];
//
//         var highHSL = colorsHSL[1];
//         var highDomain = domains[1];
//
//         while (i < max_domain) : ({ i += 1; acc += increment; }) {
//             if (acc >= highDomain) {
//                 computed[i] = jok.Color.fromHSL(highHSL);
//
//                 lowHSL = highHSL;
//                 lowDomain = highDomain;
//
//                 stepIdx += 1;
//                 highHSL = colorsHSL[stepIdx];
//                 highDomain = domains[stepIdx];
//             } else {
//                 const range = highDomain - lowDomain;
//                 const ratio = @as(f32, @floatFromInt(acc - lowDomain)) / @as(f32, @floatFromInt(range));
//
//                 const blendedHSL = [_]f32{
//                     interp_angle(lowHSL[0], highHSL[0], ratio),
//                     lowHSL[1] + ratio * (highHSL[1] - lowHSL[1]),
//                     lowHSL[2] + ratio * (highHSL[2] - lowHSL[2]),
//                     lowHSL[3] + ratio * (highHSL[3] - lowHSL[3]),
//                 };
//
//                 computed[i] = jok.Color.fromHSL(blendedHSL);
//             }
//         }
//     }
// }

fn _ComputedColorGradient(comptime Domain: type, comptime N: usize, comptime num_colors: usize) type {
    return struct {
        const Self = @This();

        colors: [N]jok.Color,
        num_slots: usize = N,

        fn init(base: _ColorGradient(Domain, num_colors)) Self {
            // XXX(zk): this is outrageous :P
            @setEvalBranchQuota(999999);

            var computed: [N]jok.Color = undefined;

            if (base.colorsHSL.len == 1) {
                const color = jok.Color.fromHSL(base.colorsHSL[0]);
                for (0..N) |i| {
                    computed[i] = color;
                }
            } else {
                var i = 0;
                var stepIdx = 0;

                var lowHSL = base.colorsHSL[0];
                var lowDomain = base.domain[0];

                {
                    const color = jok.Color.fromHSL(lowHSL);
                    while (i <= lowDomain) : (i += 1) {
                        computed[i] = color;
                    }
                }

                var highHSL = base.colorsHSL[1];
                var highDomain = base.domain[1];

                const max_domain = base.domain[base.domain.len - 1];

                while (i < max_domain) : (i += 1) {
                    if (i >= highDomain) {
                        computed[i] = jok.Color.fromHSL(highHSL);

                        lowHSL = highHSL;
                        lowDomain = highDomain;

                        stepIdx += 1;
                        highHSL = base.colorsHSL[stepIdx];
                        highDomain = base.domain[stepIdx];
                    } else {
                        const range = highDomain - lowDomain;
                        const ratio = @as(f32, @floatFromInt(i - lowDomain)) / @as(f32, @floatFromInt(range));

                        const blendedHSL = [_]f32{
                            interp_angle(lowHSL[0], highHSL[0], ratio),
                            lowHSL[1] + ratio * (highHSL[1] - lowHSL[1]),
                            lowHSL[2] + ratio * (highHSL[2] - lowHSL[2]),
                            lowHSL[3] + ratio * (highHSL[3] - lowHSL[3]),
                        };

                        computed[i] = jok.Color.fromHSL(blendedHSL);
                    }
                }

                // Fill in the rest of the colors with the last color
                const lastColor = jok.Color.fromHSL(highHSL);
                while (i < N) : (i += 1) {
                    computed[i] = lastColor;
                }
            }

            return Self{ .colors = computed, .num_slots = N };
        }

        pub fn at(self: Self, value: Domain) jok.Color {
            const colorIdx = @mod(value, self.num_slots);
            return self.colors[colorIdx];
        }
    };
}


test "ColorGradient" {
    const gradient = ColorGradient(u8).init(
        &.{ .{ 0, "#000000" }, .{ 255, "#FFFFFF" } },
    ) catch unreachable;

    {
        const color = gradient.at(0);
        const expected = jok.Color.rgb(0, 0, 0);
        try testing.expectEqual(color, expected);
    }
    {
        const color = gradient.at(255);
        const expected = jok.Color.rgb(255, 255, 255);
        try testing.expectEqual(color, expected);
    }
    {
        const color = gradient.at(127);
        const expected = jok.Color.rgb(127, 127, 127);
        try testing.expectEqual(color, expected);
    }
}
