const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const objc_dep = b.dependency("zig_objc", .{
        .target = target,
        .optimize = optimize,
    });

    const exe = b.addExecutable(.{
        .name = "zmenu",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/main.zig"),
            .target = target,
            .optimize = optimize,
            .imports = &.{
                .{ .name = "objc", .module = objc_dep.module("objc") },
            },
        }),
    });

    exe.linkFramework("AppKit");
    exe.linkFramework("Foundation");

    b.install_prefix = "bin";
    b.installArtifact(exe);

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(b.getInstallStep());
    if (b.args) |args| {
        run_cmd.addArgs(args);
    }

    const run_step = b.step("run", "Run the app");
    run_step.dependOn(&run_cmd.step);

    const search_module = b.createModule(.{
        .root_source_file = b.path("src/search.zig"),
        .target = target,
        .optimize = optimize,
    });
    const search_tests = b.addTest(.{
        .name = "search",
        .root_module = search_module,
    });
    const run_search_tests = b.addRunArtifact(search_tests);

    const test_step = b.step("test", "Run tests");
    test_step.dependOn(&run_search_tests.step);
}
