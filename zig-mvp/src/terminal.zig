const std = @import("std");
const appconfig = @import("config.zig");
const cache = @import("cache.zig");
const pid = @import("pid.zig");
const search = @import("search.zig");

const MenuItem = struct {
    label: [:0]const u8,
    index: usize,
};

pub fn run(config: appconfig.Config, allocator: std.mem.Allocator) !void {
    const items = readItems(allocator) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: stdin is empty\n", .{}) catch {};
        std.process.exit(1);
    };

    const pid_path = pid.create(allocator, config.menu_id) catch |err| {
        _ = err;
        std.fs.File.stderr().deprecatedWriter().print("zmenu: another instance is running\n", .{}) catch {};
        std.process.exit(1);
    };
    defer pid.remove(pid_path);

    var query = config.initial_query;
    if (query.len == 0 and config.menu_id.len > 0) {
        if (cache.load(allocator, config.menu_id)) |cached| {
            if (cached) |state| {
                if (state.last_query.len > 0) {
                    query = state.last_query;
                }
            }
        } else |_| {}
    }
    if (query.len == 0) {
        query = try readQueryFromTTY(allocator, config.placeholder);
    }

    const labels = try allocator.alloc([]const u8, items.len);
    for (items, 0..) |item, idx| {
        labels[idx] = item.label[0..item.label.len];
    }

    var matches = std.ArrayList(search.Match).empty;
    try matches.ensureTotalCapacity(allocator, items.len);
    var filtered = std.ArrayList(usize).empty;
    try filtered.ensureTotalCapacity(allocator, items.len);

    search.filterIndices(labels, query, config.search, &matches, &filtered);
    if (filtered.items.len == 0) {
        if (!config.accept_custom_selection) {
            std.process.exit(1);
        }
        saveCache(allocator, config, query, query);
        std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{query}) catch {};
        std.process.exit(0);
    }

    const item_index = filtered.items[0];
    const item = items[item_index];
    saveCache(allocator, config, query, item.label);
    std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{item.label}) catch {};
    std.process.exit(0);
}

fn readItems(allocator: std.mem.Allocator) ![]MenuItem {
    const stdin = std.fs.File.stdin();
    const input = try stdin.readToEndAlloc(allocator, 16 * 1024 * 1024);
    if (input.len == 0) return error.NoInput;

    var items = std.ArrayList(MenuItem).empty;
    errdefer items.deinit(allocator);

    var iter = std.mem.splitScalar(u8, input, '\n');
    while (iter.next()) |line| {
        var trimmed = line;
        if (trimmed.len > 0 and trimmed[trimmed.len - 1] == '\r') {
            trimmed = trimmed[0 .. trimmed.len - 1];
        }
        if (trimmed.len == 0) continue;

        const label = try allocator.dupeZ(u8, trimmed);
        try items.append(allocator, .{ .label = label, .index = items.items.len });
    }

    if (items.items.len == 0) return error.NoInput;

    return items.toOwnedSlice(allocator);
}

fn readQueryFromTTY(allocator: std.mem.Allocator, prompt: [:0]const u8) ![]const u8 {
    var tty = try std.fs.openFileAbsolute("/dev/tty", .{});
    defer tty.close();

    try tty.deprecatedWriter().print("{s}", .{prompt});
    const line_opt = try tty.deprecatedReader().readUntilDelimiterOrEofAlloc(allocator, '\n', 4096);
    if (line_opt == null) return "";
    const line = line_opt.?;
    return std.mem.trimRight(u8, line, "\r\n");
}

fn saveCache(allocator: std.mem.Allocator, config: appconfig.Config, query: []const u8, selection: []const u8) void {
    if (config.menu_id.len == 0) return;
    cache.save(allocator, config.menu_id, .{
        .last_query = query,
        .last_selection = selection,
    }) catch {};
}
