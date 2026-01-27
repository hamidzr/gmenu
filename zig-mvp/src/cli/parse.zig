const std = @import("std");
const appconfig = @import("../config.zig");

pub fn applySearchMethod(config: *appconfig.Config, value: []const u8) !void {
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

pub fn parseBool(value: []const u8) !bool {
    if (std.ascii.eqlIgnoreCase(value, "true") or std.ascii.eqlIgnoreCase(value, "1") or std.ascii.eqlIgnoreCase(value, "yes") or std.ascii.eqlIgnoreCase(value, "on")) {
        return true;
    }
    if (std.ascii.eqlIgnoreCase(value, "false") or std.ascii.eqlIgnoreCase(value, "0") or std.ascii.eqlIgnoreCase(value, "no") or std.ascii.eqlIgnoreCase(value, "off")) {
        return false;
    }
    return error.InvalidBool;
}

pub fn parseColorOptional(value: []const u8) !?appconfig.Color {
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
        .r = @as(f64, @floatFromInt(r)) / 255.0,
        .g = @as(f64, @floatFromInt(g)) / 255.0,
        .b = @as(f64, @floatFromInt(b)) / 255.0,
        .a = @as(f64, @floatFromInt(a)) / 255.0,
    };
}

fn parseHexByte(value: []const u8) !u8 {
    return std.fmt.parseInt(u8, value, 16);
}
