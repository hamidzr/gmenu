const std = @import("std");
const builtin = @import("builtin");
const appconfig = @import("config.zig");

const ConfigKeyVariant = struct {
    canonical: []const u8,
    camel: []const u8,
};

const config_key_variants = [_]ConfigKeyVariant{
    .{ .canonical = "title", .camel = "" },
    .{ .canonical = "prompt", .camel = "" },
    .{ .canonical = "menu_id", .camel = "menuId" },
    .{ .canonical = "search_method", .camel = "searchMethod" },
    .{ .canonical = "preserve_order", .camel = "preserveOrder" },
    .{ .canonical = "initial_query", .camel = "initialQuery" },
    .{ .canonical = "auto_accept", .camel = "autoAccept" },
    .{ .canonical = "terminal_mode", .camel = "terminalMode" },
    .{ .canonical = "follow_stdin", .camel = "followStdin" },
    .{ .canonical = "no_numeric_selection", .camel = "noNumericSelection" },
    .{ .canonical = "show_icons", .camel = "showIcons" },
    .{ .canonical = "show_score", .camel = "showScore" },
    .{ .canonical = "limit", .camel = "" },
    .{ .canonical = "min_width", .camel = "minWidth" },
    .{ .canonical = "min_height", .camel = "minHeight" },
    .{ .canonical = "max_width", .camel = "maxWidth" },
    .{ .canonical = "max_height", .camel = "maxHeight" },
    .{ .canonical = "row_height", .camel = "rowHeight" },
    .{ .canonical = "alternate_rows", .camel = "alternateRows" },
    .{ .canonical = "accept_custom_selection", .camel = "acceptCustomSelection" },
    .{ .canonical = "background_color", .camel = "backgroundColor" },
    .{ .canonical = "list_background_color", .camel = "listBackgroundColor" },
    .{ .canonical = "field_background_color", .camel = "fieldBackgroundColor" },
};

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
    var seen_keys: [config_key_variants.len]?[]const u8 = [_]?[]const u8{null} ** config_key_variants.len;
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

        const canonical_index = canonicalKeyIndex(key) orelse return error.InvalidConfigKey;
        if (seen_keys[canonical_index]) |previous| {
            if (!std.mem.eql(u8, previous, key)) return error.ConfigKeyStyleConflict;
        } else {
            seen_keys[canonical_index] = key;
        }

        try applyConfigKV(allocator, config, key, value);
    }
}

fn canonicalKeyIndex(key: []const u8) ?usize {
    for (config_key_variants, 0..) |variant, idx| {
        if (std.mem.eql(u8, key, variant.canonical)) return idx;
        if (variant.camel.len > 0 and std.mem.eql(u8, key, variant.camel)) return idx;
    }
    return null;
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

    if (envValue(allocator, "GMENU_TERMINAL_MODE")) |value| {
        config.terminal_mode = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_FOLLOW_STDIN")) |value| {
        config.follow_stdin = try parseBool(value);
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
    if (envValue(allocator, "GMENU_ROW_HEIGHT")) |value| {
        config.row_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_ALTERNATE_ROWS")) |value| {
        config.alternate_rows = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_BACKGROUND_COLOR")) |value| {
        config.background_color = try parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_LIST_BACKGROUND_COLOR")) |value| {
        config.list_background_color = try parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_FIELD_BACKGROUND_COLOR")) |value| {
        config.field_background_color = try parseColorOptional(value);
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
    if (envValue(allocator, "GMENU_SHOW_ICONS")) |value| {
        config.show_icons = try parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_SHOW_SCORE")) |value| {
        config.show_score = try parseBool(value);
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
        if (std.mem.eql(u8, arg, "--terminal")) {
            config.terminal_mode = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--follow-stdin")) {
            config.follow_stdin = true;
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
        if (std.mem.eql(u8, arg, "--row-height")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.row_height = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--background-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.background_color = try parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--list-background-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.list_background_color = try parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--field-background-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.field_background_color = try parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--alternate-rows")) {
            config.alternate_rows = true;
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
        if (std.mem.startsWith(u8, arg, "--no-numeric-selection=")) {
            const value = arg["--no-numeric-selection=".len..];
            config.no_numeric_selection = try parseBool(value);
            continue;
        }
        if (std.mem.eql(u8, arg, "--no-numeric-selection")) {
            config.no_numeric_selection = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--show-icons")) {
            config.show_icons = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--show-score")) {
            config.show_score = true;
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
    if (eqKey(key, "terminal_mode") or eqKey(key, "terminalMode")) {
        config.terminal_mode = try parseBool(value);
        return;
    }
    if (eqKey(key, "follow_stdin") or eqKey(key, "followStdin")) {
        config.follow_stdin = try parseBool(value);
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
    if (eqKey(key, "show_icons") or eqKey(key, "showIcons")) {
        config.show_icons = try parseBool(value);
        return;
    }
    if (eqKey(key, "show_score") or eqKey(key, "showScore")) {
        config.show_score = try parseBool(value);
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
    if (eqKey(key, "row_height") or eqKey(key, "rowHeight")) {
        config.row_height = try std.fmt.parseFloat(f64, value);
        return;
    }
    if (eqKey(key, "alternate_rows") or eqKey(key, "alternateRows")) {
        config.alternate_rows = try parseBool(value);
        return;
    }
    if (eqKey(key, "background_color") or eqKey(key, "backgroundColor")) {
        config.background_color = try parseColorOptional(value);
        return;
    }
    if (eqKey(key, "list_background_color") or eqKey(key, "listBackgroundColor")) {
        config.list_background_color = try parseColorOptional(value);
        return;
    }
    if (eqKey(key, "field_background_color") or eqKey(key, "fieldBackgroundColor")) {
        config.field_background_color = try parseColorOptional(value);
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
    if (std.ascii.eqlIgnoreCase(value, "fuzzy1")) {
        config.search.method = .fuzzy1;
        return;
    }
    if (std.ascii.eqlIgnoreCase(value, "fuzzy3")) {
        config.search.method = .fuzzy3;
        return;
    }
    if (std.ascii.eqlIgnoreCase(value, "default")) {
        config.search.method = .default;
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

fn parseColorOptional(value: []const u8) !?appconfig.Color {
    const trimmed = std.mem.trim(u8, value, " \t");
    if (trimmed.len == 0) return null;
    if (std.ascii.eqlIgnoreCase(trimmed, "none") or std.ascii.eqlIgnoreCase(trimmed, "default")) {
        return null;
    }
    return try parseHexColor(trimmed);
}

fn parseHexColor(value: []const u8) !appconfig.Color {
    var hex = value;
    if (hex.len > 0 and hex[0] == '#') {
        hex = hex[1..];
    }
    if (hex.len != 6 and hex.len != 8) return error.InvalidColor;

    const r = try parseHexByte(hex[0..2]);
    const g = try parseHexByte(hex[2..4]);
    const b = try parseHexByte(hex[4..6]);
    const a: u8 = if (hex.len == 8) try parseHexByte(hex[6..8]) else 255;

    return .{
        .r = @as(f64, r) / 255.0,
        .g = @as(f64, g) / 255.0,
        .b = @as(f64, b) / 255.0,
        .a = @as(f64, a) / 255.0,
    };
}

fn parseHexByte(value: []const u8) !u8 {
    return std.fmt.parseInt(u8, value, 16);
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

    try file.deprecatedWriter().print(
        \\# gmenu config
        \\title: {s}
        \\prompt: {s}
        \\search_method: fuzzy
        \\preserve_order: false
        \\initial_query: ""
        \\terminal_mode: false
        \\follow_stdin: false
        \\auto_accept: false
        \\accept_custom_selection: true
        \\no_numeric_selection: true
        \\show_icons: false
        \\show_score: false
        \\min_width: {d}
        \\min_height: {d}
        \\max_width: {d}
        \\max_height: {d}
        \\row_height: {d}
        \\alternate_rows: true
        \\background_color: ""
        \\list_background_color: ""
        \\field_background_color: ""
        \\
    ,
        .{
            defaults.title,
            defaults.placeholder,
            @as(i64, @intFromFloat(defaults.window_width)),
            @as(i64, @intFromFloat(defaults.window_height)),
            @as(i64, @intFromFloat(defaults.max_width)),
            @as(i64, @intFromFloat(defaults.max_height)),
            @as(i64, @intFromFloat(defaults.row_height)),
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
