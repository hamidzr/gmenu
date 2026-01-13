const std = @import("std");
const search = @import("search.zig");

pub const MenuItem = struct {
    label: [:0]const u8,
    index: usize,
};

pub const Model = struct {
    items: []MenuItem,
    labels: []const []const u8,
    matches: std.ArrayList(search.Match),
    filtered: std.ArrayList(usize),
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

        return .{
            .items = items,
            .labels = labels,
            .matches = matches,
            .filtered = filtered,
            .selected = -1,
        };
    }

    pub fn deinit(self: *Model, allocator: std.mem.Allocator) void {
        self.matches.deinit(allocator);
        self.filtered.deinit(allocator);
        allocator.free(self.labels);
    }

    pub fn applyFilter(self: *Model, query: []const u8, opts: search.Options) void {
        search.filterIndices(self.labels, query, opts, &self.matches, &self.filtered);
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
