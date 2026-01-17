const std = @import("std");

pub fn create(allocator: std.mem.Allocator, name: []const u8) ![]const u8 {
    const pid_name = if (name.len == 0) "zmenu" else name;
    const temp_dir = try tempDir(allocator);
    const filename = try std.mem.concat(allocator, u8, &.{ pid_name, ".pid" });
    const pid_path = try std.fs.path.join(allocator, &.{ temp_dir, filename });

    if (try existingPidIsRunning(pid_path)) {
        return error.AlreadyRunning;
    }

    var file = std.fs.createFileAbsolute(pid_path, .{ .exclusive = true, .truncate = true }) catch |err| switch (err) {
        error.PathAlreadyExists => return error.AlreadyRunning,
        else => return err,
    };
    defer file.close();

    const pid: std.posix.pid_t = @intCast(std.c.getpid());
    var buf: [32]u8 = undefined;
    const pid_line = std.fmt.bufPrint(&buf, "{d}\n", .{pid}) catch "";
    file.writeAll(pid_line) catch {};

    return pid_path;
}

pub fn remove(path: []const u8) void {
    std.fs.deleteFileAbsolute(path) catch {};
}

fn existingPidIsRunning(pid_path: []const u8) !bool {
    var file = std.fs.openFileAbsolute(pid_path, .{ .mode = .read_only }) catch |err| switch (err) {
        error.FileNotFound => return false,
        else => return err,
    };
    defer file.close();

    var buf: [64]u8 = undefined;
    const len = file.readAll(&buf) catch return false;
    const trimmed = std.mem.trim(u8, buf[0..len], " \t\r\n");
    if (trimmed.len == 0) {
        std.fs.deleteFileAbsolute(pid_path) catch {};
        return false;
    }

    const parsed_pid = std.fmt.parseInt(std.posix.pid_t, trimmed, 10) catch {
        std.fs.deleteFileAbsolute(pid_path) catch {};
        return false;
    };
    if (parsed_pid <= 0) {
        std.fs.deleteFileAbsolute(pid_path) catch {};
        return false;
    }
    if (pidIsAlive(parsed_pid)) {
        return true;
    }

    std.fs.deleteFileAbsolute(pid_path) catch {};
    return false;
}

fn pidIsAlive(pid: std.posix.pid_t) bool {
    if (pid <= 0) return false;
    std.posix.kill(pid, 0) catch |err| switch (err) {
        error.ProcessNotFound => return false,
        error.PermissionDenied => return true,
        else => return true,
    };
    return true;
}

fn tempDir(_: std.mem.Allocator) ![]const u8 {
    const env = std.posix.getenv("TMPDIR") orelse std.posix.getenv("TMP") orelse std.posix.getenv("TEMP");
    if (env) |value| return std.mem.span(value);
    return "/tmp";
}
