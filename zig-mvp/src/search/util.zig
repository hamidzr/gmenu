const std = @import("std");

pub fn hasUpperAscii(text: []const u8) bool {
    for (text) |c| {
        if (std.ascii.isUpper(c)) return true;
    }
    return false;
}

pub fn charsEqual(a: u8, b: u8, ignore_case: bool) bool {
    if (!ignore_case) return a == b;
    return std.ascii.toLower(a) == std.ascii.toLower(b);
}
