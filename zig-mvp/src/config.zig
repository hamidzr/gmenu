const search = @import("search.zig");

pub const Config = struct {
    title: [:0]const u8,
    placeholder: [:0]const u8,
    menu_id: [:0]const u8,
    initial_query: [:0]const u8,
    search: search.Options,
    terminal_mode: bool,
    auto_accept: bool,
    accept_custom_selection: bool,
    window_width: f64,
    window_height: f64,
    max_width: f64,
    max_height: f64,
    field_height: f64,
    padding: f64,
    no_numeric_selection: bool,
    numeric_column_width: f64,
    show_score: bool,
    score_column_width: f64,
};

pub fn defaults() Config {
    return .{
        .title = "zmenu",
        .placeholder = "Type to filter",
        .menu_id = "",
        .initial_query = "",
        .search = .{
            .method = .fuzzy,
            .preserve_order = false,
            .limit = 10,
        },
        .terminal_mode = false,
        .auto_accept = false,
        .accept_custom_selection = true,
        .window_width = 520,
        .window_height = 360,
        .max_width = 1920,
        .max_height = 1080,
        .field_height = 24,
        .padding = 8,
        .no_numeric_selection = false,
        .numeric_column_width = 28,
        .show_score = false,
        .score_column_width = 60,
    };
}
