const std = @import("std");
const appconfig = @import("config.zig");
const pid = @import("pid.zig");
const menu = @import("menu.zig");
const exit_codes = @import("exit_codes.zig");

pub fn run(config: appconfig.Config, allocator: std.mem.Allocator) !void {
    const items = readItems(allocator) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: stdin is empty\n", .{}) catch {};
        std.process.exit(exit_codes.unknown_error);
    };

    const pid_path = pid.create(allocator, config.menu_id) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: another instance is running\n", .{}) catch {};
        std.process.exit(exit_codes.unknown_error);
    };
    defer pid.remove(pid_path);

    var tty = std.fs.openFileAbsolute("/dev/tty", .{}) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: unable to open tty\n", .{}) catch {};
        std.process.exit(exit_codes.unknown_error);
    };
    defer tty.close();

    const original_term = enableRawMode(tty.handle) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: unable to enter raw mode\n", .{}) catch {};
        std.process.exit(exit_codes.unknown_error);
    };
    defer std.posix.tcsetattr(tty.handle, .NOW, original_term) catch {};

    var input = std.ArrayList(u8).init(allocator);
    defer input.deinit();
    if (config.initial_query.len > 0) {
        try input.appendSlice(config.initial_query);
    }

    try renderMatches(&tty, config.placeholder, input.items, items, config.no_numeric_selection);

    const reader = tty.reader();
    var final_query: []const u8 = input.items;
    var exit_code: ?u8 = null;

    while (true) {
        const ch = reader.readByte() catch {
            exit_code = exit_codes.unknown_error;
            break;
        };

        switch (ch) {
            '\r', '\n' => {
                final_query = input.items;
                break;
            },
            127, 8 => {
                if (input.items.len > 0) {
                    input.items.len -= 1;
                    try renderMatches(&tty, config.placeholder, input.items, items, config.no_numeric_selection);
                }
            },
            3 => {
                tty.deprecatedWriter().print("\n{s}Input cancelled\n", .{config.placeholder}) catch {};
                exit_code = exit_codes.user_canceled;
                break;
            },
            else => {
                if (ch >= 32 and ch <= 126) {
                    try input.append(ch);
                    try renderMatches(&tty, config.placeholder, input.items, items, config.no_numeric_selection);
                }
            },
        }
    }

    if (exit_code != null) {
        std.posix.tcsetattr(tty.handle, .NOW, original_term) catch {};
        std.process.exit(exit_code.?);
    }

    var match_count: usize = 0;
    var first_match: ?[]const u8 = null;
    for (items) |item| {
        if (containsInsensitive(item.label, final_query)) {
            match_count += 1;
            if (first_match == null) {
                first_match = item.label;
            }
        }
    }

    if (match_count == 0) {
        std.fs.File.stderr().deprecatedWriter().print("No matches found\n", .{}) catch {};
        return;
    }
    if (match_count > 1) {
        std.fs.File.stderr().deprecatedWriter().print("Multiple matches found. Picking the first one.\n", .{}) catch {};
    }

    std.fs.File.stdout().deprecatedWriter().print("\x1b[2J\x1b[H", .{}) catch {};
    std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{first_match.?}) catch {};
}

fn enableRawMode(fd: std.posix.fd_t) !std.posix.termios {
    var term = try std.posix.tcgetattr(fd);
    var raw = term;

    raw.iflag.BRKINT = false;
    raw.iflag.ICRNL = false;
    raw.iflag.INPCK = false;
    raw.iflag.ISTRIP = false;
    raw.iflag.IXON = false;
    raw.iflag.INLCR = false;
    raw.iflag.IGNCR = false;
    raw.iflag.PARMRK = false;

    raw.oflag.OPOST = false;

    raw.lflag.ECHO = false;
    raw.lflag.ECHONL = false;
    raw.lflag.ICANON = false;
    raw.lflag.ISIG = false;
    raw.lflag.IEXTEN = false;

    raw.cflag.PARENB = false;
    raw.cflag.CSIZE = .CS8;

    raw.cc[@intFromEnum(std.posix.V.MIN)] = 1;
    raw.cc[@intFromEnum(std.posix.V.TIME)] = 0;

    try std.posix.tcsetattr(fd, .NOW, raw);
    return term;
}

fn renderMatches(
    tty: *std.fs.File,
    prompt: [:0]const u8,
    query: []const u8,
    items: []menu.MenuItem,
    no_numeric_selection: bool,
) !void {
    var writer = tty.deprecatedWriter();
    try writer.print("\x1b[2J\x1b[H", .{});
    try writer.print("{s}: {s}\n", .{ prompt, query });
    try writer.print("--------------------------------\n", .{});

    var match_index: usize = 0;
    for (items) |item| {
        if (containsInsensitive(item.label, query)) {
            match_index += 1;
            if (no_numeric_selection) {
                try writer.print("{s}\n", .{item.label});
            } else {
                try writer.print("{d}. {s}\n", .{ match_index, item.label });
            }
        }
    }

    if (match_index == 0) {
        try writer.print("(no matches)\n", .{});
    }
    try writer.print("--------------------------------\n", .{});
}

fn containsInsensitive(haystack: []const u8, needle: []const u8) bool {
    if (needle.len == 0) return true;
    if (needle.len > haystack.len) return false;

    var i: usize = 0;
    while (i + needle.len <= haystack.len) : (i += 1) {
        var j: usize = 0;
        while (j < needle.len) : (j += 1) {
            if (std.ascii.toLower(haystack[i + j]) != std.ascii.toLower(needle[j])) {
                break;
            }
        }
        if (j == needle.len) return true;
    }

    return false;
}

fn readItems(allocator: std.mem.Allocator) ![]menu.MenuItem {
    const stdin = std.fs.File.stdin();
    const input = try stdin.readToEndAlloc(allocator, 16 * 1024 * 1024);
    if (input.len == 0) return error.NoInput;

    var items = std.ArrayList(menu.MenuItem).empty;
    errdefer items.deinit(allocator);

    var iter = std.mem.splitScalar(u8, input, '\n');
    while (iter.next()) |line| {
        var trimmed = line;
        if (trimmed.len > 0 and trimmed[trimmed.len - 1] == '\r') {
            trimmed = trimmed[0 .. trimmed.len - 1];
        }
        if (trimmed.len == 0) continue;

        const item = try menu.parseItem(allocator, trimmed, items.items.len, false);
        try items.append(allocator, item);
    }

    if (items.items.len == 0) return error.NoInput;

    return items.toOwnedSlice(allocator);
}
