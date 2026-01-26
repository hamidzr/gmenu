pub const SearchMethod = enum {
    direct,
    fuzzy,
    fuzzy1,
    fuzzy3,
    default,
};

pub const Options = struct {
    method: SearchMethod = .fuzzy,
    preserve_order: bool = false,
    limit: usize = 10,
    levenshtein_fallback: bool = true,
};

pub const Match = struct {
    index: usize,
    score: i32,
};
