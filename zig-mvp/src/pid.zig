const std = @import("std");

pub fn create(allocator: std.mem.Allocator, name: []const u8) ![]const u8 {
    const pid_name = if (name.len == 0) "zmenu" else name;
    const temp_dir = try tempDir(allocator);
    const filename = try std.mem.concat(allocator, u8, &.{ pid_name, ".pid" });
    const pid_path = try std.fs.path.join(allocator, &.{ temp_dir, filename });

    if (std.fs.accessAbsolute(pid_path, .{})) |_| {
        return error.AlreadyRunning;
    } else |err| switch (err) {
        error.FileNotFound => {},
        else => return err,
    }

    var file = try std.fs.createFileAbsolute(pid_path, .{});
    defer file.close();

    return pid_path;
}

pub fn remove(path: []const u8) void {
    std.fs.deleteFileAbsolute(path) catch {};
}

fn tempDir(allocator: std.mem.Allocator) ![]const u8 {
    if (std.process.getEnvVarOwned(allocator, "TMPDIR")) |dir| return dir else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (std.process.getEnvVarOwned(allocator, "TMP")) |dir| return dir else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (std.process.getEnvVarOwned(allocator, "TEMP")) |dir| return dir else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    return allocator.dupe(u8, "/tmp");
}
