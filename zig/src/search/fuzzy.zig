const std = @import("std");
const score = @import("score.zig");
const types = @import("types.zig");
const util = @import("util.zig");

pub fn directSearch(labels: []const []const u8, query: []const u8, matches: *std.ArrayList(types.Match)) void {
    for (labels, 0..) |label, idx| {
        if (containsSmartCase(label, query)) {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
}

pub fn fuzzyTokenSearch(
    labels: []const []const u8,
    query: []const u8,
    min_consecutive: usize,
    matches: *std.ArrayList(types.Match),
    scratch_indices: *std.ArrayList(usize),
) void {
    scratch_indices.clearRetainingCapacity();
    for (labels, 0..) |_, idx| {
        scratch_indices.appendAssumeCapacity(idx);
    }

    var tokens = std.mem.splitScalar(u8, query, ' ');
    var saw_token = false;
    while (tokens.next()) |token| {
        if (token.len == 0) continue;
        saw_token = true;
        fuzzySearchBruteOnIndices(labels, scratch_indices.items, token, min_consecutive, matches);
        scratch_indices.clearRetainingCapacity();
        for (matches.items) |match| {
            scratch_indices.appendAssumeCapacity(match.index);
        }
        if (scratch_indices.items.len == 0) return;
    }

    if (!saw_token) {
        matches.clearRetainingCapacity();
        for (labels, 0..) |_, idx| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
}

pub fn fuzzySearchBrute(
    labels: []const []const u8,
    query: []const u8,
    min_consecutive: usize,
    matches: *std.ArrayList(types.Match),
) void {
    matches.clearRetainingCapacity();
    if (query.len == 0) {
        for (labels, 0..) |_, idx| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
        return;
    }

    for (labels, 0..) |label, idx| {
        if (containsSmartCase(label, query)) {
            // direct substring matches should rank ahead of fuzzy subsequence matches when we sort later
            matches.appendAssumeCapacity(.{ .index = idx, .score = 1 });
        }
    }
    for (labels, 0..) |label, idx| {
        if (!containsSmartCase(label, query) and fuzzyContainsConsec(label, query, true, min_consecutive)) {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
}

fn fuzzySearchBruteOnIndices(
    labels: []const []const u8,
    indices: []const usize,
    query: []const u8,
    min_consecutive: usize,
    matches: *std.ArrayList(types.Match),
) void {
    matches.clearRetainingCapacity();
    if (query.len == 0) {
        for (indices) |idx| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
        return;
    }

    for (indices) |idx| {
        const label = labels[idx];
        if (containsSmartCase(label, query)) {
            // direct substring matches should rank ahead of fuzzy subsequence matches when we sort later
            matches.appendAssumeCapacity(.{ .index = idx, .score = 1 });
        }
    }
    for (indices) |idx| {
        const label = labels[idx];
        if (!containsSmartCase(label, query) and fuzzyContainsConsec(label, query, true, min_consecutive)) {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
}

pub fn fuzzyScoreSearch(
    labels: []const []const u8,
    query: []const u8,
    preserve_order: bool,
    matches: *std.ArrayList(types.Match),
) void {
    matches.clearRetainingCapacity();
    if (query.len == 0) return;

    for (labels, 0..) |label, idx| {
        if (score.sahilmScore(query, label)) |score_val| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = score_val });
        }
    }

    std.sort.insertion(types.Match, matches.items, {}, scoreDescIndexAsc);
    filterOutUnlikelyMatches(matches);

    if (preserve_order) {
        std.sort.insertion(types.Match, matches.items, {}, indexAsc);
    }
}

fn filterOutUnlikelyMatches(matches: *std.ArrayList(types.Match)) void {
    if (matches.items.len == 0) return;
    if (matches.items[0].score <= 0) return;

    var write: usize = 0;
    for (matches.items) |match| {
        if (match.score > 0) {
            matches.items[write] = match;
            write += 1;
        }
    }
    matches.items = matches.items[0..write];
}

fn containsSmartCase(haystack: []const u8, needle: []const u8) bool {
    if (util.hasUpperAscii(needle)) {
        return std.mem.indexOf(u8, haystack, needle) != null;
    }
    return containsInsensitive(haystack, needle);
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

fn fuzzyContainsConsec(s: []const u8, query: []const u8, ignore_case: bool, min_consecutive: usize) bool {
    if (query.len == 0) return true;

    var min = min_consecutive;
    if (min < 1) min = 1;
    if (min > query.len) min = query.len;
    if (s.len < min) return false;

    var i: usize = 0;
    while (i + min <= s.len) : (i += 1) {
        var k: usize = 0;
        while (k < min) : (k += 1) {
            if (!util.charsEqual(s[i + k], query[k], ignore_case)) {
                break;
            }
        }
        if (k != min) continue;

        var query_index = min;
        var j = i + min;
        while (j < s.len and query_index < query.len) : (j += 1) {
            if (util.charsEqual(s[j], query[query_index], ignore_case)) {
                query_index += 1;
            }
        }
        if (query_index == query.len) return true;
    }

    return false;
}

pub fn scoreDescIndexAsc(_: void, a: types.Match, b: types.Match) bool {
    if (a.score == b.score) return a.index < b.index;
    return a.score > b.score;
}

pub fn indexAsc(_: void, a: types.Match, b: types.Match) bool {
    return a.index < b.index;
}
