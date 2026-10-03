use moon_nix_runtime::serialize::{escape_nix_string, quote_posix};

#[test]
fn nix_string_escaping_protects_quotes_backslashes_and_interpolation() {
    let value = "a\\b\"${value}";

    let escaped = escape_nix_string(value);

    assert_eq!(escaped, "a\\\\b\\\"\\${value}");
}

#[test]
fn posix_quoting_keeps_apostrophes_inside_one_argument() {
    let value = "can't split";

    let quoted = quote_posix(value);

    assert_eq!(quoted, "'can'\\''t split'");
}
