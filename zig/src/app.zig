const std = @import("std");
const objc = @import("objc");
const appconfig = @import("config.zig");
const cache = @import("cache.zig");
const classes = @import("app/classes.zig");
const logic = @import("app/logic.zig");
const objc_helpers = @import("app/objc_helpers.zig");
const state = @import("app/state.zig");
const updates = @import("app/updates.zig");
const menu = @import("menu.zig");
const pid = @import("pid.zig");
const exit_codes = @import("exit_codes.zig");

const NSPoint = objc_helpers.NSPoint;
const NSSize = objc_helpers.NSSize;
const NSRect = objc_helpers.NSRect;
const nsString = objc_helpers.nsString;
const nsColor = objc_helpers.nsColor;
const nsFont = objc_helpers.nsFont;
const applyPlaceholderColor = objc_helpers.applyPlaceholderColor;
const applyColumnFont = objc_helpers.applyColumnFont;

const startUpdateQueue = updates.startUpdateQueue;
const followStdinThread = updates.followStdinThread;

const NSApplicationActivationPolicyRegular: i64 = 0;
const NSWindowStyleMaskBorderless: u64 = 0;
const NSBackingStoreBuffered: u64 = 2;

pub fn run(config: appconfig.Config) !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    var items: []menu.MenuItem = &[_]menu.MenuItem{};
    if (config.ipc_only) {
        items = allocator.alloc(menu.MenuItem, 0) catch {
            std.fs.File.stderr().deprecatedWriter().print("zmenu: unable to allocate items\n", .{}) catch {};
            std.process.exit(exit_codes.unknown_error);
        };
    } else if (!config.follow_stdin) {
        items = logic.readItems(allocator, config.show_icons) catch {
            std.fs.File.stderr().deprecatedWriter().print("zmenu: stdin is empty\n", .{}) catch {};
            std.process.exit(exit_codes.unknown_error);
        };
    } else {
        items = allocator.alloc(menu.MenuItem, 0) catch {
            std.fs.File.stderr().deprecatedWriter().print("zmenu: unable to allocate items\n", .{}) catch {};
            std.process.exit(exit_codes.unknown_error);
        };
    }

    const pid_path = pid.create(allocator, config.menu_id) catch {
        std.fs.File.stderr().deprecatedWriter().print("zmenu: another instance is running\n", .{}) catch {};
        std.process.exit(exit_codes.unknown_error);
    };
    defer pid.remove(pid_path);

    var initial_query: []const u8 = config.initial_query;
    if (initial_query.len == 0 and config.menu_id.len > 0) {
        if (cache.load(allocator, config.menu_id)) |cached| {
            if (cached) |cached_state| {
                if (cached_state.last_query.len > 0) {
                    initial_query = cached_state.last_query;
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
    const icon_width = if (config.show_icons) config.icon_column_width else 0;
    var item_width = list_width - numeric_width - icon_width;
    if (item_width < 0) item_width = 0;

    const window_rect = NSRect{
        .origin = .{ .x = 0, .y = 0 },
        .size = .{ .width = window_width, .height = window_height },
    };

    const WindowClass = classes.windowClass();
    const window = WindowClass.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithContentRect:styleMask:backing:defer:", .{
        window_rect,
        style,
        NSBackingStoreBuffered,
        false,
    });

    window.msgSend(void, "center", .{});
    window.msgSend(void, "setTitle:", .{nsString(config.title)});
    if (config.background_color) |color| {
        window.msgSend(void, "setBackgroundColor:", .{nsColor(color)});
    }

    // add subtle border around window
    window.msgSend(void, "setHasShadow:", .{true});
    const content_view = window.msgSend(objc.Object, "contentView", .{});
    content_view.msgSend(void, "setWantsLayer:", .{true});
    const layer = content_view.msgSend(objc.Object, "layer", .{});
    const NSColor = objc.getClass("NSColor").?;
    const border_color = NSColor.msgSend(objc.Object, "colorWithSRGBRed:green:blue:alpha:", .{
        @as(f64, 0.5),
        @as(f64, 0.5),
        @as(f64, 0.5),
        @as(f64, 0.3),
    });
    const cg_color = border_color.msgSend(objc.c.id, "CGColor", .{});
    layer.msgSend(void, "setBorderColor:", .{cg_color});
    layer.msgSend(void, "setBorderWidth:", .{@as(f64, 1.0)});
    layer.msgSend(void, "setCornerRadius:", .{@as(f64, 0.0)});

    var match_label_width: f64 = 100.0;
    var search_width: f64 = list_width - match_label_width;
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

    const font_size = @max(config.field_height * 0.65, 15.0);
    const text_font = nsFont(font_size);
    const text_color = if (config.text_color) |color| nsColor(color) else null;
    const secondary_text_color = if (config.secondary_text_color) |color| nsColor(color) else null;

    const SearchField = classes.searchFieldClass();
    const text_field = SearchField.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{field_rect});

    text_field.msgSend(void, "setEditable:", .{true});
    text_field.msgSend(void, "setSelectable:", .{true});
    text_field.msgSend(void, "setBezeled:", .{true});
    text_field.msgSend(void, "setBordered:", .{false});
    text_field.msgSend(void, "setFocusRingType:", .{@as(c_uint, 1)}); // NSFocusRingTypeNone = 1
    text_field.msgSend(void, "setAlignment:", .{@as(c_ulong, 0)}); // NSTextAlignmentLeft

    // Ensure proper text baseline alignment
    const cell = text_field.msgSend(objc.Object, "cell", .{});
    cell.msgSend(void, "setUsesSingleLineMode:", .{true});
    cell.msgSend(void, "setLineBreakMode:", .{@as(c_ulong, 2)}); // NSLineBreakByTruncatingTail

    if (text_color) |color| {
        text_field.msgSend(void, "setTextColor:", .{color});
    }
    text_field.msgSend(void, "setFont:", .{text_font});
    if (config.field_background_color) |color| {
        text_field.msgSend(void, "setDrawsBackground:", .{true});
        text_field.msgSend(void, "setBackgroundColor:", .{nsColor(color)});
    }

    // Set placeholder with custom color - must be done after setting bezeled/bordered
    if (secondary_text_color) |color| {
        applyPlaceholderColor(text_field, config.placeholder, color);
    } else {
        text_field.msgSend(void, "setPlaceholderString:", .{nsString(config.placeholder)});
    }

    const handler = classes.makeInputHandler();
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
    if (secondary_text_color) |color| {
        match_label.msgSend(void, "setTextColor:", .{color});
    }
    match_label.msgSend(void, "setFont:", .{text_font});

    const table_frame = NSRect{
        .origin = .{ .x = 0, .y = 0 },
        .size = .{ .width = list_width, .height = list_height },
    };

    const NSTableView = objc.getClass("NSTableView").?;
    const table_view = NSTableView.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{table_frame});

    const table_font = nsFont(@max(config.row_height * 0.6, 14.0));

    table_view.msgSend(void, "setHeaderView:", .{@as(objc.c.id, null)});
    table_view.msgSend(void, "setAllowsMultipleSelection:", .{false});
    table_view.msgSend(void, "setAllowsEmptySelection:", .{true});
    table_view.msgSend(void, "setRowHeight:", .{config.row_height});
    table_view.msgSend(void, "setUsesAlternatingRowBackgroundColors:", .{config.alternate_rows});
    table_view.msgSend(void, "setSelectionHighlightStyle:", .{@as(c_long, 1)}); // NSTableViewSelectionHighlightStyleRegular
    table_view.msgSend(void, "setTarget:", .{handler});
    table_view.msgSend(void, "setDoubleAction:", .{objc.sel("onSubmit:")});
    table_view.msgSend(void, "setIntercellSpacing:", .{NSSize{ .width = 0, .height = 0 }});
    table_view.msgSend(void, "setColumnAutoresizingStyle:", .{@as(c_ulong, 1)}); // NSTableViewLastColumnOnlyAutoresizingStyle

    const NSTableColumn = objc.getClass("NSTableColumn").?;
    if (!config.no_numeric_selection) {
        const index_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
            .msgSend(objc.Object, "initWithIdentifier:", .{nsString("index")});
        index_column.msgSend(void, "setWidth:", .{numeric_width});
        index_column.msgSend(void, "setMinWidth:", .{numeric_width});
        index_column.msgSend(void, "setMaxWidth:", .{numeric_width});
        index_column.msgSend(void, "setResizingMask:", .{@as(c_ulong, 0)}); // no resizing
        applyColumnFont(index_column, table_font, secondary_text_color);
        table_view.msgSend(void, "addTableColumn:", .{index_column});
    }
    if (config.show_icons) {
        const icon_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
            .msgSend(objc.Object, "initWithIdentifier:", .{nsString("icon")});
        icon_column.msgSend(void, "setWidth:", .{icon_width});
        icon_column.msgSend(void, "setMinWidth:", .{icon_width});
        icon_column.msgSend(void, "setMaxWidth:", .{icon_width});
        icon_column.msgSend(void, "setResizingMask:", .{@as(c_ulong, 0)}); // no resizing
        const NSImageCell = objc.getClass("NSImageCell").?;
        const image_cell = NSImageCell.msgSend(objc.Object, "alloc", .{})
            .msgSend(objc.Object, "init", .{});
        icon_column.msgSend(void, "setDataCell:", .{image_cell});
        table_view.msgSend(void, "addTableColumn:", .{icon_column});
    }
    const table_column = NSTableColumn.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithIdentifier:", .{nsString("items")});
    table_column.msgSend(void, "setWidth:", .{item_width});
    table_column.msgSend(void, "setResizingMask:", .{@as(c_ulong, 1)}); // autoresize with table
    applyColumnFont(table_column, table_font, text_color);
    table_view.msgSend(void, "addTableColumn:", .{table_column});

    const NSScrollView = objc.getClass("NSScrollView").?;
    const scroll_view = NSScrollView.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{list_rect});

    scroll_view.msgSend(void, "setDocumentView:", .{table_view});
    scroll_view.msgSend(void, "setHasVerticalScroller:", .{true});
    scroll_view.msgSend(void, "setAutohidesScrollers:", .{true});
    if (config.list_background_color) |color| {
        const list_color = nsColor(color);
        table_view.msgSend(void, "setBackgroundColor:", .{list_color});
        table_view.msgSend(void, "setGridStyleMask:", .{@as(c_ulong, 0)});
        table_view.msgSend(void, "setUsesAlternatingRowBackgroundColors:", .{false});
        scroll_view.msgSend(void, "setDrawsBackground:", .{true});
        scroll_view.msgSend(void, "setBackgroundColor:", .{list_color});
        scroll_view.msgSend(void, "setBorderType:", .{@as(c_ulong, 0)});
    }
    // Note: setSelectionHighlightColor: is not a valid NSTableView method
    // Custom selection colors require using setSelectionHighlightStyle: with NSTableViewSelectionHighlightStyleNone
    // and implementing custom cell rendering. For now, we use the default system selection color.
    _ = config.selection_color; // prevent unused variable warning

    content_view.msgSend(void, "addSubview:", .{scroll_view});
    content_view.msgSend(void, "addSubview:", .{text_field});
    if (match_label_width > 0) {
        content_view.msgSend(void, "addSubview:", .{match_label});
    }

    const queue = try startUpdateQueue(config);
    defer queue.deinit();

    var app_state = state.AppState{
        .model = try menu.Model.init(allocator, items),
        .table_view = table_view,
        .text_field = text_field,
        .match_label = match_label,
        .handler = handler,
        .config = config,
        .pid_path = pid_path,
        .allocator = allocator,
        .update_queue = queue.queue,
        .ipc_path = queue.ipc_path,
        .had_focus = false,
    };
    defer app_state.model.deinit(allocator);
    state.g_state = &app_state;

    if (queue.queue) |queue_ptr| {
        if (config.follow_stdin and !config.ipc_only) {
            _ = std.Thread.spawn(.{}, followStdinThread, .{queue_ptr}) catch {};
        }
        if (config.follow_stdin or config.ipc_only) {
            const NSTimer = objc.getClass("NSTimer").?;
            _ = NSTimer.msgSend(objc.Object, "scheduledTimerWithTimeInterval:target:selector:userInfo:repeats:", .{
                @as(f64, 0.2),
                handler,
                objc.sel("onUpdateTimer:"),
                @as(objc.c.id, null),
                true,
            });
        }
    }

    const data_source = classes.makeDataSource();
    table_view.msgSend(void, "setDataSource:", .{data_source});
    table_view.msgSend(void, "setDelegate:", .{data_source});

    if (initial_query.len > 0) {
        const initial_query_z = try allocator.dupeZ(u8, initial_query);
        text_field.msgSend(void, "setStringValue:", .{nsString(initial_query_z)});
        logic.applyFilter(&app_state, initial_query);
    } else {
        logic.applyFilter(&app_state, "");
    }

    app.msgSend(void, "activateIgnoringOtherApps:", .{true});
    window.msgSend(void, "makeKeyAndOrderFront:", .{@as(objc.c.id, null)});
    _ = window.msgSend(bool, "makeFirstResponder:", .{text_field});
    app.msgSend(void, "run", .{});
}
