const std = @import("std");
const appconfig = @import("../config.zig");
const ipc = @import("../ipc.zig");
const menu = @import("../menu.zig");

const ipc_max_payload: usize = 1024 * 1024;

pub const UpdateKind = enum {
    append,
    prepend,
    set,
};

pub const UpdateSource = enum {
    stdin,
    ipc,
};

pub const ItemUpdate = struct {
    kind: UpdateKind,
    source: UpdateSource,
    line: []const u8,
    batch: u64,
};

pub const UpdateQueue = struct {
    allocator: std.mem.Allocator,
    mutex: std.Thread.Mutex = .{},
    items: std.ArrayList(ItemUpdate),
    next_batch: u64,

    pub fn init(allocator: std.mem.Allocator) UpdateQueue {
        return .{
            .allocator = allocator,
            .items = std.ArrayList(ItemUpdate).empty,
            .next_batch = 1,
        };
    }

    pub fn pushOwned(self: *UpdateQueue, kind: UpdateKind, source: UpdateSource, line: []const u8, batch: u64) void {
        self.mutex.lock();
        defer self.mutex.unlock();
        self.items.append(self.allocator, .{ .kind = kind, .source = source, .line = line, .batch = batch }) catch self.allocator.free(line);
    }

    pub fn nextBatchId(self: *UpdateQueue) u64 {
        self.mutex.lock();
        defer self.mutex.unlock();
        const batch = self.next_batch;
        self.next_batch += 1;
        return batch;
    }

    pub fn reset(self: *UpdateQueue) void {
        self.mutex.lock();
        defer self.mutex.unlock();
        for (self.items.items) |update| {
            self.allocator.free(update.line);
        }
        self.items.clearRetainingCapacity();
        self.next_batch = 1;
    }

    pub fn drain(self: *UpdateQueue) []const ItemUpdate {
        self.mutex.lock();
        defer self.mutex.unlock();
        if (self.items.items.len == 0) {
            return &[_]ItemUpdate{};
        }
        const out = self.allocator.alloc(ItemUpdate, self.items.items.len) catch {
            for (self.items.items) |update| {
                self.allocator.free(update.line);
            }
            self.items.clearRetainingCapacity();
            return &[_]ItemUpdate{};
        };
        @memcpy(out, self.items.items);
        self.items.clearRetainingCapacity();
        return out;
    }
};

pub const QueueState = struct {
    queue: ?*UpdateQueue,
    ipc_path: ?[]const u8,

    pub fn deinit(self: QueueState) void {
        if (self.queue) |queue| {
            queue.reset();
            std.heap.c_allocator.destroy(queue);
        }
        if (self.ipc_path) |path| {
            std.fs.deleteFileAbsolute(path) catch {};
        }
    }
};

pub fn startUpdateQueue(config: appconfig.Config) !QueueState {
    if (!config.follow_stdin and !config.ipc_only) {
        return .{ .queue = null, .ipc_path = null };
    }

    const queue = try std.heap.c_allocator.create(UpdateQueue);
    queue.* = UpdateQueue.init(std.heap.c_allocator);

    const ipc_path = startIpcServer(queue, config.menu_id);
    if (config.ipc_only) {
        queue.reset();
    }

    return .{ .queue = queue, .ipc_path = ipc_path };
}

pub fn followStdinThread(queue: *UpdateQueue) void {
    var reader = std.fs.File.stdin().deprecatedReader();
    while (true) {
        const line_opt = reader.readUntilDelimiterOrEofAlloc(queue.allocator, '\n', 64 * 1024) catch return;
        if (line_opt == null) return;
        var line = line_opt.?;
        const trimmed = std.mem.trimRight(u8, line, "\r\n");
        if (trimmed.len == 0) {
            queue.allocator.free(line);
            continue;
        }
        if (trimmed.len != line.len) {
            const copy = queue.allocator.dupe(u8, trimmed) catch {
                queue.allocator.free(line);
                continue;
            };
            queue.allocator.free(line);
            line = copy;
        }
        queue.pushOwned(.append, .stdin, line, 0);
    }
}

pub fn menuItemFromIpc(allocator: std.mem.Allocator, payload: []const u8) ?menu.MenuItem {
    const parsed = std.json.parseFromSlice(ipc.Item, allocator, payload, .{
        .ignore_unknown_fields = true,
    }) catch return null;
    defer parsed.deinit();

    const item = parsed.value;
    const label = std.mem.trim(u8, item.label, " \t\r\n");
    if (label.len == 0) return null;

    const label_z = allocator.dupeZ(u8, label) catch return null;
    errdefer allocator.free(label_z);
    const icon = menu.iconKindFromName(item.icon);
    const payload_copy = allocator.dupe(u8, payload) catch return null;

    return .{ .label = label_z, .index = 0, .icon = icon, .ipc_payload = payload_copy };
}

fn startIpcServer(queue: *UpdateQueue, menu_id: []const u8) ?[]const u8 {
    const path = ipc.socketPath(std.heap.c_allocator, menu_id) catch return null;
    var server = openIpcServer(path) catch return null;

    const server_ptr = std.heap.c_allocator.create(std.net.Server) catch {
        server.deinit();
        return null;
    };
    server_ptr.* = server;

    _ = std.Thread.spawn(.{}, ipcServerLoop, .{ server_ptr, queue }) catch {
        server.deinit();
        std.heap.c_allocator.destroy(server_ptr);
        return null;
    };

    return path;
}

fn openIpcServer(path: []const u8) !std.net.Server {
    std.fs.deleteFileAbsolute(path) catch |err| switch (err) {
        error.FileNotFound => {},
        else => return err,
    };

    const address = try std.net.Address.initUnix(path);
    return std.net.Address.listen(address, .{});
}

fn ipcServerLoop(server: *std.net.Server, queue: *UpdateQueue) void {
    while (true) {
        const conn = server.accept() catch continue;
        handleIpcConnection(conn.stream, queue);
        conn.stream.close();
    }
}

fn handleIpcConnection(stream: std.net.Stream, queue: *UpdateQueue) void {
    var buf: [4096]u8 = undefined;
    var reader = stream.reader(&buf);
    const io_reader = reader.interface();

    while (true) {
        const line_opt = readLineAlloc(io_reader, std.heap.c_allocator, 64) catch return;
        if (line_opt == null) return;
        const line = line_opt.?;
        defer std.heap.c_allocator.free(line);

        const trimmed = std.mem.trim(u8, line, " \t\r");
        if (trimmed.len == 0) continue;

        const payload_len = std.fmt.parseInt(usize, trimmed, 10) catch continue;
        if (payload_len == 0 or payload_len > ipc_max_payload) return;

        const payload = std.heap.c_allocator.alloc(u8, payload_len) catch return;
        defer std.heap.c_allocator.free(payload);

        io_reader.readSliceAll(payload) catch return;

        handleIpcPayload(queue, payload);
    }
}

fn readLineAlloc(reader: *std.Io.Reader, allocator: std.mem.Allocator, max_len: usize) !?[]u8 {
    var buffer = std.ArrayList(u8).empty;
    errdefer buffer.deinit(allocator);

    while (buffer.items.len < max_len) {
        var byte: [1]u8 = undefined;
        const n = reader.readSliceShort(&byte) catch return null;
        if (n == 0) {
            if (buffer.items.len == 0) return null;
            break;
        }
        if (byte[0] == '\n') break;
        if (byte[0] == '\r') continue;
        try buffer.append(allocator, byte[0]);
    }

    const slice = try buffer.toOwnedSlice(allocator);
    return slice;
}

fn handleIpcPayload(queue: *UpdateQueue, payload: []const u8) void {
    const parsed = std.json.parseFromSlice(ipc.Message, std.heap.c_allocator, payload, .{ .ignore_unknown_fields = true }) catch return;
    defer parsed.deinit();

    const kind = updateKindFromCommand(parsed.value.cmd) orelse return;
    const items = parsed.value.items orelse return;

    const batch_id = queue.nextBatchId();
    for (items) |item| {
        const label = std.mem.trim(u8, item.label, " \t\r\n");
        if (label.len == 0) continue;

        var json_out: std.Io.Writer.Allocating = .init(queue.allocator);
        defer json_out.deinit();
        std.json.Stringify.value(item, .{}, &json_out.writer) catch continue;
        const payload_copy = queue.allocator.dupe(u8, json_out.written()) catch continue;
        queue.pushOwned(kind, .ipc, payload_copy, batch_id);
    }
}

fn updateKindFromCommand(command: []const u8) ?UpdateKind {
    if (std.ascii.eqlIgnoreCase(command, "set")) return .set;
    if (std.ascii.eqlIgnoreCase(command, "append")) return .append;
    if (std.ascii.eqlIgnoreCase(command, "prepend")) return .prepend;
    return null;
}
