const std = @import("std");
const objc = @import("objc");
const appconfig = @import("config.zig");
const menu = @import("menu.zig");
const cache = @import("cache.zig");
const pid = @import("pid.zig");
const exit_codes = @import("exit_codes.zig");

const NSApplicationActivationPolicyRegular: i64 = 0;
const NSWindowStyleMaskBorderless: u64 = 0;
const NSBackingStoreBuffered: u64 = 2;
const NSEventModifierFlagControl: u64 = 1 << 18;

const NSPoint = extern struct {
    x: f64,
    y: f64,
};

const NSSize = extern struct {
    width: f64,
    height: f64,
};

const NSRect = extern struct {
    origin: NSPoint,
    size: NSSize,
};

const UpdateQueue = struct {
    allocator: std.mem.Allocator,
    mutex: std.Thread.Mutex = .{},
    items: std.ArrayList([]const u8),

    pub fn init(allocator: std.mem.Allocator) UpdateQueue {
        return .{
            .allocator = allocator,
            .items = std.ArrayList([]const u8).empty,
        };
    }

    pub fn pushOwned(self: *UpdateQueue, line: []const u8) void {
        self.mutex.lock();
        defer self.mutex.unlock();
        self.items.append(self.allocator, line) catch self.allocator.free(line);
    }

    pub fn drain(self: *UpdateQueue) []const []const u8 {
        self.mutex.lock();
        defer self.mutex.unlock();
        if (self.items.items.len == 0) {
            return &[_][]const u8{};
        }
        const out = self.allocator.alloc([]const u8, self.items.items.len) catch {
            self.items.clearRetainingCapacity();
            return &[_][]const u8{};
        };
        @memcpy(out, self.items.items);
        self.items.clearRetainingCapacity();
        return out;
    }
};

const AppState = struct {
    model: menu.Model,
    table_view: objc.Object,
    text_field: objc.Object,
    match_label: objc.Object,
    handler: objc.Object,
    config: appconfig.Config,
    pid_path: ?[]const u8,
    allocator: std.mem.Allocator,
    update_queue: ?*UpdateQueue,
};

const digit_labels = [_][:0]const u8{ "1", "2", "3", "4", "5", "6", "7", "8", "9" };

var g_state: ?*AppState = null;

fn nsString(str: [:0]const u8) objc.Object {
    const NSString = objc.getClass("NSString").?;
    return NSString.msgSend(objc.Object, "stringWithUTF8String:", .{str});
}

fn quit(state: *AppState, code: u8) void {
    if (state.pid_path) |path| {
        pid.remove(path);
    }
    std.process.exit(code);
}

fn updateSelection(state: *AppState) void {
    const row = state.model.selectedRow() orelse {
        state.table_view.msgSend(void, "deselectAll:", .{@as(objc.c.id, null)});
        return;
    };

    const NSIndexSet = objc.getClass("NSIndexSet").?;
    const index_set = NSIndexSet.msgSend(objc.Object, "indexSetWithIndex:", .{@as(c_ulong, @intCast(row))});
    state.table_view.msgSend(void, "selectRowIndexes:byExtendingSelection:", .{ index_set, false });
    state.table_view.msgSend(void, "scrollRowToVisible:", .{@as(c_long, @intCast(row))});
}

fn updateMatchLabel(state: *AppState) void {
    var buf: [32]u8 = undefined;
    const label_z = std.fmt.bufPrintZ(
        &buf,
        "[{d}/{d}]",
        .{ state.model.match_count, state.model.items.len },
    ) catch return;
    state.match_label.msgSend(void, "setStringValue:", .{nsString(label_z)});
}

fn applyFilter(state: *AppState, query: []const u8) void {
    state.model.applyFilter(query, state.config.search);
    state.table_view.msgSend(void, "reloadData", .{});
    updateSelection(state);
    updateMatchLabel(state);
    if (state.config.auto_accept and state.model.filtered.items.len == 1) {
        acceptSelection(state);
    }
}

fn moveSelection(state: *AppState, delta: isize) void {
    state.model.moveSelection(delta);
    updateSelection(state);
}

fn readItems(allocator: std.mem.Allocator, parse_icons: bool) ![]menu.MenuItem {
    const stdin = std.fs.File.stdin();
    const input = try stdin.readToEndAlloc(allocator, 16 * 1024 * 1024);
    if (input.len == 0) return error.NoInput;

    var items = std.ArrayList(menu.MenuItem).empty;
    errdefer items.deinit(allocator);

    var iter = std.mem.splitScalar(u8, input, '\n');
    while (iter.next()) |line| {
        var trimmed = line;
        if (trimmed.len > 0 and trimmed[trimmed.len - 1] == '\r') {
            trimmed = trimmed[0 .. trimmed.len - 1];
        }
        if (trimmed.len == 0) continue;

        const item = try menu.parseItem(allocator, trimmed, items.items.len, parse_icons);
        try items.append(allocator, item);
    }

    if (items.items.len == 0) return error.NoInput;

    return items.toOwnedSlice(allocator);
}

fn followStdinThread(queue: *UpdateQueue) void {
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
        queue.pushOwned(line);
    }
}

fn controlTextDidChange(target: objc.c.id, sel: objc.c.SEL, notification: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;

    const state = g_state orelse return;
    if (notification == null) return;

    const notification_obj = objc.Object.fromId(notification);
    const field = notification_obj.msgSend(objc.Object, "object", .{});
    const text = field.msgSend(objc.Object, "stringValue", .{});
    const utf8_ptr = text.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) {
        applyFilter(state, "");
        return;
    }

    const query = std.mem.sliceTo(utf8_ptr.?, 0);
    applyFilter(state, query);
}

fn controlTextViewDoCommandBySelector(
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

    const state = g_state orelse return false;

    if (command == objc.sel("moveUp:").value) {
        moveSelection(state, -1);
        return true;
    }
    if (command == objc.sel("moveDown:").value) {
        moveSelection(state, 1);
        return true;
    }
    if (command == objc.sel("insertTab:").value) {
        moveSelection(state, 1);
        return true;
    }
    if (command == objc.sel("insertBacktab:").value) {
        moveSelection(state, -1);
        return true;
    }

    return false;
}

fn currentQuery(state: *AppState) []const u8 {
    const text = state.text_field.msgSend(objc.Object, "stringValue", .{});
    const utf8_ptr = text.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return "";
    return std.mem.sliceTo(utf8_ptr.?, 0);
}

fn saveCache(state: *AppState, selection: []const u8) void {
    if (state.config.menu_id.len == 0) return;
    const query = currentQuery(state);
    cache.save(state.allocator, state.config.menu_id, .{
        .last_query = query,
        .last_selection = selection,
    }) catch {};
}

fn acceptSelection(state: *AppState) void {
    if (state.model.filtered.items.len == 0) {
        if (!state.config.accept_custom_selection) {
            return;
        }
        const query = currentQuery(state);
        saveCache(state, query);
        std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{query}) catch {};
        quit(state, 0);
    }

    if (state.model.selectedItem()) |item| {
        saveCache(state, item.label);
        std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{item.label}) catch {};
        quit(state, 0);
    }

    const item_index = state.model.filtered.items[0];
    const item = state.model.items[item_index];
    saveCache(state, item.label);
    std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{item.label}) catch {};
    quit(state, 0);
}

fn onSubmit(target: objc.c.id, sel: objc.c.SEL, sender: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = sender;

    const state = g_state orelse return;
    acceptSelection(state);
}

fn numberOfRowsInTableView(target: objc.c.id, sel: objc.c.SEL, table: objc.c.id) callconv(.c) c_long {
    _ = target;
    _ = sel;
    _ = table;

    const state = g_state orelse return 0;
    return @intCast(state.model.filtered.items.len);
}

fn columnIsIndex(column: objc.Object) bool {
    const identifier = column.msgSend(objc.Object, "identifier", .{});
    const utf8_ptr = identifier.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return false;
    const name = std.mem.sliceTo(utf8_ptr.?, 0);
    return std.mem.eql(u8, name, "index");
}

fn columnIsIcon(column: objc.Object) bool {
    const identifier = column.msgSend(objc.Object, "identifier", .{});
    const utf8_ptr = identifier.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return false;
    const name = std.mem.sliceTo(utf8_ptr.?, 0);
    return std.mem.eql(u8, name, "icon");
}

fn columnIsScore(column: objc.Object) bool {
    const identifier = column.msgSend(objc.Object, "identifier", .{});
    const utf8_ptr = identifier.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return false;
    const name = std.mem.sliceTo(utf8_ptr.?, 0);
    return std.mem.eql(u8, name, "score");
}

fn iconLabel(kind: menu.IconKind) [:0]const u8 {
    return switch (kind) {
        .app => "APP",
        .file => "FILE",
        .folder => "DIR",
        .info => "INFO",
        else => "",
    };
}

fn tableViewObjectValue(
    target: objc.c.id,
    sel: objc.c.SEL,
    table: objc.c.id,
    column: objc.c.id,
    row: c_long,
) callconv(.c) objc.c.id {
    _ = target;
    _ = sel;
    _ = table;

    const state = g_state orelse return null;
    if (row < 0) return null;

    const row_index: usize = @intCast(row);
    if (row_index >= state.model.filtered.items.len) return null;

    if (!state.config.no_numeric_selection and column != null) {
        const column_obj = objc.Object.fromId(column);
        if (columnIsIndex(column_obj)) {
            if (row_index < digit_labels.len) {
                return nsString(digit_labels[row_index]).value;
            }
            return nsString("").value;
        }
    }
    if (state.config.show_icons and column != null) {
        const column_obj = objc.Object.fromId(column);
        if (columnIsIcon(column_obj)) {
            const item_index = state.model.filtered.items[row_index];
            const label = iconLabel(state.model.items[item_index].icon);
            if (label.len == 0) return nsString("").value;
            return nsString(label).value;
        }
    }
    if (state.config.show_score and column != null) {
        const column_obj = objc.Object.fromId(column);
        if (columnIsScore(column_obj)) {
            const item_index = state.model.filtered.items[row_index];
            const score = state.model.scores[item_index];
            if (score <= 0) return nsString("").value;
            var buf: [32]u8 = undefined;
            const score_z = std.fmt.bufPrintZ(&buf, "{d}", .{score}) catch return nsString("").value;
            return nsString(score_z).value;
        }
    }

    const item_index = state.model.filtered.items[row_index];
    const item = state.model.items[item_index];
    return nsString(item.label).value;
}

fn tableViewSelectionDidChange(target: objc.c.id, sel: objc.c.SEL, notification: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;

    const state = g_state orelse return;
    if (notification == null) return;

    const notification_obj = objc.Object.fromId(notification);
    const table = notification_obj.msgSend(objc.Object, "object", .{});
    const selected_row = table.msgSend(c_long, "selectedRow", .{});
    if (selected_row < 0) {
        state.model.selected = -1;
        return;
    }

    const row_index: usize = @intCast(selected_row);
    if (row_index >= state.model.filtered.items.len) {
        state.model.selected = -1;
        return;
    }

    state.model.selected = @intCast(selected_row);
}

fn cancelOperation(target: objc.c.id, sel: objc.c.SEL, sender: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = sender;
    if (g_state) |state| {
        quit(state, exit_codes.user_canceled);
    }
    std.process.exit(exit_codes.user_canceled);
}

fn onFocusLossTimer(target: objc.c.id, sel: objc.c.SEL, timer: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = timer;
    if (g_state) |state| {
        quit(state, exit_codes.user_canceled);
    }
    std.process.exit(exit_codes.user_canceled);
}

fn onUpdateTimer(target: objc.c.id, sel: objc.c.SEL, timer: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = timer;

    const state = g_state orelse return;
    const queue = state.update_queue orelse return;
    const lines = queue.drain();
    if (lines.len == 0) return;

    var new_items = std.ArrayList(menu.MenuItem).empty;
    defer new_items.deinit(state.allocator);
    new_items.ensureTotalCapacity(state.allocator, lines.len) catch {
        for (lines) |line| queue.allocator.free(line);
        queue.allocator.free(lines);
        return;
    };

    const start = state.model.items.len;
    for (lines, 0..) |line, idx| {
        const item = menu.parseItem(state.allocator, line, start + idx, state.config.show_icons) catch {
            queue.allocator.free(line);
            continue;
        };
        new_items.appendAssumeCapacity(item);
        queue.allocator.free(line);
    }
    queue.allocator.free(lines);

    if (new_items.items.len == 0) return;
    state.model.appendItems(state.allocator, new_items.items) catch return;
    applyFilter(state, currentQuery(state));
}

fn scheduleFocusLossCancel() void {
    const state = g_state orelse return;
    const NSTimer = objc.getClass("NSTimer").?;
    _ = NSTimer.msgSend(objc.Object, "scheduledTimerWithTimeInterval:target:selector:userInfo:repeats:", .{
        0.04,
        state.handler,
        objc.sel("onFocusLossTimer:"),
        @as(objc.c.id, null),
        false,
    });
}

fn resignKeyWindow(target: objc.c.id, sel: objc.c.SEL) callconv(.c) void {
    _ = sel;
    if (target == null) return;

    const obj = objc.Object.fromId(target);
    const NSWindow = objc.getClass("NSWindow").?;
    obj.msgSendSuper(NSWindow, void, "resignKeyWindow", .{});
    scheduleFocusLossCancel();
}

fn keyDown(target: objc.c.id, sel: objc.c.SEL, event: objc.c.id) callconv(.c) void {
    _ = sel;
    if (target == null) return;
    if (event == null) return;

    const obj = objc.Object.fromId(target);
    const state = g_state;

    if (state != null) {
        const event_obj = objc.Object.fromId(event);
        const chars = event_obj.msgSend(objc.Object, "charactersIgnoringModifiers", .{});
        const utf8_ptr = chars.msgSend(?[*:0]const u8, "UTF8String", .{});
        if (utf8_ptr != null) {
            const text = std.mem.sliceTo(utf8_ptr.?, 0);
            if (text.len == 1) {
                const ch = text[0];
                const modifiers = event_obj.msgSend(c_ulong, "modifierFlags", .{});
                if ((modifiers & NSEventModifierFlagControl) != 0 and (ch == 'l' or ch == 'L')) {
                    state.?.text_field.msgSend(void, "setStringValue:", .{nsString("")});
                    applyFilter(state.?, "");
                    return;
                }

                if (!state.?.config.no_numeric_selection and ch >= '1' and ch <= '9') {
                    const index: usize = @intCast(ch - '1');
                    if (index < state.?.model.filtered.items.len) {
                        state.?.model.selected = @intCast(index);
                        updateSelection(state.?);
                        acceptSelection(state.?);
                    }
                    return;
                }
            }
        }
    }

    const NSTextField = objc.getClass("NSTextField").?;
    obj.msgSendSuper(NSTextField, void, "keyDown:", .{event});
}

fn windowCanBecomeKey(target: objc.c.id, sel: objc.c.SEL) callconv(.c) bool {
    _ = target;
    _ = sel;
    return true;
}

fn windowCanBecomeMain(target: objc.c.id, sel: objc.c.SEL) callconv(.c) bool {
    _ = target;
    _ = sel;
    return true;
}

fn inputHandlerClass() objc.Class {
    if (objc.getClass("ZigInputHandler")) |cls| return cls;

    const NSObject = objc.getClass("NSObject").?;
    const cls = objc.allocateClassPair(NSObject, "ZigInputHandler").?;
    if (!cls.addMethod("controlTextDidChange:", controlTextDidChange)) {
        @panic("failed to add controlTextDidChange: method");
    }
    if (!cls.addMethod("control:textView:doCommandBySelector:", controlTextViewDoCommandBySelector)) {
        @panic("failed to add control:textView:doCommandBySelector: method");
    }
    if (!cls.addMethod("onFocusLossTimer:", onFocusLossTimer)) {
        @panic("failed to add onFocusLossTimer: method");
    }
    if (!cls.addMethod("onUpdateTimer:", onUpdateTimer)) {
        @panic("failed to add onUpdateTimer: method");
    }
    if (!cls.addMethod("onSubmit:", onSubmit)) {
        @panic("failed to add onSubmit: method");
    }
    objc.registerClassPair(cls);
    return cls;
}

fn dataSourceClass() objc.Class {
    if (objc.getClass("ZigTableDataSource")) |cls| return cls;

    const NSObject = objc.getClass("NSObject").?;
    const cls = objc.allocateClassPair(NSObject, "ZigTableDataSource").?;
    if (!cls.addMethod("numberOfRowsInTableView:", numberOfRowsInTableView)) {
        @panic("failed to add numberOfRowsInTableView: method");
    }
    if (!cls.addMethod("tableView:objectValueForTableColumn:row:", tableViewObjectValue)) {
        @panic("failed to add tableView:objectValueForTableColumn:row: method");
    }
    if (!cls.addMethod("tableViewSelectionDidChange:", tableViewSelectionDidChange)) {
        @panic("failed to add tableViewSelectionDidChange: method");
    }
    objc.registerClassPair(cls);
    return cls;
}

fn searchFieldClass() objc.Class {
    if (objc.getClass("ZigSearchField")) |cls| return cls;

    const NSTextField = objc.getClass("NSTextField").?;
    const cls = objc.allocateClassPair(NSTextField, "ZigSearchField").?;
    if (!cls.addMethod("cancelOperation:", cancelOperation)) {
        @panic("failed to add cancelOperation: method");
    }
    if (!cls.addMethod("keyDown:", keyDown)) {
        @panic("failed to add keyDown: method");
    }
    objc.registerClassPair(cls);
    return cls;
}

fn windowClass() objc.Class {
    if (objc.getClass("ZigBorderlessWindow")) |cls| return cls;

    const NSWindow = objc.getClass("NSWindow").?;
    const cls = objc.allocateClassPair(NSWindow, "ZigBorderlessWindow").?;
    if (!cls.addMethod("canBecomeKeyWindow", windowCanBecomeKey)) {
        @panic("failed to add canBecomeKeyWindow method");
    }
    if (!cls.addMethod("canBecomeMainWindow", windowCanBecomeMain)) {
        @panic("failed to add canBecomeMainWindow method");
    }
    if (!cls.addMethod("resignKeyWindow", resignKeyWindow)) {
        @panic("failed to add resignKeyWindow method");
    }
    objc.registerClassPair(cls);
    return cls;
}

fn makeInputHandler() objc.Object {
    const cls = inputHandlerClass();
    return cls.msgSend(objc.Object, "alloc", .{}).msgSend(objc.Object, "init", .{});
}

fn makeDataSource() objc.Object {
    const cls = dataSourceClass();
    return cls.msgSend(objc.Object, "alloc", .{}).msgSend(objc.Object, "init", .{});
}

pub fn run(config: appconfig.Config) !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    var items: []menu.MenuItem = &[_]menu.MenuItem{};
    if (!config.follow_stdin) {
        items = readItems(allocator, config.show_icons) catch {
            std.fs.File.stderr().deprecatedWriter().print("zmenu: stdin is empty\n", .{}) catch {};
            std.process.exit(exit_codes.unknown_error);
        };
    }

    const pid_path = pid.create(allocator, config.menu_id) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: another instance is running\n", .{}) catch {};
        std.process.exit(exit_codes.unknown_error);
    };

    var initial_query: []const u8 = config.initial_query;
    if (initial_query.len == 0 and config.menu_id.len > 0) {
        if (cache.load(allocator, config.menu_id)) |cached| {
            if (cached) |state| {
                if (state.last_query.len > 0) {
                    initial_query = state.last_query;
                }
            }
        } else |_| {}
    }

    var pool = objc.AutoreleasePool.init();
    defer pool.deinit();

    const NSApplication = objc.getClass("NSApplication").?;
    const app = NSApplication.msgSend(objc.Object, "sharedApplication", .{});
    _ = app.msgSend(bool, "setActivationPolicy:", .{NSApplicationActivationPolicyRegular});

    const style: u64 = NSWindowStyleMaskBorderless;
    var window_width = config.window_width;
    var window_height = config.window_height;
    if (config.max_width > 0 and window_width > config.max_width) {
        window_width = config.max_width;
    }
    if (config.max_height > 0 and window_height > config.max_height) {
        window_height = config.max_height;
    }
    const field_height = config.field_height;
    const padding = config.padding;
    const list_width = window_width - (padding * 2.0);
    const list_height = window_height - field_height - (padding * 3.0);
    const numeric_width = if (config.no_numeric_selection) 0 else config.numeric_column_width;
    const score_width = if (config.show_score) config.score_column_width else 0;
    const icon_width = if (config.show_icons) config.icon_column_width else 0;
    var item_width = list_width - numeric_width - icon_width - score_width;
    if (item_width < 0) item_width = 0;

    const window_rect = NSRect{
        .origin = .{ .x = 0, .y = 0 },
        .size = .{ .width = window_width, .height = window_height },
    };

    const WindowClass = windowClass();
    const window = WindowClass.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithContentRect:styleMask:backing:defer:", .{
        window_rect,
        style,
        NSBackingStoreBuffered,
        false,
    });

    window.msgSend(void, "center", .{});
    window.msgSend(void, "setTitle:", .{nsString(config.title)});

    const content_view = window.msgSend(objc.Object, "contentView", .{});

    var match_label_width = 100.0;
    var search_width = list_width - match_label_width;
    if (search_width < 0) {
        search_width = list_width;
        match_label_width = 0;
    }

    const field_rect = NSRect{
        .origin = .{ .x = padding, .y = window_height - padding - field_height },
        .size = .{ .width = search_width, .height = field_height },
    };

    const match_rect = NSRect{
        .origin = .{ .x = padding + search_width, .y = window_height - padding - field_height },
        .size = .{ .width = match_label_width, .height = field_height },
    };

    const list_rect = NSRect{
        .origin = .{ .x = padding, .y = padding },
        .size = .{ .width = list_width, .height = list_height },
    };

    const SearchField = searchFieldClass();
    const text_field = SearchField.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{field_rect});

    text_field.msgSend(void, "setPlaceholderString:", .{nsString(config.placeholder)});
    text_field.msgSend(void, "setEditable:", .{true});
    text_field.msgSend(void, "setSelectable:", .{true});

    const handler = makeInputHandler();
    text_field.msgSend(void, "setDelegate:", .{handler});
    text_field.msgSend(void, "setTarget:", .{handler});
    text_field.msgSend(void, "setAction:", .{objc.sel("onSubmit:")});

    const NSTextField = objc.getClass("NSTextField").?;
    const match_label = NSTextField.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{match_rect});
    match_label.msgSend(void, "setBezeled:", .{false});
    match_label.msgSend(void, "setDrawsBackground:", .{false});
    match_label.msgSend(void, "setEditable:", .{false});
    match_label.msgSend(void, "setSelectable:", .{false});
    match_label.msgSend(void, "setAlignment:", .{@as(c_int, 2)});

    const table_frame = NSRect{
        .origin = .{ .x = 0, .y = 0 },
        .size = .{ .width = list_width, .height = list_height },
    };

    const NSTableView = objc.getClass("NSTableView").?;
    const table_view = NSTableView.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{table_frame});

    table_view.msgSend(void, "setHeaderView:", .{@as(objc.c.id, null)});
    table_view.msgSend(void, "setAllowsMultipleSelection:", .{false});
    table_view.msgSend(void, "setAllowsEmptySelection:", .{true});
    table_view.msgSend(void, "setRowHeight:", .{config.row_height});
    table_view.msgSend(void, "setUsesAlternatingRowBackgroundColors:", .{config.alternate_rows});
    table_view.msgSend(void, "setTarget:", .{handler});
    table_view.msgSend(void, "setDoubleAction:", .{objc.sel("onSubmit:")});

    const NSTableColumn = objc.getClass("NSTableColumn").?;
    if (!config.no_numeric_selection) {
        const index_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
            .msgSend(objc.Object, "initWithIdentifier:", .{nsString("index")});
        index_column.msgSend(void, "setWidth:", .{numeric_width});
        table_view.msgSend(void, "addTableColumn:", .{index_column});
    }
    if (config.show_icons) {
        const icon_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
            .msgSend(objc.Object, "initWithIdentifier:", .{nsString("icon")});
        icon_column.msgSend(void, "setWidth:", .{icon_width});
        table_view.msgSend(void, "addTableColumn:", .{icon_column});
    }
    const table_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithIdentifier:", .{nsString("items")});
    table_column.msgSend(void, "setWidth:", .{item_width});
    table_view.msgSend(void, "addTableColumn:", .{table_column});
    if (config.show_score) {
        const score_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
            .msgSend(objc.Object, "initWithIdentifier:", .{nsString("score")});
        score_column.msgSend(void, "setWidth:", .{score_width});
        table_view.msgSend(void, "addTableColumn:", .{score_column});
    }

    const NSScrollView = objc.getClass("NSScrollView").?;
    const scroll_view = NSScrollView.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{list_rect});

    scroll_view.msgSend(void, "setDocumentView:", .{table_view});
    scroll_view.msgSend(void, "setHasVerticalScroller:", .{true});

    content_view.msgSend(void, "addSubview:", .{scroll_view});
    content_view.msgSend(void, "addSubview:", .{text_field});
    if (match_label_width > 0) {
        content_view.msgSend(void, "addSubview:", .{match_label});
    }

    var update_queue: UpdateQueue = undefined;
    var update_queue_ptr: ?*UpdateQueue = null;
    if (config.follow_stdin) {
        update_queue = UpdateQueue.init(std.heap.c_allocator);
        update_queue_ptr = &update_queue;
    }

    var state = AppState{
        .model = try menu.Model.init(allocator, items),
        .table_view = table_view,
        .text_field = text_field,
        .match_label = match_label,
        .handler = handler,
        .config = config,
        .pid_path = pid_path,
        .allocator = allocator,
        .update_queue = update_queue_ptr,
    };
    defer state.model.deinit(allocator);
    g_state = &state;

    if (config.follow_stdin and update_queue_ptr != null) {
        _ = std.Thread.spawn(.{}, followStdinThread, .{update_queue_ptr.?}) catch {};
        const NSTimer = objc.getClass("NSTimer").?;
        _ = NSTimer.msgSend(objc.Object, "scheduledTimerWithTimeInterval:target:selector:userInfo:repeats:", .{
            0.2,
            handler,
            objc.sel("onUpdateTimer:"),
            @as(objc.c.id, null),
            true,
        });
    }

    const data_source = makeDataSource();
    table_view.msgSend(void, "setDataSource:", .{data_source});
    table_view.msgSend(void, "setDelegate:", .{data_source});

    if (initial_query.len > 0) {
        const initial_query_z = try allocator.dupeZ(u8, initial_query);
        text_field.msgSend(void, "setStringValue:", .{nsString(initial_query_z)});
        applyFilter(&state, initial_query);
    } else {
        applyFilter(&state, "");
    }

    app.msgSend(void, "activateIgnoringOtherApps:", .{true});
    window.msgSend(void, "makeKeyAndOrderFront:", .{@as(objc.c.id, null)});
    window.msgSend(void, "makeFirstResponder:", .{text_field});
    app.msgSend(void, "run", .{});
}
