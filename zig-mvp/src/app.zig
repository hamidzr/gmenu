const std = @import("std");
const objc = @import("objc");
const appconfig = @import("config.zig");
const menu = @import("menu.zig");

const NSApplicationActivationPolicyRegular: i64 = 0;
const NSWindowStyleMaskBorderless: u64 = 0;
const NSBackingStoreBuffered: u64 = 2;

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

const AppState = struct {
    model: menu.Model,
    table_view: objc.Object,
    config: appconfig.Config,
};

var g_state: ?*AppState = null;

fn nsString(str: [:0]const u8) objc.Object {
    const NSString = objc.getClass("NSString").?;
    return NSString.msgSend(objc.Object, "stringWithUTF8String:", .{str});
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

fn applyFilter(state: *AppState, query: []const u8) void {
    state.model.applyFilter(query, state.config.search);
    state.table_view.msgSend(void, "reloadData", .{});
    updateSelection(state);
}

fn moveSelection(state: *AppState, delta: isize) void {
    state.model.moveSelection(delta);
    updateSelection(state);
}

fn readItems(allocator: std.mem.Allocator) ![]menu.MenuItem {
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

        const label = try allocator.dupeZ(u8, trimmed);
        try items.append(allocator, .{ .label = label, .index = items.items.len });
    }

    if (items.items.len == 0) return error.NoInput;

    return items.toOwnedSlice(allocator);
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

    if (command == objc.sel("moveUp:")) {
        moveSelection(state, -1);
        return true;
    }
    if (command == objc.sel("moveDown:")) {
        moveSelection(state, 1);
        return true;
    }
    if (command == objc.sel("insertTab:")) {
        moveSelection(state, 1);
        return true;
    }
    if (command == objc.sel("insertBacktab:")) {
        moveSelection(state, -1);
        return true;
    }

    return false;
}

fn onSubmit(target: objc.c.id, sel: objc.c.SEL, sender: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;
    _ = sender;

    const state = g_state orelse return;
    const item = state.model.selectedItem() orelse {
        std.process.exit(1);
    };

    std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{item.label}) catch {};
    std.process.exit(0);
}

fn numberOfRowsInTableView(target: objc.c.id, sel: objc.c.SEL, table: objc.c.id) callconv(.c) c_long {
    _ = target;
    _ = sel;
    _ = table;

    const state = g_state orelse return 0;
    return @intCast(state.model.filtered.items.len);
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
    _ = column;

    const state = g_state orelse return null;
    if (row < 0) return null;

    const row_index: usize = @intCast(row);
    if (row_index >= state.model.filtered.items.len) return null;

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
    std.process.exit(1);
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

    const items = readItems(allocator) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: stdin is empty\n", .{}) catch {};
        std.process.exit(1);
    };

    var pool = objc.AutoreleasePool.init();
    defer pool.deinit();

    const NSApplication = objc.getClass("NSApplication").?;
    const app = NSApplication.msgSend(objc.Object, "sharedApplication", .{});
    _ = app.msgSend(bool, "setActivationPolicy:", .{NSApplicationActivationPolicyRegular});

    const style: u64 = NSWindowStyleMaskBorderless;
    const window_width = config.window_width;
    const window_height = config.window_height;
    const field_height = config.field_height;
    const padding = config.padding;
    const list_width = window_width - (padding * 2.0);
    const list_height = window_height - field_height - (padding * 3.0);

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

    const field_rect = NSRect{
        .origin = .{ .x = padding, .y = window_height - padding - field_height },
        .size = .{ .width = list_width, .height = field_height },
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

    const NSTableColumn = objc.getClass("NSTableColumn").?;
    const table_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithIdentifier:", .{nsString("items")});
    table_column.msgSend(void, "setWidth:", .{list_width});
    table_view.msgSend(void, "addTableColumn:", .{table_column});

    const NSScrollView = objc.getClass("NSScrollView").?;
    const scroll_view = NSScrollView.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{list_rect});

    scroll_view.msgSend(void, "setDocumentView:", .{table_view});
    scroll_view.msgSend(void, "setHasVerticalScroller:", .{true});

    content_view.msgSend(void, "addSubview:", .{scroll_view});
    content_view.msgSend(void, "addSubview:", .{text_field});

    var state = AppState{
        .model = try menu.Model.init(allocator, items),
        .table_view = table_view,
        .config = config,
    };
    defer state.model.deinit(allocator);
    g_state = &state;

    const data_source = makeDataSource();
    table_view.msgSend(void, "setDataSource:", .{data_source});
    table_view.msgSend(void, "setDelegate:", .{data_source});

    applyFilter(&state, "");

    app.msgSend(void, "activateIgnoringOtherApps:", .{true});
    window.msgSend(void, "makeKeyAndOrderFront:", .{@as(objc.c.id, null)});
    window.msgSend(void, "makeFirstResponder:", .{text_field});
    app.msgSend(void, "run", .{});
}
