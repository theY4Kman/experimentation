const std = @import("std");
const jok = @import("jok");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const zm = b.dependency("zm", .{});
    const exe = jok.createDesktopApp(
        b,
        "Zig Mandelbrot",
        "src/main.zig",
        target,
        optimize,
        .{
            .additional_deps = &.{
                .{ .name = "zm", .mod = zm.module("zm") },
            },
        },
    );

    exe.root_module.addImport("zm", zm.module("zm"));

    const install_cmd = b.addInstallArtifact(exe, .{});

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(&install_cmd.step);

    const run_step = b.step("run", "Run game");
    run_step.dependOn(&run_cmd.step);
}
