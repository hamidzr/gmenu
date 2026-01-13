const app = @import("app.zig");
const config = @import("config.zig");

pub fn main() !void {
    try app.run(config.defaults());
}
