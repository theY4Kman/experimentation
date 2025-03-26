const std = @import("std");
const jok = @import("jok");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const zm = b.dependency("zm", .{});
    const spice = b.dependency("spice", .{});
    const clap = b.dependency("clap", .{});

    const zm_mod = zm.module("zm");
    const spice_mod = spice.module("spice");
    const clap_mod = clap.module("clap");

    const exe = jok.createDesktopApp(
        b,
        "zig-mandelbrot",
        "src/main.zig",
        target,
        optimize,
        .{
            .no_audio = true,
            .additional_deps = &.{
                .{ .name = "zm", .mod = zm_mod },
                .{ .name = "spice", .mod = spice_mod },
                .{ .name = "clap", .mod = clap_mod },
            },
        },
    );

    exe.root_module.addImport("zm", zm_mod);
    exe.root_module.addImport("spice", spice_mod);
    exe.root_module.addImport("clap", clap_mod);

    const install_cmd = b.addInstallArtifact(exe, .{});

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(&install_cmd.step);

    const run_step = b.step("run", "Run game");
    run_step.dependOn(&run_cmd.step);

    const tests = jok.createTest(
        b,
        "test",
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
    b.installArtifact(tests);

    const run_tests = b.addRunArtifact(tests);

    const test_step = b.step("test", "Run tests");
    test_step.dependOn(&run_tests.step);
}
