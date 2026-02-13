const std = @import("std");

const first_char_match_bonus: i32 = 10;
const match_following_separator_bonus: i32 = 20;
const camel_case_match_bonus: i32 = 20;
const adjacent_match_bonus: i32 = 5;
const unmatched_leading_char_penalty: i32 = -5;
const max_unmatched_leading_char_penalty: i32 = -15;

pub fn sahilmScore(pattern: []const u8, candidate: []const u8) ?i32 {
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
                    best_score += @max(penalty, max_unmatched_leading_char_penalty);
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
