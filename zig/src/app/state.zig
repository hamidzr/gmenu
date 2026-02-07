const std = @import("std");
const objc = @import("objc");
const appconfig = @import("../config.zig");
const menu = @import("../menu.zig");
const updates = @import("updates.zig");

pub const AppState = struct {
    model: menu.Model,
    table_view: objc.Object,
    index_column: ?objc.Object,
    text_field: objc.Object,
    match_label: objc.Object,
    handler: objc.Object,
    config: appconfig.Config,
    pid_path: ?[]const u8,
    ipc_path: ?[]const u8,
    allocator: std.mem.Allocator,
    update_queue: ?*updates.UpdateQueue,
    had_focus: bool,
};

pub const digit_labels = [_][:0]const u8{ "1", "2", "3", "4", "5", "6", "7", "8", "9" };

pub var g_state: ?*AppState = null;
