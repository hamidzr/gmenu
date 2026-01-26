const std = @import("std");
const types = @import("types.zig");
const util = @import("util.zig");

const levenshtein_distance_threshold: usize = 2;

pub fn appendFallbackMatches(labels: []const []const u8, query: []const u8, matches: *std.ArrayList(types.Match)) void {
    if (query.len == 0) return;

    var max_len: usize = query.len;
    for (labels) |label| {
        if (label.len > max_len) max_len = label.len;
    }

    const row_buffer = std.heap.page_allocator.alloc(usize, max_len + 1) catch return;
    defer std.heap.page_allocator.free(row_buffer);

    const ignore_case = !util.hasUpperAscii(query);
    for (labels, 0..) |label, idx| {
        var already_matched = false;
        for (matches.items) |match| {
            if (match.index == idx) {
                already_matched = true;
                break;
            }
        }
        if (already_matched) continue;

        const distance = levenshteinDistance(query, label, row_buffer[0 .. label.len + 1], ignore_case);
        if (distance <= levenshtein_distance_threshold) {
            matches.appendAssumeCapacity(.{
                .index = idx,
                .score = distanceScore(distance),
            });
        }
    }
}

fn levenshteinDistance(a: []const u8, b: []const u8, row: []usize, ignore_case: bool) usize {
    const b_len = b.len;
    var j: usize = 0;
    while (j <= b_len) : (j += 1) {
        row[j] = j;
    }

    var i: usize = 0;
    while (i < a.len) : (i += 1) {
        var prev = row[0];
        row[0] = i + 1;
        j = 0;
        while (j < b_len) : (j += 1) {
            const old = row[j + 1];
            const cost: usize = if (util.charsEqual(a[i], b[j], ignore_case)) 0 else 1;
            const deletion = old + 1;
            const insertion = row[j] + 1;
            const substitution = prev + cost;
            row[j + 1] = min3(deletion, insertion, substitution);
            prev = old;
        }
    }

    return row[b_len];
}

fn min3(a: usize, b: usize, c: usize) usize {
    var out = if (a < b) a else b;
    if (c < out) out = c;
    return out;
}

fn distanceScore(distance: usize) i32 {
    const capped: i32 = if (distance > std.math.maxInt(i32)) std.math.maxInt(i32) else @as(i32, @intCast(distance));
    return -capped;
}
