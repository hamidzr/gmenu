const search = @import("search.zig");

pub const Config = struct {
    title: [:0]const u8,
    placeholder: [:0]const u8,
    search: search.Options,
    window_width: f64,
    window_height: f64,
    field_height: f64,
    padding: f64,
};

pub fn defaults() Config {
    return .{
        .title = "zmenu",
        .placeholder = "Type to filter",
        .search = .{
            .method = .fuzzy,
            .preserve_order = false,
            .limit = 10,
        },
        .window_width = 520,
        .window_height = 360,
        .field_height = 24,
        .padding = 8,
    };
}
