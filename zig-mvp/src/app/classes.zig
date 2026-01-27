const objc = @import("objc");
const callbacks = @import("callbacks.zig");

fn inputHandlerClass() objc.Class {
    if (objc.getClass("ZigInputHandler")) |cls| return cls;

    const NSObject = objc.getClass("NSObject").?;
    const cls = objc.allocateClassPair(NSObject, "ZigInputHandler").?;
    if (!cls.addMethod("controlTextDidChange:", callbacks.controlTextDidChange)) {
        @panic("failed to add controlTextDidChange: method");
    }
    if (!cls.addMethod("control:textView:doCommandBySelector:", callbacks.controlTextViewDoCommandBySelector)) {
        @panic("failed to add control:textView:doCommandBySelector: method");
    }
    if (!cls.addMethod("onFocusLossTimer:", callbacks.onFocusLossTimer)) {
        @panic("failed to add onFocusLossTimer: method");
    }
    if (!cls.addMethod("onUpdateTimer:", callbacks.onUpdateTimer)) {
        @panic("failed to add onUpdateTimer: method");
    }
    if (!cls.addMethod("onSubmit:", callbacks.onSubmit)) {
        @panic("failed to add onSubmit: method");
    }
    objc.registerClassPair(cls);
    return cls;
}

fn dataSourceClass() objc.Class {
    if (objc.getClass("ZigTableDataSource")) |cls| return cls;

    const NSObject = objc.getClass("NSObject").?;
    const cls = objc.allocateClassPair(NSObject, "ZigTableDataSource").?;
    if (!cls.addMethod("numberOfRowsInTableView:", callbacks.numberOfRowsInTableView)) {
        @panic("failed to add numberOfRowsInTableView: method");
    }
    if (!cls.addMethod("tableView:objectValueForTableColumn:row:", callbacks.tableViewObjectValue)) {
        @panic("failed to add tableView:objectValueForTableColumn:row: method");
    }
    if (!cls.addMethod("tableView:shouldSelectRow:", callbacks.tableViewShouldSelectRow)) {
        @panic("failed to add tableView:shouldSelectRow: method");
    }
    if (!cls.addMethod("tableViewSelectionDidChange:", callbacks.tableViewSelectionDidChange)) {
        @panic("failed to add tableViewSelectionDidChange: method");
    }
    objc.registerClassPair(cls);
    return cls;
}

pub fn searchFieldClass() objc.Class {
    if (objc.getClass("ZigSearchField")) |cls| return cls;

    const NSTextField = objc.getClass("NSTextField").?;
    const cls = objc.allocateClassPair(NSTextField, "ZigSearchField").?;
    if (!cls.addMethod("cancelOperation:", callbacks.cancelOperation)) {
        @panic("failed to add cancelOperation: method");
    }
    if (!cls.addMethod("keyDown:", callbacks.keyDown)) {
        @panic("failed to add keyDown: method");
    }
    if (!cls.addMethod("performKeyEquivalent:", callbacks.performKeyEquivalent)) {
        @panic("failed to add performKeyEquivalent: method");
    }
    if (!cls.addMethod("becomeFirstResponder", callbacks.becomeFirstResponder)) {
        @panic("failed to add becomeFirstResponder method");
    }
    objc.registerClassPair(cls);
    return cls;
}

pub fn windowClass() objc.Class {
    if (objc.getClass("ZigBorderlessWindow")) |cls| return cls;

    const NSWindow = objc.getClass("NSWindow").?;
    const cls = objc.allocateClassPair(NSWindow, "ZigBorderlessWindow").?;
    if (!cls.addMethod("canBecomeKeyWindow", callbacks.windowCanBecomeKey)) {
        @panic("failed to add canBecomeKeyWindow method");
    }
    if (!cls.addMethod("canBecomeMainWindow", callbacks.windowCanBecomeMain)) {
        @panic("failed to add canBecomeMainWindow method");
    }
    if (!cls.addMethod("resignKeyWindow", callbacks.resignKeyWindow)) {
        @panic("failed to add resignKeyWindow method");
    }
    objc.registerClassPair(cls);
    return cls;
}

pub fn makeInputHandler() objc.Object {
    const cls = inputHandlerClass();
    return cls.msgSend(objc.Object, "alloc", .{}).msgSend(objc.Object, "init", .{});
}

pub fn makeDataSource() objc.Object {
    const cls = dataSourceClass();
    return cls.msgSend(objc.Object, "alloc", .{}).msgSend(objc.Object, "init", .{});
}
