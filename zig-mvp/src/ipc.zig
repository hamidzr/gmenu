const std = @import("std");

pub const Item = struct {
    id: []const u8 = "",
    label: []const u8,
    icon: ?[]const u8 = null,
};

pub const Message = struct {
    v: u32 = 1,
    cmd: []const u8,
    items: ?[]Item = null,
};

pub fn socketPath(allocator: std.mem.Allocator, menu_id: []const u8) ![]const u8 {
    const dir = try tempDir(allocator);
    const name = if (menu_id.len > 0)
        try std.fmt.allocPrint(allocator, "zmenu.{s}.sock", .{menu_id})
    else
        try allocator.dupe(u8, "zmenu.sock");

    return std.fs.path.join(allocator, &.{ dir, name });
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
