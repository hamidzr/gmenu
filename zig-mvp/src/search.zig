const std = @import("std");

pub const SearchMethod = enum {
    direct,
    fuzzy,
    fuzzy1,
    fuzzy3,
    default,
};

pub const Options = struct {
    method: SearchMethod = .fuzzy,
    preserve_order: bool = false,
    limit: usize = 10,
};

pub const Match = struct {
    index: usize,
    score: i32,
};

const first_char_match_bonus: i32 = 10;
const match_following_separator_bonus: i32 = 20;
const camel_case_match_bonus: i32 = 20;
const adjacent_match_bonus: i32 = 5;
const unmatched_leading_char_penalty: i32 = -5;
const max_unmatched_leading_char_penalty: i32 = -15;

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
        .direct => directSearch(labels, query, matches),
        .fuzzy => fuzzyTokenSearch(labels, trimmed, 2, matches, out_indices),
        .default => fuzzyTokenSearch(labels, trimmed, 2, matches, out_indices),
        .fuzzy3 => fuzzySearchBrute(labels, query, 2, matches),
        .fuzzy1 => fuzzyScoreSearch(labels, query, opts.preserve_order, matches),
    }

    out_indices.clearRetainingCapacity();
    const limit = effectiveLimit(opts.limit, matches.items.len);
    var i: usize = 0;
    while (i < limit) : (i += 1) {
        out_indices.appendAssumeCapacity(matches.items[i].index);
    }
}

fn directSearch(labels: []const []const u8, query: []const u8, matches: *std.ArrayList(Match)) void {
    for (labels, 0..) |label, idx| {
        if (containsSmartCase(label, query)) {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
}

fn fuzzyTokenSearch(
    labels: []const []const u8,
    query: []const u8,
    min_consecutive: usize,
    matches: *std.ArrayList(Match),
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

fn fuzzySearchBrute(labels: []const []const u8, query: []const u8, min_consecutive: usize, matches: *std.ArrayList(Match)) void {
    matches.clearRetainingCapacity();
    if (query.len == 0) {
        for (labels, 0..) |_, idx| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
        return;
    }

    for (labels, 0..) |label, idx| {
        if (containsSmartCase(label, query)) {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
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
    matches: *std.ArrayList(Match),
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
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
    for (indices) |idx| {
        const label = labels[idx];
        if (!containsSmartCase(label, query) and fuzzyContainsConsec(label, query, true, min_consecutive)) {
            matches.appendAssumeCapacity(.{ .index = idx, .score = 0 });
        }
    }
}

fn fuzzyScoreSearch(labels: []const []const u8, query: []const u8, preserve_order: bool, matches: *std.ArrayList(Match)) void {
    matches.clearRetainingCapacity();
    if (query.len == 0) return;

    for (labels, 0..) |label, idx| {
        if (sahilmScore(query, label)) |score| {
            matches.appendAssumeCapacity(.{ .index = idx, .score = score });
        }
    }

    std.sort.insertion(Match, matches.items, {}, scoreDescIndexAsc);
    filterOutUnlikelyMatches(matches);

    if (preserve_order) {
        std.sort.insertion(Match, matches.items, {}, indexAsc);
    }
}

fn filterOutUnlikelyMatches(matches: *std.ArrayList(Match)) void {
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
    if (hasUpperAscii(needle)) {
        return std.mem.indexOf(u8, haystack, needle) != null;
    }
    return containsInsensitive(haystack, needle);
}

fn hasUpperAscii(text: []const u8) bool {
    for (text) |c| {
        if (std.ascii.isUpper(c)) return true;
    }
    return false;
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
            if (!charsEqual(s[i + k], query[k], ignore_case)) {
                break;
            }
        }
        if (k != min) continue;

        var query_index = min;
        var j = i + min;
        while (j < s.len and query_index < query.len) : (j += 1) {
            if (charsEqual(s[j], query[query_index], ignore_case)) {
                query_index += 1;
            }
        }
        if (query_index == query.len) return true;
    }

    return false;
}

fn charsEqual(a: u8, b: u8, ignore_case: bool) bool {
    if (!ignore_case) return a == b;
    return std.ascii.toLower(a) == std.ascii.toLower(b);
}

fn sahilmScore(pattern: []const u8, candidate: []const u8) ?i32 {
    if (pattern.len == 0) return null;
    if (candidate.len == 0) return null;

    var pattern_index: usize = 0;
    var best_score: i32 = -1;
    var matched_index: isize = -1;
    var total_score: i32 = 0;
    var curr_adjacent_bonus: i32 = 0;
    var last: u8 = 0;
    var last_index: isize = -1;
    var last_match_index: isize = -1;
    var matched_count: usize = 0;

    var j: usize = 0;
    while (j < candidate.len) : (j += 1) {
        const candidate_char = candidate[j];

        if (pattern_index < pattern.len and equalFold(candidate_char, pattern[pattern_index])) {
            var score: i32 = 0;
            if (j == 0) score += first_char_match_bonus;
            if (std.ascii.isLower(last) and std.ascii.isUpper(candidate_char)) {
                score += camel_case_match_bonus;
            }
            if (j != 0 and isSeparator(last)) {
                score += match_following_separator_bonus;
            }
            if (matched_count > 0) {
                const bonus = adjacentCharBonus(last_index, last_match_index, curr_adjacent_bonus);
                score += bonus;
                curr_adjacent_bonus += bonus;
            }
            if (score > best_score) {
                best_score = score;
                matched_index = @intCast(j);
            }
        }

        var nextp: u8 = 0;
        if (pattern_index + 1 < pattern.len) {
            nextp = pattern[pattern_index + 1];
        }
        var nextc: u8 = 0;
        if (j + 1 < candidate.len) {
            nextc = candidate[j + 1];
        }

        if (pattern_index < pattern.len and (equalFold(nextp, nextc) or nextc == 0)) {
            if (matched_index > -1) {
                if (matched_count == 0) {
                    const penalty = @as(i32, @intCast(matched_index)) * unmatched_leading_char_penalty;
                    best_score += maxInt(penalty, max_unmatched_leading_char_penalty);
                }
                total_score += best_score;
                matched_count += 1;
                last_match_index = matched_index;
                best_score = -1;
                pattern_index += 1;
            }
        }

        last_index = @intCast(j);
        last = candidate_char;
        if (pattern_index >= pattern.len) break;
    }

    total_score += @as(i32, @intCast(matched_count)) - @as(i32, @intCast(candidate.len));
    if (matched_count == pattern.len) return total_score;
    return null;
}

fn adjacentCharBonus(i: isize, last_match: isize, current_bonus: i32) i32 {
    if (last_match == i) {
        return current_bonus * 2 + adjacent_match_bonus;
    }
    return 0;
}

fn isSeparator(c: u8) bool {
    return switch (c) {
        '/', '-', '_', ' ', '.', '\\' => true,
        else => false,
    };
}

fn equalFold(a: u8, b: u8) bool {
    if (a == b) return true;
    return std.ascii.toLower(a) == std.ascii.toLower(b);
}

fn maxInt(a: i32, b: i32) i32 {
    if (a > b) return a;
    return b;
}

fn effectiveLimit(limit: usize, count: usize) usize {
    if (limit == 0 or limit > count) return count;
    return limit;
}

fn scoreDescIndexAsc(_: void, a: Match, b: Match) bool {
    if (a.score == b.score) return a.index < b.index;
    return a.score > b.score;
}

fn indexAsc(_: void, a: Match, b: Match) bool {
    return a.index < b.index;
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
    filterIndices(labels[0..], "al zz", .{ .method = .fuzzy }, &matches, &out);
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

    filterIndices(labels[0..], "abc", .{ .method = .fuzzy3 }, &matches, &out);
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
