const std = @import("std");
const app = @import("app.zig");
const cli = @import("cli.zig");

pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const config = cli.parse(allocator) catch |err| {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: {s}\n", .{@errorName(err)}) catch {};
        std.process.exit(1);
    };

    try app.run(config);
}
