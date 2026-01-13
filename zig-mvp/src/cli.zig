const std = @import("std");
const builtin = @import("builtin");
const appconfig = @import("config.zig");

pub fn parse(allocator: std.mem.Allocator) !appconfig.Config {
    var config = appconfig.defaults();

    const args = try std.process.argsAlloc(allocator);
    defer std.process.argsFree(allocator, args);

    if (try resolveMenuID(allocator, args)) |menu_id| {
        config.menu_id = menu_id;
    }

    if (hasFlag(args, "--init-config")) {
        const path = try writeDefaultConfig(allocator, config.menu_id);
        std.fs.File.stdout().deprecatedWriter().print("zmenu: wrote config to {s}\n", .{path}) catch {};
        std.process.exit(0);
    }

    try loadConfigFile(allocator, config.menu_id, &config);
    try applyEnv(allocator, &config);
    try applyArgs(allocator, args, &config);

    return config;
}

fn resolveMenuID(allocator: std.mem.Allocator, args: []const [:0]const u8) !?[:0]const u8 {
    var menu_id: ?[:0]const u8 = null;
    if (envValue(allocator, "GMENU_MENU_ID")) |value| {
        menu_id = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    if (findArgValue(args, "--menu-id", "-m")) |value| {
        menu_id = try allocator.dupeZ(u8, value);
    }

    return menu_id;
}

fn loadConfigFile(allocator: std.mem.Allocator, menu_id: [:0]const u8, config: *appconfig.Config) !void {
    const path = try findConfigPath(allocator, menu_id);
    if (path == null) return;

    var file = std.fs.openFileAbsolute(path.?, .{}) catch |err| switch (err) {
        error.FileNotFound => return,
        else => return err,
    };
    defer file.close();

    const contents = try file.readToEndAlloc(allocator, 64 * 1024);
    var iter = std.mem.splitScalar(u8, contents, '\n');
    while (iter.next()) |line| {
        var trimmed = std.mem.trim(u8, line, " \t\r");
        if (trimmed.len == 0 or trimmed[0] == '#') continue;
        if (std.mem.indexOfScalar(u8, trimmed, '#')) |idx| {
            trimmed = std.mem.trim(u8, trimmed[0..idx], " \t");
        }
        if (trimmed.len == 0) continue;

        const colon = std.mem.indexOfScalar(u8, trimmed, ':') orelse continue;
        const key = std.mem.trim(u8, trimmed[0..colon], " \t");
        var value = std.mem.trim(u8, trimmed[colon + 1 ..], " \t");
        value = stripQuotes(value);
        if (key.len == 0) continue;

        try applyConfigKV(allocator, config, key, value);
    }
}

fn applyEnv(allocator: std.mem.Allocator, config: *appconfig.Config) !void {
    if (envValue(allocator, "GMENU_TITLE")) |value| {
        config.title = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_PROMPT")) |value| {
        config.placeholder = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MENU_ID")) |value| {
        config.menu_id = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_INITIAL_QUERY")) |value| {
        config.initial_query = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    if (envValue(allocator, "GMENU_MIN_WIDTH")) |value| {
        config.window_width = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MIN_HEIGHT")) |value| {
        config.window_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MAX_WIDTH")) |value| {
        config.max_width = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MAX_HEIGHT")) |value| {
        config.max_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    if (envValue(allocator, "GMENU_SEARCH_METHOD")) |value| {
        try applySearchMethod(config, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_PRESERVE_ORDER")) |value| {
        config.search.preserve_order = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_AUTO_ACCEPT")) |value| {
        config.auto_accept = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_ACCEPT_CUSTOM_SELECTION")) |value| {
        config.accept_custom_selection = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_NO_NUMERIC_SELECTION")) |value| {
        config.no_numeric_selection = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
}

fn applyArgs(allocator: std.mem.Allocator, args: []const [:0]const u8, config: *appconfig.Config) !void {
    var i: usize = 1;
    while (i < args.len) : (i += 1) {
        const arg = args[i];

        if (std.mem.eql(u8, arg, "--menu-id") or std.mem.eql(u8, arg, "-m")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.menu_id = try allocator.dupeZ(u8, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--initial-query") or std.mem.eql(u8, arg, "-q")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.initial_query = try allocator.dupeZ(u8, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--title") or std.mem.eql(u8, arg, "-t")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.title = try allocator.dupeZ(u8, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--prompt") or std.mem.eql(u8, arg, "-p")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.placeholder = try allocator.dupeZ(u8, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--search-method") or std.mem.eql(u8, arg, "-s")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            try applySearchMethod(config, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--min-width")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.window_width = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--min-height")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.window_height = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--max-width")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.max_width = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--max-height")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.max_height = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--preserve-order") or std.mem.eql(u8, arg, "-o")) {
            config.search.preserve_order = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--auto-accept")) {
            config.auto_accept = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--no-numeric-selection")) {
            config.no_numeric_selection = true;
            continue;
        }
    }
}

fn applyConfigKV(allocator: std.mem.Allocator, config: *appconfig.Config, key: []const u8, value: []const u8) !void {
    if (eqKey(key, "title")) {
        config.title = try allocator.dupeZ(u8, value);
        return;
    }
    if (eqKey(key, "prompt")) {
        config.placeholder = try allocator.dupeZ(u8, value);
        return;
    }
    if (eqKey(key, "menu_id") or eqKey(key, "menuId")) {
        config.menu_id = try allocator.dupeZ(u8, value);
        return;
    }
    if (eqKey(key, "initial_query") or eqKey(key, "initialQuery")) {
        config.initial_query = try allocator.dupeZ(u8, value);
        return;
    }
    if (eqKey(key, "search_method") or eqKey(key, "searchMethod")) {
        try applySearchMethod(config, value);
        return;
    }
    if (eqKey(key, "preserve_order") or eqKey(key, "preserveOrder")) {
        config.search.preserve_order = try parseBool(value);
        return;
    }
    if (eqKey(key, "auto_accept") or eqKey(key, "autoAccept")) {
        config.auto_accept = try parseBool(value);
        return;
    }
    if (eqKey(key, "accept_custom_selection") or eqKey(key, "acceptCustomSelection")) {
        config.accept_custom_selection = try parseBool(value);
        return;
    }
    if (eqKey(key, "no_numeric_selection") or eqKey(key, "noNumericSelection")) {
        config.no_numeric_selection = try parseBool(value);
        return;
    }
    if (eqKey(key, "limit")) {
        config.search.limit = try std.fmt.parseInt(usize, value, 10);
        return;
    }
    if (eqKey(key, "min_width") or eqKey(key, "minWidth")) {
        config.window_width = try std.fmt.parseFloat(f64, value);
        return;
    }
    if (eqKey(key, "min_height") or eqKey(key, "minHeight")) {
        config.window_height = try std.fmt.parseFloat(f64, value);
        return;
    }
    if (eqKey(key, "max_width") or eqKey(key, "maxWidth")) {
        config.max_width = try std.fmt.parseFloat(f64, value);
        return;
    }
    if (eqKey(key, "max_height") or eqKey(key, "maxHeight")) {
        config.max_height = try std.fmt.parseFloat(f64, value);
        return;
    }
}

fn applySearchMethod(config: *appconfig.Config, value: []const u8) !void {
    if (std.ascii.eqlIgnoreCase(value, "direct")) {
        config.search.method = .direct;
        return;
    }
    if (std.ascii.eqlIgnoreCase(value, "fuzzy")) {
        config.search.method = .fuzzy;
        return;
    }
    return error.InvalidSearchMethod;
}

fn parseBool(value: []const u8) !bool {
    if (std.ascii.eqlIgnoreCase(value, "true") or std.ascii.eqlIgnoreCase(value, "1") or std.ascii.eqlIgnoreCase(value, "yes") or std.ascii.eqlIgnoreCase(value, "on")) {
        return true;
    }
    if (std.ascii.eqlIgnoreCase(value, "false") or std.ascii.eqlIgnoreCase(value, "0") or std.ascii.eqlIgnoreCase(value, "no") or std.ascii.eqlIgnoreCase(value, "off")) {
        return false;
    }
    return error.InvalidBool;
}

fn stripQuotes(value: []const u8) []const u8 {
    if (value.len >= 2 and ((value[0] == '"' and value[value.len - 1] == '"') or (value[0] == '\'' and value[value.len - 1] == '\''))) {
        return value[1 .. value.len - 1];
    }
    return value;
}

fn eqKey(a: []const u8, b: []const u8) bool {
    return std.mem.eql(u8, a, b);
}

fn envValue(allocator: std.mem.Allocator, name: []const u8) ![:0]const u8 {
    const value = try std.process.getEnvVarOwned(allocator, name);
    return allocator.dupeZ(u8, value);
}

fn findArgValue(args: []const [:0]const u8, long_flag: []const u8, short_flag: []const u8) ?[]const u8 {
    var i: usize = 1;
    while (i < args.len) : (i += 1) {
        const arg = args[i];
        if (std.mem.eql(u8, arg, long_flag) or std.mem.eql(u8, arg, short_flag)) {
            if (i + 1 < args.len) return args[i + 1];
            return null;
        }
        if (std.mem.startsWith(u8, arg, long_flag)) {
            if (arg.len > long_flag.len and arg[long_flag.len] == '=') {
                return arg[long_flag.len + 1 ..];
            }
        }
    }
    return null;
}

fn writeDefaultConfig(allocator: std.mem.Allocator, menu_id: [:0]const u8) ![]const u8 {
    const path = try defaultConfigPath(allocator, menu_id);
    const dir = std.fs.path.dirname(path) orelse return error.InvalidPath;
    try makePathAbsolute(dir);

    const defaults = appconfig.defaults();
    var file = try std.fs.createFileAbsolute(path, .{ .truncate = true });
    defer file.close();

    try file.writer().print(
        \\# gmenu config
        \\title: {s}
        \\prompt: {s}
        \\search_method: fuzzy
        \\preserve_order: false
        \\initial_query: ""
        \\auto_accept: false
        \\accept_custom_selection: true
        \\no_numeric_selection: false
        \\min_width: {d}
        \\min_height: {d}
        \\max_width: {d}
        \\max_height: {d}
        \\
    ,
        .{
            defaults.title,
            defaults.placeholder,
            @as(i64, @intFromFloat(defaults.window_width)),
            @as(i64, @intFromFloat(defaults.window_height)),
            @as(i64, @intFromFloat(defaults.max_width)),
            @as(i64, @intFromFloat(defaults.max_height)),
        },
    );

    return path;
}

fn defaultConfigPath(allocator: std.mem.Allocator, menu_id: [:0]const u8) ![]const u8 {
    const home = std.process.getEnvVarOwned(allocator, "HOME") catch |err| switch (err) {
        error.EnvironmentVariableNotFound => return error.MissingHome,
        else => return err,
    };

    const config_home = if (std.process.getEnvVarOwned(allocator, "XDG_CONFIG_HOME")) |dir| dir else |err| blk: {
        if (err != error.EnvironmentVariableNotFound) return err;
        if (builtin.os.tag == .macos) break :blk try std.fs.path.join(allocator, &.{ home, "Library", "Application Support" });
        break :blk try std.fs.path.join(allocator, &.{ home, ".config" });
    };

    const gmenu_config = try std.fs.path.join(allocator, &.{ config_home, "gmenu" });
    if (menu_id.len > 0) {
        return std.fs.path.join(allocator, &.{ gmenu_config, menu_id, "config.yaml" });
    }
    return std.fs.path.join(allocator, &.{ gmenu_config, "config.yaml" });
}

fn makePathAbsolute(path: []const u8) !void {
    if (!std.fs.path.isAbsolute(path)) {
        return std.fs.cwd().makePath(path);
    }
    var root = try std.fs.openDirAbsolute("/", .{});
    defer root.close();
    const trimmed = std.mem.trimLeft(u8, path, "/");
    if (trimmed.len == 0) return;
    try root.makePath(trimmed);
}

fn hasFlag(args: []const [:0]const u8, flag: []const u8) bool {
    var i: usize = 1;
    while (i < args.len) : (i += 1) {
        if (std.mem.eql(u8, args[i], flag)) return true;
    }
    return false;
}

fn findConfigPath(allocator: std.mem.Allocator, menu_id: [:0]const u8) !?[]const u8 {
    const home = std.process.getEnvVarOwned(allocator, "HOME") catch |err| switch (err) {
        error.EnvironmentVariableNotFound => return null,
        else => return err,
    };

    const config_home = if (std.process.getEnvVarOwned(allocator, "XDG_CONFIG_HOME")) |dir| dir else |err| blk: {
        if (err != error.EnvironmentVariableNotFound) return err;
        if (builtin.os.tag == .macos) break :blk try std.fs.path.join(allocator, &.{ home, "Library", "Application Support" });
        break :blk try std.fs.path.join(allocator, &.{ home, ".config" });
    };

    const gmenu_config = try std.fs.path.join(allocator, &.{ config_home, "gmenu" });
    if (menu_id.len > 0) {
        const scoped = try std.fs.path.join(allocator, &.{ gmenu_config, menu_id, "config.yaml" });
        if (pathExists(scoped)) return scoped;
    }
    const base = try std.fs.path.join(allocator, &.{ gmenu_config, "config.yaml" });
    if (pathExists(base)) return base;

    const gmenu_home = try std.fs.path.join(allocator, &.{ home, ".gmenu" });
    if (menu_id.len > 0) {
        const scoped = try std.fs.path.join(allocator, &.{ gmenu_home, menu_id, "config.yaml" });
        if (pathExists(scoped)) return scoped;
    }
    const home_base = try std.fs.path.join(allocator, &.{ gmenu_home, "config.yaml" });
    if (pathExists(home_base)) return home_base;

    return null;
}

fn pathExists(path: []const u8) bool {
    if (std.fs.accessAbsolute(path, .{})) |_| {
        return true;
    } else |err| switch (err) {
        error.FileNotFound => return false,
        else => return false,
    }
}
