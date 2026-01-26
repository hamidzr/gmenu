const std = @import("std");
const appconfig = @import("../config.zig");
const parse = @import("parse.zig");

pub fn resolveMenuIDFromArgs(allocator: std.mem.Allocator, args: []const [:0]const u8) !?[:0]const u8 {
    if (findArgValue(args, "--menu-id", "-m")) |value| {
        return try allocator.dupeZ(u8, value);
    }
    return null;
}

pub fn applyArgs(allocator: std.mem.Allocator, args: []const [:0]const u8, config: *appconfig.Config) !void {
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
            try parse.applySearchMethod(config, args[i]);
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
        if (std.mem.eql(u8, arg, "--ipc-only")) {
            config.ipc_only = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--preserve-order") or std.mem.eql(u8, arg, "-o")) {
            config.search.preserve_order = true;
            continue;
        }
        if (std.mem.startsWith(u8, arg, "--levenshtein-fallback=")) {
            const value = arg["--levenshtein-fallback=".len..];
            config.search.levenshtein_fallback = try parse.parseBool(value);
            continue;
        }
        if (std.mem.eql(u8, arg, "--levenshtein-fallback")) {
            config.search.levenshtein_fallback = true;
            continue;
        }
        if (std.mem.startsWith(u8, arg, "--no-levenshtein-fallback=")) {
            const value = arg["--no-levenshtein-fallback=".len..];
            config.search.levenshtein_fallback = !try parse.parseBool(value);
            continue;
        }
        if (std.mem.eql(u8, arg, "--no-levenshtein-fallback")) {
            config.search.levenshtein_fallback = false;
            continue;
        }
        if (std.mem.eql(u8, arg, "--auto-accept")) {
            config.auto_accept = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--accept-custom-selection")) {
            config.accept_custom_selection = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--no-custom-selection")) {
            config.accept_custom_selection = false;
            continue;
        }
        if (std.mem.eql(u8, arg, "--no-numeric-selection")) {
            config.no_numeric_selection = true;
            continue;
        }
        if (std.mem.startsWith(u8, arg, "--no-numeric-selection=")) {
            const value = arg["--no-numeric-selection=".len..];
            config.no_numeric_selection = try parse.parseBool(value);
            continue;
        }
        if (std.mem.eql(u8, arg, "--show-icons")) {
            config.show_icons = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--limit")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.search.limit = try std.fmt.parseInt(usize, args[i], 10);
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
        if (std.mem.eql(u8, arg, "--field-height")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.field_height = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--padding")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.padding = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--numeric-column-width")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.numeric_column_width = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--icon-column-width")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.icon_column_width = try std.fmt.parseFloat(f64, args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--alternate-rows")) {
            config.alternate_rows = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--background-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.background_color = try parse.parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--list-background-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.list_background_color = try parse.parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--field-background-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.field_background_color = try parse.parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--text-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.text_color = try parse.parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--secondary-text-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.secondary_text_color = try parse.parseColorOptional(args[i]);
            continue;
        }
        if (std.mem.eql(u8, arg, "--selection-color")) {
            i += 1;
            if (i >= args.len) return error.MissingValue;
            config.selection_color = try parse.parseColorOptional(args[i]);
            continue;
        }
    }
}

pub fn hasFlag(args: []const [:0]const u8, flag: []const u8) bool {
    var i: usize = 1;
    while (i < args.len) : (i += 1) {
        if (std.mem.eql(u8, args[i], flag)) return true;
    }
    return false;
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
