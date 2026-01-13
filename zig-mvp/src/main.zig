const std = @import("std");
const objc = @import("objc");

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

fn nsString(str: [:0]const u8) objc.Object {
    const NSString = objc.getClass("NSString").?;
    return NSString.msgSend(objc.Object, "stringWithUTF8String:", .{str});
}

fn onSubmit(target: objc.c.id, sel: objc.c.SEL, sender: objc.c.id) callconv(.c) void {
    _ = target;
    _ = sel;

    const sender_obj = objc.Object.fromId(sender);
    const text = sender_obj.msgSend(objc.Object, "stringValue", .{});
    const utf8_ptr = text.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return;

    const slice = std.mem.sliceTo(utf8_ptr.?, 0);
    std.fs.File.stdout().deprecatedWriter().print("{s}\n", .{slice}) catch {};

    const NSApplication = objc.getClass("NSApplication").?;
    const app = NSApplication.msgSend(objc.Object, "sharedApplication", .{});
    app.msgSend(void, "terminate:", .{@as(objc.c.id, null)});
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

fn handlerClass() objc.Class {
    if (objc.getClass("ZigSubmitHandler")) |cls| return cls;

    const NSObject = objc.getClass("NSObject").?;
    const cls = objc.allocateClassPair(NSObject, "ZigSubmitHandler").?;
    if (!cls.addMethod("onSubmit:", onSubmit)) {
        @panic("failed to add onSubmit: method");
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

fn makeHandler() objc.Object {
    const cls = handlerClass();
    return cls.msgSend(objc.Object, "alloc", .{}).msgSend(objc.Object, "init", .{});
}

pub fn main() !void {
    var pool = objc.AutoreleasePool.init();
    defer pool.deinit();

    const NSApplication = objc.getClass("NSApplication").?;
    const app = NSApplication.msgSend(objc.Object, "sharedApplication", .{});
    _ = app.msgSend(bool, "setActivationPolicy:", .{NSApplicationActivationPolicyRegular});

    const style: u64 = NSWindowStyleMaskBorderless;

    const window_width: f64 = 520;
    const field_height: f64 = 24;
    const padding: f64 = 6;
    const window_height: f64 = field_height + (padding * 2.0);

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

    const content_view = window.msgSend(objc.Object, "contentView", .{});

    const field_rect = NSRect{
        .origin = .{ .x = padding, .y = padding },
        .size = .{ .width = window_width - (padding * 2.0), .height = field_height },
    };

    const NSTextField = objc.getClass("NSTextField").?;
    const text_field = NSTextField.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithFrame:", .{field_rect});

    text_field.msgSend(void, "setPlaceholderString:", .{nsString("Type and press Enter")});
    text_field.msgSend(void, "setEditable:", .{true});
    text_field.msgSend(void, "setSelectable:", .{true});

    const handler = makeHandler();
    text_field.msgSend(void, "setTarget:", .{handler});
    text_field.msgSend(void, "setAction:", .{objc.sel("onSubmit:")});

    content_view.msgSend(void, "addSubview:", .{text_field});
    app.msgSend(void, "activateIgnoringOtherApps:", .{true});
    window.msgSend(void, "makeKeyAndOrderFront:", .{@as(objc.c.id, null)});
    window.msgSend(void, "makeFirstResponder:", .{text_field});
    app.msgSend(void, "run", .{});
}
