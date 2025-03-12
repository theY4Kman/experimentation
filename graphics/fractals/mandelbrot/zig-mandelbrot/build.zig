const std = @import("std");
const jok = @import("jok");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const zm = b.dependency("zm", .{});
    const spice = b.dependency("spice", .{});

    const zm_mod = zm.module("zm");
    const spice_mod = spice.module("spice");

    const exe = jok.createDesktopApp(
        b,
        "zig-mandelbrot",
        "src/main.zig",
        target,
        optimize,
        .{
            .additional_deps = &.{
                .{ .name = "zm", .mod = zm_mod },
                .{ .name = "spice", .mod = spice_mod },
            },
        },
    );

    exe.root_module.addImport("zm", zm_mod);
    exe.root_module.addImport("spice", spice_mod);

    const install_cmd = b.addInstallArtifact(exe, .{});

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(&install_cmd.step);

    const run_step = b.step("run", "Run game");
    run_step.dependOn(&run_cmd.step);
}
