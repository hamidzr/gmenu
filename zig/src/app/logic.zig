const std = @import("std");
const objc = @import("objc");
const cache = @import("../cache.zig");
const exit_codes = @import("../exit_codes.zig");
const menu = @import("../menu.zig");
const pid = @import("../pid.zig");
const objc_helpers = @import("objc_helpers.zig");
const state = @import("state.zig");

pub fn quit(app_state: *state.AppState, code: u8) void {
    if (app_state.pid_path) |path| {
        pid.remove(path);
    }
    if (app_state.ipc_path) |path| {
        std.fs.deleteFileAbsolute(path) catch {};
    }
    std.process.exit(code);
}

pub fn updateSelection(app_state: *state.AppState) void {
    const row = app_state.model.selectedRow() orelse {
        app_state.table_view.msgSend(void, "deselectAll:", .{@as(objc.c.id, null)});
        return;
    };

    const NSIndexSet = objc.getClass("NSIndexSet").?;
    const index_set = NSIndexSet.msgSend(objc.Object, "indexSetWithIndex:", .{@as(c_ulong, @intCast(row))});
    app_state.table_view.msgSend(void, "selectRowIndexes:byExtendingSelection:", .{ index_set, false });
    app_state.table_view.msgSend(void, "scrollRowToVisible:", .{@as(c_long, @intCast(row))});
}

pub fn updateMatchLabel(app_state: *state.AppState) void {
    var buf: [32]u8 = undefined;
    const label_z = std.fmt.bufPrintZ(
        &buf,
        "[{d}/{d}]",
        .{ app_state.model.match_count, app_state.model.items.len },
    ) catch return;
    app_state.match_label.msgSend(void, "setStringValue:", .{objc_helpers.nsString(label_z)});
}

pub fn applyFilter(app_state: *state.AppState, query: []const u8) void {
    app_state.model.applyFilter(query, app_state.config.search);
    if (app_state.index_column) |index_column| {
        const enabled = app_state.config.numericSelectionEnabled(app_state.model.filtered.items.len, query);
        index_column.msgSend(void, "setHidden:", .{!enabled});
    }
    app_state.table_view.msgSend(void, "reloadData", .{});
    updateSelection(app_state);
    updateMatchLabel(app_state);
    if (app_state.config.auto_accept and app_state.model.filtered.items.len == 1) {
        acceptSelection(app_state);
    }
}

pub fn moveSelection(app_state: *state.AppState, delta: isize) void {
    app_state.model.moveSelection(delta);
    updateSelection(app_state);
}

pub fn currentQuery(app_state: *state.AppState) []const u8 {
    const text = app_state.text_field.msgSend(objc.Object, "stringValue", .{});
    const utf8_ptr = text.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return "";
    return std.mem.sliceTo(utf8_ptr.?, 0);
}

pub fn saveCache(app_state: *state.AppState, selection: []const u8) void {
    if (app_state.config.menu_id.len == 0) return;
    const query = currentQuery(app_state);
    cache.save(app_state.allocator, app_state.config.menu_id, .{
        .last_query = query,
        .last_selection = selection,
        .last_selection_time = std.time.timestamp(),
    }) catch {};
}

pub fn acceptSelection(app_state: *state.AppState) void {
    if (app_state.config.ipc_only) {
        if (app_state.model.selectedItem()) |item| {
            saveCache(app_state, item.label);
            const payload = item.ipc_payload orelse item.label;
            std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{payload}) catch {};
            quit(app_state, 0);
        }
        if (app_state.model.filtered.items.len == 0) {
            quit(app_state, exit_codes.user_canceled);
        }
        const item_index = app_state.model.filtered.items[0];
        const item = app_state.model.items[item_index];
        saveCache(app_state, item.label);
        const payload = item.ipc_payload orelse item.label;
        std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{payload}) catch {};
        quit(app_state, 0);
    }

    if (app_state.model.filtered.items.len == 0) {
        if (!app_state.config.accept_custom_selection) {
            return;
        }
        const query = currentQuery(app_state);
        saveCache(app_state, query);
        std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{query}) catch {};
        quit(app_state, 0);
    }

    if (app_state.model.selectedItem()) |item| {
        saveCache(app_state, item.label);
        std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{item.label}) catch {};
        quit(app_state, 0);
    }

    const item_index = app_state.model.filtered.items[0];
    const item = app_state.model.items[item_index];
    saveCache(app_state, item.label);
    std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{item.label}) catch {};
    quit(app_state, 0);
}

pub fn readItems(allocator: std.mem.Allocator, parse_icons: bool) ![]menu.MenuItem {
    var input = try menu.readStdinLines(allocator, menu.stdin_max_bytes);
    defer input.deinit(allocator);

    if (input.lines.len == 0) return error.NoInput;

    var items = std.ArrayList(menu.MenuItem).empty;
    errdefer items.deinit(allocator);

    for (input.lines) |line| {
        const item = try menu.parseItem(allocator, line, items.items.len, parse_icons);
        try items.append(allocator, item);
    }

    if (items.items.len == 0) return error.NoInput;

    return items.toOwnedSlice(allocator);
}
