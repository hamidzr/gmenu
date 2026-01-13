const std = @import("std");
const search = @import("search.zig");

pub const MenuItem = struct {
    label: [:0]const u8,
    index: usize,
    icon: IconKind,
};

pub const IconKind = enum {
    none,
    app,
    file,
    folder,
    info,
};

pub fn parseItem(allocator: std.mem.Allocator, line: []const u8, index: usize, parse_icon: bool) !MenuItem {
    var icon: IconKind = .none;
    var label = line;

    if (parse_icon and line.len >= 3 and line[0] == '[') {
        if (std.mem.indexOfScalar(u8, line, ']')) |close_idx| {
            const raw = line[1..close_idx];
            if (iconFromName(raw)) |kind| {
                icon = kind;
                label = std.mem.trimLeft(u8, line[close_idx + 1 ..], " \t");
            }
        }
    }

    const label_z = try allocator.dupeZ(u8, label);
    return .{ .label = label_z, .index = index, .icon = icon };
}

fn iconFromName(name: []const u8) ?IconKind {
    if (std.ascii.eqlIgnoreCase(name, "app") or std.ascii.eqlIgnoreCase(name, "application")) {
        return .app;
    }
    if (std.ascii.eqlIgnoreCase(name, "file")) {
        return .file;
    }
    if (std.ascii.eqlIgnoreCase(name, "folder") or std.ascii.eqlIgnoreCase(name, "dir") or std.ascii.eqlIgnoreCase(name, "directory")) {
        return .folder;
    }
    if (std.ascii.eqlIgnoreCase(name, "info")) {
        return .info;
    }
    return null;
}

pub const Model = struct {
    items: []MenuItem,
    labels: []const []const u8,
    matches: std.ArrayList(search.Match),
    filtered: std.ArrayList(usize),
    scores: []u32,
    selected: isize,

    pub fn init(allocator: std.mem.Allocator, items: []MenuItem) !Model {
        const labels = try allocator.alloc([]const u8, items.len);
        for (items, 0..) |item, idx| {
            labels[idx] = item.label[0..item.label.len];
        }

        var matches = std.ArrayList(search.Match).empty;
        try matches.ensureTotalCapacity(allocator, items.len);

        var filtered = std.ArrayList(usize).empty;
        try filtered.ensureTotalCapacity(allocator, items.len);

        const scores = try allocator.alloc(u32, items.len);
        @memset(scores, 0);

        return .{
            .items = items,
            .labels = labels,
            .matches = matches,
            .filtered = filtered,
            .scores = scores,
            .selected = -1,
        };
    }

    pub fn deinit(self: *Model, allocator: std.mem.Allocator) void {
        self.matches.deinit(allocator);
        self.filtered.deinit(allocator);
        allocator.free(self.labels);
        allocator.free(self.scores);
    }

    pub fn applyFilter(self: *Model, query: []const u8, opts: search.Options) void {
        @memset(self.scores, 0);
        search.filterIndices(self.labels, query, opts, &self.matches, &self.filtered);
        for (self.matches.items) |match| {
            if (match.index < self.scores.len) {
                self.scores[match.index] = match.score;
            }
        }
        if (self.filtered.items.len == 0) {
            self.selected = -1;
        } else {
            self.selected = 0;
        }
    }

    pub fn moveSelection(self: *Model, delta: isize) void {
        if (self.filtered.items.len == 0) {
            self.selected = -1;
            return;
        }

        const count: isize = @intCast(self.filtered.items.len);
        var next = self.selected;
        if (next < 0) next = 0;
        next += delta;
        while (next < 0) next += count;
        while (next >= count) next -= count;
        self.selected = next;
    }

    pub fn selectedRow(self: *Model) ?usize {
        if (self.selected < 0) return null;
        const row: usize = @intCast(self.selected);
        if (row >= self.filtered.items.len) return null;
        return row;
    }

    pub fn selectedItem(self: *Model) ?MenuItem {
        const row = self.selectedRow() orelse return null;
        const item_index = self.filtered.items[row];
        return self.items[item_index];
    }
};
