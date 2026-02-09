const std = @import("std");
const fuzzy = @import("search/fuzzy.zig");
const levenshtein = @import("search/levenshtein.zig");
const types = @import("search/types.zig");

pub const SearchMethod = types.SearchMethod;
pub const Options = types.Options;
pub const Match = types.Match;

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
        for (labels, 0..) |_, idx| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
        const limit = effectiveLimit(opts.limit, matches.items.len);
        var i: usize = 0;
        while (i < limit) : (i += 1) {
            out_indices.appendAssumeCapacity(matches.items[i].index);
        }
        return;
    }

    switch (opts.method) {
        .direct => fuzzy.directSearch(labels, query, matches),
        .fuzzy => fuzzy.fuzzyTokenSearch(labels, trimmed, 2, matches, out_indices),
        .default => fuzzy.fuzzyTokenSearch(labels, trimmed, 2, matches, out_indices),
        .fuzzy3 => fuzzy.fuzzySearchBrute(labels, query, 2, matches),
        .fuzzy1 => fuzzy.fuzzyScoreSearch(labels, query, opts.preserve_order, matches),
    }

    if (opts.levenshtein_fallback and matches.items.len < 3 and trimmed.len > 0) {
        levenshtein.appendFallbackMatches(labels, trimmed, matches);
        std.sort.insertion(Match, matches.items, {}, fuzzy.scoreDescIndexAsc);
    }

    out_indices.clearRetainingCapacity();
    const limit = effectiveLimit(opts.limit, matches.items.len);
    var i: usize = 0;
    while (i < limit) : (i += 1) {
        out_indices.appendAssumeCapacity(matches.items[i].index);
    }
}

fn effectiveLimit(limit: usize, count: usize) usize {
    if (limit == 0 or limit > count) return count;
    return limit;
}

test "direct smart-case matches only uppercase when query has uppercase" {
    const labels = [_][]const u8{ "Alpha", "bravo", "BRAVO" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "BR", .{ .method = .direct }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{2}, out.items);

    out.clearRetainingCapacity();
    matches.clearRetainingCapacity();
    filterIndices(labels[0..], "br", .{ .method = .direct }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 1, 2 }, out.items);
}

test "fuzzy tokenized requires all tokens" {
    const labels = [_][]const u8{ "alpha bravo", "alpha", "bravo" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "al br", .{ .method = .fuzzy }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{0}, out.items);

    out.clearRetainingCapacity();
    matches.clearRetainingCapacity();
    filterIndices(labels[0..], "al zz", .{ .method = .fuzzy, .levenshtein_fallback = false }, &matches, &out);
    try std.testing.expectEqual(@as(usize, 0), out.items.len);
}

test "fuzzy3 orders direct matches before fuzzy" {
    const labels = [_][]const u8{ "abXc", "abc" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "abc", .{ .method = .fuzzy3, .levenshtein_fallback = false }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 1, 0 }, out.items);
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

test "levenshtein fallback returns closest matches when enabled" {
    const labels = [_][]const u8{ "ab", "abce", "wxyz" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "abc", .{ .method = .direct }, &matches, &out);
    try std.testing.expectEqual(@as(usize, 2), out.items.len);
    try std.testing.expectEqual(@as(usize, 1), out.items[0]);
    try std.testing.expectEqual(@as(usize, 0), out.items[1]);
}

test "fuzzy matches agenda after direct hit for anda" {
    const labels = [_][]const u8{ "agenda", "anda" };
    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "anda", .{ .method = .fuzzy }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 1, 0 }, out.items);
}

test "today query returns only lock screen and today note" {
    const labels = [_][]const u8{
        "[OS] Lock Screen - put the display to sleep immediately",
        "Edit: ~/notes/today.md",
        "Edit: ~/notes/agenda.md",
        "Open: ~/notes/week.md",
    };

    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "today", .{ .method = .fuzzy, .levenshtein_fallback = true }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 1, 0 }, out.items);
}

test "fuzzy ranks direct substring above subsequence match" {
    const labels = [_][]const u8{
        // matches "today" via subsequence (to + d + a + y), but not as a verbatim substring
        "to d a y",
        // matches "today" as a direct substring
        "today.md",
    };

    var matches = std.ArrayList(Match).empty;
    var out = std.ArrayList(usize).empty;
    defer matches.deinit(std.testing.allocator);
    defer out.deinit(std.testing.allocator);

    try matches.ensureTotalCapacity(std.testing.allocator, labels.len);
    try out.ensureTotalCapacity(std.testing.allocator, labels.len);

    filterIndices(labels[0..], "today", .{ .method = .fuzzy, .levenshtein_fallback = true }, &matches, &out);
    try std.testing.expectEqualSlices(usize, &[_]usize{ 1, 0 }, out.items);
}
