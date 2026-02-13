const std = @import("std");
const objc = @import("objc");
const menu = @import("../menu.zig");
const exit_codes = @import("../exit_codes.zig");
const objc_helpers = @import("objc_helpers.zig");
const updates = @import("updates.zig");
const logic = @import("logic.zig");
const state = @import("state.zig");

const nsString = objc_helpers.nsString;
const columnIsIndex = objc_helpers.columnIsIndex;
const columnIsIcon = objc_helpers.columnIsIcon;
const iconImage = objc_helpers.iconImage;

const NSEventModifierFlagShift: u64 = 1 << 17;
const NSEventModifierFlagControl: u64 = 1 << 18;
const NSEventModifierFlagOption: u64 = 1 << 19;
const NSEventModifierFlagCommand: u64 = 1 << 20;

pub fn controlTextDidChange(target: objc.c.id, sel: objc.c.SEL, notification: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;

    const app_state = state.g_state orelse return;
    if (notification == null) return;

    const notification_obj = objc.Object.fromId(notification);
    const field = notification_obj.msgSend(objc.Object, "object", .{});
    const text = field.msgSend(objc.Object, "stringValue", .{});
    const utf8_ptr = text.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) {
        logic.applyFilter(app_state, "");
        return;
    }

    const query: []const u8 = std.mem.sliceTo(utf8_ptr.?, 0);
    if (currentEventChar()) |event_char| {
        if (event_char.modifiers == 0 and event_char.char >= '1' and event_char.char <= '9') {
            var query_for_mode = query;
            if (query.len > 0 and query[query.len - 1] == event_char.char) {
                query_for_mode = query[0 .. query.len - 1];
            }
            if (handleNumericShortcutWithQuery(app_state, event_char.char, query_for_mode)) {
                return;
            }
        }
    }
    logic.applyFilter(app_state, query);
}

pub fn controlTextViewDoCommandBySelector(
    target: objc.c.id,
    sel: objc.c.SEL,
    control: objc.c.id,
    text_view: objc.c.id,
    command: objc.c.SEL,
) callconv(.c) bool {
    _ = target;
    _ = sel;
    _ = control;
    _ = text_view;

    const app_state = state.g_state orelse return false;

    if (command == objc.sel("moveUp:").value) {
        logic.moveSelection(app_state, -1);
        return true;
    }
    if (command == objc.sel("moveDown:").value) {
        logic.moveSelection(app_state, 1);
        return true;
    }
    if (command == objc.sel("insertTab:").value) {
        logic.moveSelection(app_state, 1);
        return true;
    }
    if (command == objc.sel("insertBacktab:").value) {
        logic.moveSelection(app_state, -1);
        return true;
    }
    if (command == objc.sel("insertText:").value or command == objc.sel("noop:").value) {
        if (currentEventChar()) |event_char| {
            if (event_char.modifiers == 0 and handleNumericShortcut(app_state, event_char.char)) {
                return true;
            }
        }
    }

    return false;
}

pub fn onSubmit(target: objc.c.id, sel: objc.c.SEL, sender: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = sender;

    const app_state = state.g_state orelse return;
    logic.acceptSelection(app_state);
}

pub fn numberOfRowsInTableView(target: objc.c.id, sel: objc.c.SEL, table: objc.c.id) callconv(.c) c_long {
    _ = target;
    _ = sel;
    _ = table;

    const app_state = state.g_state orelse return 0;
    return @intCast(app_state.model.filtered.items.len);
}

pub fn tableViewObjectValue(
    target: objc.c.id,
    sel: objc.c.SEL,
    table: objc.c.id,
    column: objc.c.id,
    row: c_long,
) callconv(.c) objc.c.id {
    _ = target;
    _ = sel;
    _ = table;

    const app_state = state.g_state orelse return null;
    if (row < 0) return null;

    const row_index: usize = @intCast(row);
    if (row_index >= app_state.model.filtered.items.len) return null;

    if (column != null) {
        const column_obj = objc.Object.fromId(column);
        if (columnIsIndex(column_obj)) {
            const query = logic.currentQuery(app_state);
            if (app_state.config.numericSelectionEnabled(app_state.model.filtered.items.len, query) and row_index < state.digit_labels.len) {
                return nsString(state.digit_labels[row_index]).value;
            }
            return nsString("").value;
        }
    }
    if (app_state.config.show_icons and column != null) {
        const column_obj = objc.Object.fromId(column);
        if (columnIsIcon(column_obj)) {
            const item_index = app_state.model.filtered.items[row_index];
            const image = iconImage(app_state.model.items[item_index].icon) orelse return null;
            return image.value;
        }
    }
    const item_index = app_state.model.filtered.items[row_index];
    const item = app_state.model.items[item_index];
    return nsString(item.label).value;
}

pub fn tableViewShouldSelectRow(
    target: objc.c.id,
    sel: objc.c.SEL,
    table: objc.c.id,
    row: c_long,
) callconv(.c) bool {
    _ = target;
    _ = sel;
    _ = table;

    const app_state = state.g_state orelse return true;
    if (row < 0) {
        app_state.model.selected = -1;
        return true;
    }

    const row_index: usize = @intCast(row);
    if (row_index >= app_state.model.filtered.items.len) {
        app_state.model.selected = -1;
        return true;
    }

    app_state.model.selected = @intCast(row);
    return true;
}

pub fn tableViewSelectionDidChange(target: objc.c.id, sel: objc.c.SEL, notification: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;

    const app_state = state.g_state orelse return;
    if (notification == null) return;

    const notification_obj = objc.Object.fromId(notification);
    const table = notification_obj.msgSend(objc.Object, "object", .{});
    const selected_row = table.msgSend(c_long, "selectedRow", .{});
    if (selected_row < 0) {
        app_state.model.selected = -1;
        return;
    }

    const row_index: usize = @intCast(selected_row);
    if (row_index >= app_state.model.filtered.items.len) {
        app_state.model.selected = -1;
        return;
    }

    app_state.model.selected = @intCast(selected_row);
}

pub fn cancelOperation(target: objc.c.id, sel: objc.c.SEL, sender: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = sender;
    if (state.g_state) |app_state| {
        logic.quit(app_state, exit_codes.user_canceled);
    }
    std.process.exit(exit_codes.user_canceled);
}

pub fn onFocusLossTimer(target: objc.c.id, sel: objc.c.SEL, timer: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = timer;
    if (state.g_state) |app_state| {
        logic.quit(app_state, exit_codes.user_canceled);
    }
    std.process.exit(exit_codes.user_canceled);
}

pub fn onUpdateTimer(target: objc.c.id, sel: objc.c.SEL, timer: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = timer;

    const app_state = state.g_state orelse return;
    const queue = app_state.update_queue orelse return;
    const updates_slice = queue.drain();
    if (updates_slice.len == 0) return;

    var latest_set_batch: ?u64 = null;
    for (updates_slice) |update| {
        if (update.kind != .set) continue;
        if (latest_set_batch == null or update.batch > latest_set_batch.?) {
            latest_set_batch = update.batch;
        }
    }

    var set_items = std.ArrayList(menu.MenuItem).empty;
    defer set_items.deinit(app_state.allocator);
    var prepend_items = std.ArrayList(menu.MenuItem).empty;
    defer prepend_items.deinit(app_state.allocator);
    var append_items = std.ArrayList(menu.MenuItem).empty;
    defer append_items.deinit(app_state.allocator);

    for (updates_slice) |update| {
        if (update.kind == .set and latest_set_batch != null and update.batch != latest_set_batch.?) {
            queue.allocator.free(update.line);
            continue;
        }
        const item = switch (update.source) {
            .stdin => menu.parseItem(app_state.allocator, update.line, 0, app_state.config.show_icons) catch {
                queue.allocator.free(update.line);
                continue;
            },
            .ipc => updates.menuItemFromIpc(app_state.allocator, update.line) orelse {
                queue.allocator.free(update.line);
                continue;
            },
        };
        queue.allocator.free(update.line);
        switch (update.kind) {
            .set => set_items.append(app_state.allocator, item) catch {},
            .prepend => prepend_items.append(app_state.allocator, item) catch {},
            .append => append_items.append(app_state.allocator, item) catch {},
        }
    }
    queue.allocator.free(updates_slice);

    if (set_items.items.len == 0 and prepend_items.items.len == 0 and append_items.items.len == 0) return;
    if (set_items.items.len > 0) {
        app_state.model.setItems(app_state.allocator, set_items.items) catch return;
    }
    if (prepend_items.items.len > 0) {
        app_state.model.prependItems(app_state.allocator, prepend_items.items) catch return;
    }
    if (append_items.items.len > 0) {
        app_state.model.appendItems(app_state.allocator, append_items.items) catch return;
    }
    logic.applyFilter(app_state, logic.currentQuery(app_state));
}

fn scheduleFocusLossCancel() void {
    const app_state = state.g_state orelse return;
    const NSTimer = objc.getClass("NSTimer").?;
    _ = NSTimer.msgSend(objc.Object, "scheduledTimerWithTimeInterval:target:selector:userInfo:repeats:", .{
        @as(f64, 0.04),
        app_state.handler,
        objc.sel("onFocusLossTimer:"),
        @as(objc.c.id, null),
        false,
    });
}

pub fn resignKeyWindow(target: objc.c.id, sel: objc.c.SEL) callconv(.c) void {
    _ = sel;
    if (target == null) return;

    const obj = objc.Object.fromId(target);
    const NSWindow = objc.getClass("NSWindow").?;
    obj.msgSendSuper(NSWindow, void, "resignKeyWindow", .{});
    if (state.g_state) |app_state| {
        if (!app_state.had_focus) return;
    }
    scheduleFocusLossCancel();
}

pub fn keyDown(target: objc.c.id, sel: objc.c.SEL, event: objc.c.id) callconv(.c) void {
    _ = sel;
    if (target == null) return;
    if (event == null) return;

    const obj = objc.Object.fromId(target);

    if (state.g_state) |app_state| {
        if (eventChar(event)) |ec| {
            if ((ec.modifiers & NSEventModifierFlagCommand) != 0 and (ec.char == 'a' or ec.char == 'A')) {
                app_state.text_field.msgSend(void, "selectText:", .{@as(objc.c.id, null)});
                return;
            }

            if ((ec.modifiers & NSEventModifierFlagControl) != 0 and (ec.char == 'l' or ec.char == 'L')) {
                app_state.text_field.msgSend(void, "setStringValue:", .{nsString("")});
                logic.applyFilter(app_state, "");
                return;
            }

            if (handleNumericShortcut(app_state, ec.char)) {
                return;
            }
        }
    }

    const NSTextField = objc.getClass("NSTextField").?;
    obj.msgSendSuper(NSTextField, void, "keyDown:", .{event});
}

pub fn performKeyEquivalent(target: objc.c.id, sel: objc.c.SEL, event: objc.c.id) callconv(.c) bool {
    _ = sel;
    if (target == null) return false;
    if (event == null) return false;

    const obj = objc.Object.fromId(target);
    if (eventChar(event)) |ec| {
        if ((ec.modifiers & NSEventModifierFlagCommand) != 0 and (ec.char == 'a' or ec.char == 'A')) {
            obj.msgSend(void, "selectText:", .{@as(objc.c.id, null)});
            return true;
        }
    }

    const NSTextField = objc.getClass("NSTextField").?;
    return obj.msgSendSuper(NSTextField, bool, "performKeyEquivalent:", .{event});
}

const EventChar = struct {
    char: u8,
    modifiers: c_ulong,
};

/// extract single-char + modifier flags from an NSEvent id
fn eventChar(event_id: objc.c.id) ?EventChar {
    if (event_id == null) return null;
    const event_obj = objc.Object.fromId(event_id);
    const chars = event_obj.msgSend(objc.Object, "charactersIgnoringModifiers", .{});
    const utf8_ptr = chars.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return null;
    const text = std.mem.sliceTo(utf8_ptr.?, 0);
    if (text.len != 1) return null;

    const modifiers = event_obj.msgSend(c_ulong, "modifierFlags", .{});
    const masked = modifiers & (NSEventModifierFlagShift | NSEventModifierFlagControl | NSEventModifierFlagOption | NSEventModifierFlagCommand);
    return .{ .char = text[0], .modifiers = masked };
}

fn currentEventChar() ?EventChar {
    const NSApplication = objc.getClass("NSApplication").?;
    const app = NSApplication.msgSend(objc.Object, "sharedApplication", .{});
    return eventChar(app.msgSend(objc.c.id, "currentEvent", .{}));
}

fn handleNumericShortcut(app_state: *state.AppState, ch: u8) bool {
    const query = logic.currentQuery(app_state);
    return handleNumericShortcutWithQuery(app_state, ch, query);
}

fn handleNumericShortcutWithQuery(app_state: *state.AppState, ch: u8, query_for_mode: []const u8) bool {
    if (!app_state.config.numericSelectionEnabled(app_state.model.filtered.items.len, query_for_mode)) return false;
    if (ch < '1' or ch > '9') return false;

    const index: usize = @intCast(ch - '1');
    if (index < app_state.model.filtered.items.len) {
        app_state.model.selected = @intCast(index);
        logic.updateSelection(app_state);
        logic.acceptSelection(app_state);
    }
    return true;
}

pub fn becomeFirstResponder(target: objc.c.id, sel: objc.c.SEL) callconv(.c) bool {
    _ = sel;
    if (target == null) return false;

    const obj = objc.Object.fromId(target);
    const NSTextField = objc.getClass("NSTextField").?;
    const accepted = obj.msgSendSuper(NSTextField, bool, "becomeFirstResponder", .{});
    if (accepted) {
        if (state.g_state) |app_state| {
            app_state.had_focus = true;
        }
        obj.msgSend(void, "selectText:", .{@as(objc.c.id, null)});
    }
    return accepted;
}

pub fn windowCanBecomeKey(target: objc.c.id, sel: objc.c.SEL) callconv(.c) bool {
    _ = target;
    _ = sel;
    return true;
}

pub fn windowCanBecomeMain(target: objc.c.id, sel: objc.c.SEL) callconv(.c) bool {
    _ = target;
    _ = sel;
    return true;
}
