const std = @import("std");
const builtin = @import("builtin");

pub const State = struct {
    last_query: []const u8,
    last_selection: []const u8,
};

pub fn load(allocator: std.mem.Allocator, menu_id: []const u8) !?State {
    const path = try cacheFilePath(allocator, menu_id);
    var file = std.fs.openFileAbsolute(path, .{}) catch |err| switch (err) {
        error.FileNotFound => return null,
        else => return err,
    };
    defer file.close();

    const contents = try file.readToEndAlloc(allocator, 64 * 1024);
    var last_query: []const u8 = "";
    var last_selection: []const u8 = "";

    var iter = std.mem.splitScalar(u8, contents, '\n');
    while (iter.next()) |line| {
        const trimmed = std.mem.trim(u8, line, " \t\r");
        if (trimmed.len == 0) continue;
        if (std.mem.startsWith(u8, trimmed, "last_query:")) {
            last_query = std.mem.trim(u8, trimmed["last_query:".len..], " \t");
            continue;
        }
        if (std.mem.startsWith(u8, trimmed, "last_selection:")) {
            last_selection = std.mem.trim(u8, trimmed["last_selection:".len..], " \t");
            continue;
        }
    }

    if (last_query.len == 0 and last_selection.len == 0) return null;

    return .{
        .last_query = last_query,
        .last_selection = last_selection,
    };
}

pub fn save(allocator: std.mem.Allocator, menu_id: []const u8, state: State) !void {
    const dir = try cacheDir(allocator, menu_id);
    try std.fs.cwd().makePath(dir);
    const path = try cacheFilePath(allocator, menu_id);

    var file = try std.fs.createFileAbsolute(path, .{ .truncate = true });
    defer file.close();

    try file.deprecatedWriter().print(
        "last_query: {s}\nlast_selection: {s}\n",
        .{ state.last_query, state.last_selection },
    );
}

fn cacheFilePath(allocator: std.mem.Allocator, menu_id: []const u8) ![]const u8 {
    const dir = try cacheDir(allocator, menu_id);
    return std.fs.path.join(allocator, &.{ dir, "cache.yaml" });
}

fn cacheDir(allocator: std.mem.Allocator, menu_id: []const u8) ![]const u8 {
    const root = try cacheRoot(allocator);
    const base = try std.fs.path.join(allocator, &.{ root, "gmenu" });
    if (menu_id.len == 0) return base;
    return std.fs.path.join(allocator, &.{ base, menu_id });
}

fn cacheRoot(allocator: std.mem.Allocator) ![]const u8 {
    if (std.process.getEnvVarOwned(allocator, "XDG_CACHE_HOME")) |dir| return dir else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    const home = try std.process.getEnvVarOwned(allocator, "HOME");
    if (builtin.os.tag == .macos) {
        return std.fs.path.join(allocator, &.{ home, "Library", "Caches" });
    }
    return std.fs.path.join(allocator, &.{ home, ".cache" });
}
