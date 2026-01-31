const std = @import("std");
const builtin = @import("builtin");

pub fn defaultConfigPath(allocator: std.mem.Allocator, menu_id: [:0]const u8) ![]const u8 {
    const home = try homeDir(allocator);
    if (home) |home_dir| {
        const gmenu_config = try std.fs.path.join(allocator, &.{ home_dir, ".config", "gmenu" });
        if (menu_id.len > 0) {
            return std.fs.path.join(allocator, &.{ gmenu_config, menu_id, "config.yaml" });
        }
        return std.fs.path.join(allocator, &.{ gmenu_config, "config.yaml" });
    }

    const config_home = try userConfigDir(allocator, home);
    if (config_home) |dir| {
        const gmenu_config = try std.fs.path.join(allocator, &.{ dir, "gmenu" });
        if (menu_id.len > 0) {
            return std.fs.path.join(allocator, &.{ gmenu_config, menu_id, "config.yaml" });
        }
        return std.fs.path.join(allocator, &.{ gmenu_config, "config.yaml" });
    }

    return error.MissingHome;
}

pub fn findConfigPath(allocator: std.mem.Allocator, menu_id: [:0]const u8) !?[]const u8 {
    const home = try homeDir(allocator);
    const config_home = try userConfigDir(allocator, home);

    if (menu_id.len > 0) {
        if (home) |home_dir| {
            const scoped = try std.fs.path.join(allocator, &.{ home_dir, ".config", "gmenu", menu_id, "config.yaml" });
            if (pathExists(scoped)) return scoped;
            const gmenu_home = try std.fs.path.join(allocator, &.{ home_dir, ".gmenu", menu_id, "config.yaml" });
            if (pathExists(gmenu_home)) return gmenu_home;
        }
        if (config_home) |dir| {
            const scoped = try std.fs.path.join(allocator, &.{ dir, "gmenu", menu_id, "config.yaml" });
            if (pathExists(scoped)) return scoped;
        }
    }

    if (home) |home_dir| {
        const base = try std.fs.path.join(allocator, &.{ home_dir, ".config", "gmenu", "config.yaml" });
        if (pathExists(base)) return base;
        const home_base = try std.fs.path.join(allocator, &.{ home_dir, ".gmenu", "config.yaml" });
        if (pathExists(home_base)) return home_base;
    }
    if (config_home) |dir| {
        const base = try std.fs.path.join(allocator, &.{ dir, "gmenu", "config.yaml" });
        if (pathExists(base)) return base;
    }

    return null;
}

pub fn makePathAbsolute(path: []const u8) !void {
    if (!std.fs.path.isAbsolute(path)) {
        return std.fs.cwd().makePath(path);
    }
    var root = try std.fs.openDirAbsolute("/", .{});
    defer root.close();
    const trimmed = std.mem.trimLeft(u8, path, "/");
    if (trimmed.len == 0) return;
    try root.makePath(trimmed);
}

fn homeDir(allocator: std.mem.Allocator) !?[]const u8 {
    return std.process.getEnvVarOwned(allocator, "HOME") catch |err| switch (err) {
        error.EnvironmentVariableNotFound => null,
        else => return err,
    };
}

fn userConfigDir(allocator: std.mem.Allocator, home: ?[]const u8) !?[]const u8 {
    if (std.process.getEnvVarOwned(allocator, "XDG_CONFIG_HOME")) |dir| {
        return @as(?[]const u8, dir);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (home == null) return null;
    if (builtin.os.tag == .macos) {
        const path = try std.fs.path.join(allocator, &.{ home.?, "Library", "Application Support" });
        return @as(?[]const u8, path);
    }
    const path = try std.fs.path.join(allocator, &.{ home.?, ".config" });
    return @as(?[]const u8, path);
}

fn pathExists(path: []const u8) bool {
    if (std.fs.accessAbsolute(path, .{})) |_| {
        return true;
    } else |err| switch (err) {
        error.FileNotFound => return false,
        else => return false,
    }
}
