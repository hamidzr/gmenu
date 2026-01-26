const std = @import("std");
const objc = @import("objc");
const appconfig = @import("../config.zig");
const menu = @import("../menu.zig");

pub const NSPoint = extern struct {
    x: f64,
    y: f64,
};

pub const NSSize = extern struct {
    width: f64,
    height: f64,
};

pub const NSRect = extern struct {
    origin: NSPoint,
    size: NSSize,
};

pub fn nsString(str: [:0]const u8) objc.Object {
    const NSString = objc.getClass("NSString").?;
    return NSString.msgSend(objc.Object, "stringWithUTF8String:", .{str});
}

pub fn nsColor(color: appconfig.Color) objc.Object {
    const NSColor = objc.getClass("NSColor").?;
    return NSColor.msgSend(objc.Object, "colorWithSRGBRed:green:blue:alpha:", .{
        color.r,
        color.g,
        color.b,
        color.a,
    });
}

pub fn nsFont(size: f64) objc.Object {
    const NSFont = objc.getClass("NSFont").?;
    return NSFont.msgSend(objc.Object, "systemFontOfSize:", .{size});
}

pub fn applyPlaceholderColor(field: objc.Object, placeholder: [:0]const u8, color: objc.Object) void {
    const NSDictionary = objc.getClass("NSDictionary").?;
    const NSAttributedString = objc.getClass("NSAttributedString").?;
    const key = nsString("NSForegroundColorAttributeName");
    const value = NSDictionary.msgSend(objc.Object, "dictionaryWithObject:forKey:", .{ color, key });
    const attributed = NSAttributedString.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "initWithString:attributes:", .{ nsString(placeholder), value });
    field.msgSend(void, "setPlaceholderAttributedString:", .{attributed});
}

pub fn applyColumnFont(column: objc.Object, font: objc.Object, text_color: ?objc.Object) void {
    const NSTextFieldCell = objc.getClass("NSTextFieldCell").?;
    const cell = NSTextFieldCell.msgSend(objc.Object, "alloc", .{})
        .msgSend(objc.Object, "init", .{});
    cell.msgSend(void, "setFont:", .{font});
    if (text_color) |color| {
        cell.msgSend(void, "setTextColor:", .{color});
    }
    column.msgSend(void, "setDataCell:", .{cell});
}

pub fn columnIsIndex(column: objc.Object) bool {
    const identifier = column.msgSend(objc.Object, "identifier", .{});
    const utf8_ptr = identifier.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return false;
    const name = std.mem.sliceTo(utf8_ptr.?, 0);
    return std.mem.eql(u8, name, "index");
}

pub fn columnIsIcon(column: objc.Object) bool {
    const identifier = column.msgSend(objc.Object, "identifier", .{});
    const utf8_ptr = identifier.msgSend(?[*:0]const u8, "UTF8String", .{});
    if (utf8_ptr == null) return false;
    const name = std.mem.sliceTo(utf8_ptr.?, 0);
    return std.mem.eql(u8, name, "icon");
}

pub fn iconImage(kind: menu.IconKind) ?objc.Object {
    const name: [:0]const u8 = switch (kind) {
        .app => "NSApplicationIcon",
        .file => "NSGenericDocument",
        .folder => "NSFolder",
        .info => "NSInfo",
        else => return null,
    };
    const NSImage = objc.getClass("NSImage").?;
    const image = NSImage.msgSend(objc.Object, "imageNamed:", .{nsString(name)});
    if (image.value == null) return null;
    return image;
}
