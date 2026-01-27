const std = @import("std");
const appconfig = @import("../config.zig");
const parse = @import("parse.zig");

pub fn applyEnv(allocator: std.mem.Allocator, config: *appconfig.Config) !void {
    if (envValue(allocator, "GMENU_TITLE")) |value| {
        config.title = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_PROMPT")) |value| {
        config.placeholder = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MENU_ID")) |value| {
        config.menu_id = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_INITIAL_QUERY")) |value| {
        config.initial_query = value;
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    if (envValue(allocator, "GMENU_TERMINAL_MODE")) |value| {
        config.terminal_mode = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_FOLLOW_STDIN")) |value| {
        config.follow_stdin = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_IPC_ONLY")) |value| {
        config.ipc_only = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    if (envValue(allocator, "GMENU_MIN_WIDTH")) |value| {
        config.window_width = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MIN_HEIGHT")) |value| {
        config.window_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MAX_WIDTH")) |value| {
        config.max_width = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_MAX_HEIGHT")) |value| {
        config.max_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_ROW_HEIGHT")) |value| {
        config.row_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_FIELD_HEIGHT")) |value| {
        config.field_height = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_PADDING")) |value| {
        config.padding = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_NUMERIC_COLUMN_WIDTH")) |value| {
        config.numeric_column_width = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_ICON_COLUMN_WIDTH")) |value| {
        config.icon_column_width = try std.fmt.parseFloat(f64, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_ALTERNATE_ROWS")) |value| {
        config.alternate_rows = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_BACKGROUND_COLOR")) |value| {
        config.background_color = try parse.parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_LIST_BACKGROUND_COLOR")) |value| {
        config.list_background_color = try parse.parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_FIELD_BACKGROUND_COLOR")) |value| {
        config.field_background_color = try parse.parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_TEXT_COLOR")) |value| {
        config.text_color = try parse.parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_SECONDARY_TEXT_COLOR")) |value| {
        config.secondary_text_color = try parse.parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_SELECTION_COLOR")) |value| {
        config.selection_color = try parse.parseColorOptional(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }

    if (envValue(allocator, "GMENU_SEARCH_METHOD")) |value| {
        try parse.applySearchMethod(config, value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_PRESERVE_ORDER")) |value| {
        config.search.preserve_order = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_LEVENSHTEIN_FALLBACK")) |value| {
        config.search.levenshtein_fallback = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_AUTO_ACCEPT")) |value| {
        config.auto_accept = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_ACCEPT_CUSTOM_SELECTION")) |value| {
        config.accept_custom_selection = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_NO_NUMERIC_SELECTION")) |value| {
        config.no_numeric_selection = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
    if (envValue(allocator, "GMENU_SHOW_ICONS")) |value| {
        config.show_icons = try parse.parseBool(value);
    } else |err| {
        if (err != error.EnvironmentVariableNotFound) return err;
    }
}

fn envValue(allocator: std.mem.Allocator, name: []const u8) ![:0]const u8 {
    const value = try std.process.getEnvVarOwned(allocator, name);
    defer allocator.free(value);
    return allocator.dupeZ(u8, value);
}
