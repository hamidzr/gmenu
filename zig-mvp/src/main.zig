const std = @import("std");
const objc = @import("objc");

const NSApplicationActivationPolicyRegular: i64 = 0;

const NSWindowStyleMaskTitled: u64 = 1 << 0;
const NSWindowStyleMaskClosable: u64 = 1 << 1;
const NSWindowStyleMaskMiniaturizable: u64 = 1 << 2;

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

    const style: u64 = NSWindowStyleMaskTitled |
        NSWindowStyleMaskClosable |
        NSWindowStyleMaskMiniaturizable;

    const window_rect = NSRect{
        .origin = .{ .x = 0, .y = 0 },
        .size = .{ .width = 520, .height = 140 },
    };

    const NSWindow = objc.getClass("NSWindow").?;
    const window = NSWindow.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithContentRect:styleMask:backing:defer:", .{
        window_rect,
        style,
        NSBackingStoreBuffered,
        false,
    });

    window.msgSend(void, "center", .{});
    window.msgSend(void, "setTitle:", .{nsString("gmenu zig-mvp")});

    const content_view = window.msgSend(objc.Object, "contentView", .{});

    const field_rect = NSRect{
        .origin = .{ .x = 20, .y = 60 },
        .size = .{ .width = 480, .height = 24 },
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
    window.msgSend(void, "makeKeyAndOrderFront:", .{@as(objc.c.id, null)});
    window.msgSend(void, "makeFirstResponder:", .{text_field});

    app.msgSend(void, "activateIgnoringOtherApps:", .{true});
    app.msgSend(void, "run", .{});
}
