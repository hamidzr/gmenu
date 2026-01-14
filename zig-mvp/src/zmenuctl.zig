const std = @import("std");
const ipc = @import("ipc.zig");

pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    const args = try std.process.argsAlloc(allocator);
    defer std.process.argsFree(allocator, args);

    var menu_id: []const u8 = "";
    var socket_override: ?[]const u8 = null;
    var read_stdin = false;

    var cmd_index: ?usize = null;
    var i: usize = 1;
    while (i < args.len) : (i += 1) {
        const arg = args[i];
        if (std.mem.eql(u8, arg, "--menu-id") or std.mem.eql(u8, arg, "-m")) {
            i += 1;
            if (i >= args.len) return usage();
            menu_id = args[i];
            continue;
        }
        if (std.mem.eql(u8, arg, "--socket")) {
            i += 1;
            if (i >= args.len) return usage();
            socket_override = args[i];
            continue;
        }
        if (std.mem.eql(u8, arg, "--stdin")) {
            read_stdin = true;
            continue;
        }
        if (std.mem.eql(u8, arg, "--help") or std.mem.eql(u8, arg, "-h")) {
            return usage();
        }

        cmd_index = i;
        break;
    }

    if (cmd_index == null) {
        return usage();
    }

    const cmd = args[cmd_index.?];
    if (!isSupportedCommand(cmd)) {
        return usage();
    }

    var items = std.ArrayList(ipc.Item).empty;
    defer items.deinit(allocator);

    if (read_stdin) {
        try readItemsFromStdin(allocator, &items);
    } else {
        const start = cmd_index.? + 1;
        if (start < args.len) {
            for (args[start..]) |arg| {
                if (arg.len == 0) continue;
                try items.append(allocator, .{ .label = arg });
            }
        }
    }

    if (items.items.len == 0) {
        std.fs.File.stderr().deprecatedWriter().print("zmenuctl: no items provided\n", .{}) catch {};
        std.process.exit(1);
    }

    const msg = ipc.Message{
        .cmd = cmd,
        .items = items.items,
    };

    var json_out: std.Io.Writer.Allocating = .init(allocator);
    defer json_out.deinit();
    try std.json.Stringify.value(msg, .{}, &json_out.writer);
    const payload = json_out.written();

    const socket_path = socket_override orelse try ipc.socketPath(allocator, menu_id);
    const stream = try std.net.connectUnixSocket(socket_path);
    defer stream.close();

    var buf: [4096]u8 = undefined;
    var writer = stream.writer(&buf);
    const header = try std.fmt.allocPrint(allocator, "{d}\n", .{payload.len});
    defer allocator.free(header);
    try writer.interface.writeAll(header);
    try writer.interface.writeAll(payload);
    try writer.interface.flush();
}

fn isSupportedCommand(cmd: []const u8) bool {
    return std.ascii.eqlIgnoreCase(cmd, "set") or std.ascii.eqlIgnoreCase(cmd, "append") or std.ascii.eqlIgnoreCase(cmd, "prepend");
}

fn readItemsFromStdin(allocator: std.mem.Allocator, items: *std.ArrayList(ipc.Item)) !void {
    const stdin = std.fs.File.stdin();
    const input = try stdin.readToEndAlloc(allocator, 16 * 1024 * 1024);

    var iter = std.mem.splitScalar(u8, input, '\n');
    while (iter.next()) |line| {
        var trimmed = line;
        if (trimmed.len > 0 and trimmed[trimmed.len - 1] == '\r') {
            trimmed = trimmed[0 .. trimmed.len - 1];
        }
        if (trimmed.len == 0) continue;
        try items.append(allocator, .{ .label = trimmed });
    }
}

fn usage() void {
    std.fs.File.stdout().deprecatedWriter().print(
        \\zmenuctl usage:
        \\  zmenuctl [--menu-id <id>] [--socket <path>] [--stdin] <set|append|prepend> [items...]
        \\
        \\Examples:
        \\  zmenuctl --menu-id demo set --stdin
        \\  zmenuctl --menu-id demo append "new item"
        \\
    , .{}) catch {};
}
