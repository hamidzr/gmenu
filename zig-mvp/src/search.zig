const std = @import("std");

pub const SearchMethod = enum {
    direct,
    fuzzy,
};

pub const Options = struct {
    method: SearchMethod = .fuzzy,
    preserve_order: bool = false,
    limit: usize = 10,
};

pub const Match = struct {
    index: usize,
    score: u32,
};

pub fn filterIndices(
    labels: []const []const u8,
    query: []const u8,
    opts: Options,
    matches: *std.ArrayList(Match),
    out_indices: *std.ArrayList(usize),
) void {
    matches.clearRetainingCapacity();
    out_indices.clearRetainingCapacity();

    const trimmed = std.mem.trim(u8, query, " \t\r\n");
    if (trimmed.len == 0) {
        const limit = effectiveLimit(opts.limit, labels.len);
        var i: usize = 0;
        while (i < limit) : (i += 1) {
            out_indices.appendAssumeCapacity(i);
        }
        return;
    }

    switch (opts.method) {
        .direct => {
            for (labels, 0..) |label, idx| {
                if (containsInsensitive(label, trimmed)) {
                    matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
                }
            }
        },
        .fuzzy => {
            for (labels, 0..) |label, idx| {
                if (fuzzyScoreTokens(label, trimmed)) |score| {
                    matches.appendAssumeCapacity(.{ .index = idx, .score = score });
                }
            }
        },
    }

    if (opts.method == .fuzzy and !opts.preserve_order) {
        std.sort.insertion(Match, matches.items, {}, matchLess);
    }

    const limit = effectiveLimit(opts.limit, matches.items.len);
    var i: usize = 0;
    while (i < limit) : (i += 1) {
        out_indices.appendAssumeCapacity(matches.items[i].index);
    }
}

pub fn containsInsensitive(haystack: []const u8, needle: []const u8) bool {
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

pub fn fuzzyScoreTokens(label: []const u8, query: []const u8) ?u32 {
    const trimmed = std.mem.trim(u8, query, " \t\r\n");
    if (trimmed.len == 0) return 0;

    var tokens = std.mem.tokenizeAny(u8, trimmed, " \t\r\n");
    var total: u32 = 0;
    var saw_token = false;
    while (tokens.next()) |token| {
        if (token.len == 0) continue;
        saw_token = true;
        const score = fuzzyScoreToken(label, token) orelse return null;
        total += score;
    }

    if (!saw_token) return 0;
    return total;
}

fn fuzzyScoreToken(label: []const u8, token: []const u8) ?u32 {
    if (token.len == 0) return 0;

    var score: u32 = 0;
    var last_index: isize = -1;
    var has_adjacent = false;

    var i: usize = 0;
    while (i < token.len) : (i += 1) {
        const needle = std.ascii.toLower(token[i]);
        var found = false;
        var j: usize = @intCast(last_index + 1);
        while (j < label.len) : (j += 1) {
            if (std.ascii.toLower(label[j]) == needle) {
                if (last_index >= 0 and @as(isize, @intCast(j)) == last_index + 1) {
                    has_adjacent = true;
                }
                score += @intCast(j);
                last_index = @intCast(j);
                found = true;
                break;
            }
        }
        if (!found) return null;
    }

    if (token.len >= 2 and !has_adjacent) return null;
    return score;
}

fn effectiveLimit(limit: usize, count: usize) usize {
    if (limit == 0 or limit > count) return count;
    return limit;
}

fn matchLess(_: void, a: Match, b: Match) bool {
    if (a.score == b.score) return a.index < b.index;
    return a.score < b.score;
}

test "direct match is case-insensitive" {
    const labels = [_][]const u8{ "Alpha", "bravo", "CHARLIE" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "BR", .{ .method = .direct }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{1}, out.items);
}

test "fuzzy token requires adjacent match for length >= 2" {
    try std.testing.expect(fuzzyScoreTokens("abcdef", "ab") != null);
    try std.testing.expect(fuzzyScoreTokens("abcdef", "ac") == null);
}

test "fuzzy tokens require all tokens" {
    try std.testing.expect(fuzzyScoreTokens("alpha bravo", "al br") != null);
    try std.testing.expect(fuzzyScoreTokens("alpha bravo", "al zz") == null);
}

test "fuzzy sorting honors preserve_order" {
    const labels = [_][]const u8{ "abXc", "abc" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "abc", .{ .method = .fuzzy, .preserve_order = false }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 1, 0 }, out.items);

    filterIndices(labels[0..], "abc", .{ .method = .fuzzy, .preserve_order = true }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 0, 1 }, out.items);
}

test "limit caps results" {
    const labels = [_][]const u8{
        "item0",  "item1",  "item2",  "item3", "item4",
        "item5",  "item6",  "item7",  "item8", "item9",
        "item10", "item11", "item12",
    };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "item", .{ .method = .direct, .limit = 10 }, &matches, &out);
    try std.testing.expectEqual(@as(usize, 10), out.items.len);
    try std.testing.expectEqual(@as(usize, 0), out.items[0]);
    try std.testing.expectEqual(@as(usize, 9), out.items[9]);
}
