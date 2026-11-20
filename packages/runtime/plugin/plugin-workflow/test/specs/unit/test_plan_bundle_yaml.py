"""The YAML subset parser and the frontmatter split of plan-bundle.py."""

import math
import unittest


from plan_bundle_support import pb


def value(text: str):
    result = pb.parse_yaml(text)
    assert isinstance(result, pb.YamlOk), result
    return result.value


def failure(text: str) -> tuple[int, str]:
    result = pb.parse_yaml(text)
    assert isinstance(result, pb.YamlFail), result
    return result.line, result.message


class BlockCollections(unittest.TestCase):
    def test_nested_mapping_and_sequence(self):
        text = "a: 1\nb:\n  c: x\n  d:\n    - 1\n    - two\n"
        self.assertEqual(value(text), {"a": 1, "b": {"c": "x", "d": [1, "two"]}})

    def test_sequence_at_the_indentation_of_its_key(self):
        self.assertEqual(value("list:\n- a\n- b\nnext: 1\n"), {"list": ["a", "b"], "next": 1})

    def test_compact_mapping_in_a_sequence(self):
        text = "items:\n  - id: c1\n    text: one\n  - id: c2\n"
        self.assertEqual(value(text), {"items": [{"id": "c1", "text": "one"}, {"id": "c2"}]})

    def test_compact_nested_sequence(self):
        self.assertEqual(value("- - a\n  - b\n- c\n"), [["a", "b"], "c"])

    def test_empty_values_are_null(self):
        self.assertEqual(value("a:\nb: ~\n"), {"a": None, "b": None})

    def test_empty_sequence_entry_is_null(self):
        self.assertEqual(value("-\n- x\n"), [None, "x"])

    def test_comments_and_blank_lines(self):
        self.assertEqual(value("# head\na: 1 # note\n\n  # indented\nb: 'x' # q\n"), {"a": 1, "b": "x"})

    def test_empty_document_is_null(self):
        self.assertIsNone(value("# only a comment\n\n"))

    def test_byte_order_mark_and_crlf(self):
        self.assertEqual(value("\ufeffa: 1\r\nb: 2\r\n"), {"a": 1, "b": 2})

    def test_quoted_keys(self):
        self.assertEqual(value("'a b': 1\n\"c\": 2\n"), {"a b": 1, "c": 2})

    def test_key_named_proto_is_plain_data(self):
        self.assertEqual(value("__proto__: x\n"), {"__proto__": "x"})


class Scalars(unittest.TestCase):
    def test_core_schema_null_and_booleans(self):
        self.assertEqual(
            value("[~, null, Null, NULL, true, True, TRUE, false, False, FALSE, yes]"),
            [None, None, None, None, True, True, True, False, False, False, "yes"],
        )

    def test_integers(self):
        parsed = value("[1, -2, +3, 0o17, 0x1F, 007]")
        self.assertEqual(parsed, [1, -2, 3, 15, 31, 7])
        self.assertTrue(all(type(item) is int for item in parsed))

    def test_floats_and_special_numbers(self):
        parsed = value("[1.5, .5, 1., 1e3, -.inf, .Inf, .NaN]")
        self.assertEqual(parsed[:5], [1.5, 0.5, 1.0, 1000.0, -math.inf])
        self.assertEqual(parsed[5], math.inf)
        self.assertTrue(math.isnan(parsed[6]))

    def test_timestamp_stays_a_string(self):
        self.assertEqual(value("at: 2026-10-28T09:00:00Z\n"), {"at": "2026-10-28T09:00:00Z"})

    def test_integer_too_large(self):
        self.assertEqual(failure("n: 9007199254740992\n"), (1, "integer 9007199254740992 is too large to hold exactly"))

    def test_plain_scalar_folds_continuation_lines(self):
        self.assertEqual(value("a: one\n  two\n\n  three\nb: x\n"), {"a": "one two\nthree", "b": "x"})

    def test_plain_scalar_cannot_start_with_an_indicator(self):
        self.assertEqual(failure("a: @x\n"), (1, 'a plain scalar cannot start with "@"'))
        self.assertEqual(failure("`x: 1\n"), (1, 'a plain scalar cannot start with "`"'))

    def test_mapping_value_not_allowed_in_a_plain_scalar(self):
        self.assertEqual(failure("a: b: c\n"), (1, "a mapping value is not allowed here"))

    def test_single_quoted_doubles_a_quote_and_folds(self):
        self.assertEqual(value("a: 'it''s\n  folded'\n"), {"a": "it's folded"})

    def test_double_quoted_escapes(self):
        text = 'a: "t\\tn\\n\\x41\\u00e9\\U0001F600\\\\\\"\\/\\_"\n'
        self.assertEqual(value(text), {"a": "t\tn\nA\u00e9\U0001f600\\\"/\xa0"})

    def test_escaped_surrogate_pair_is_one_character(self):
        self.assertEqual(value('a: "\\uD83D\\uDE00"\n'), {"a": "\U0001f600"})

    def test_double_quoted_line_continuation(self):
        self.assertEqual(value('a: "one \\\n  two"\n'), {"a": "one two"})

    def test_invalid_escape(self):
        self.assertEqual(failure('a: "\\q"\n'), (1, "invalid escape sequence \\q in a double-quoted scalar"))

    def test_escape_beyond_unicode(self):
        self.assertEqual(failure('a: "\\U00110000"\n'), (1, "escape \\U00110000 is not a code point"))

    def test_unclosed_quotes(self):
        self.assertEqual(failure('a: "open\n'), (1, "double-quoted scalar is not closed"))
        self.assertEqual(failure("a: 'open\n"), (1, "single-quoted scalar is not closed"))

    def test_text_after_a_quoted_value(self):
        self.assertEqual(failure("a: 'x' y\n"), (1, "unexpected text after the value"))


class BlockScalars(unittest.TestCase):
    def test_literal_clip(self):
        self.assertEqual(value("a: |\n  one\n   two\n\nb: 1\n"), {"a": "one\n two\n", "b": 1})

    def test_folded_strip(self):
        self.assertEqual(value("a: >-\n  one\n  two\n\n  three\n"), {"a": "one two\nthree"})

    def test_folded_keeps_more_indented_lines(self):
        self.assertEqual(value("a: >\n  one\n    code\n  two\n"), {"a": "one\n  code\ntwo\n"})

    def test_keep_chomping(self):
        self.assertEqual(value("a: |+\n  one\n\n\nb: 1\n"), {"a": "one\n\n\n", "b": 1})

    def test_explicit_indentation(self):
        self.assertEqual(value("a: |2\n    one\n  two\n"), {"a": "  one\ntwo\n"})

    def test_empty_block_scalar(self):
        self.assertEqual(value("a: |\nb: 1\n"), {"a": "", "b": 1})

    def test_invalid_header(self):
        self.assertEqual(failure("a: |x\n  one\n"), (1, "invalid block scalar header"))

    def test_leading_empty_line_more_indented(self):
        self.assertEqual(
            failure("a: |\n     \n  one\n"),
            (2, "a leading empty line of a block scalar is indented more than its content"),
        )


class FlowCollections(unittest.TestCase):
    def test_flow_mapping_and_sequence(self):
        self.assertEqual(
            value("g: {by: x/1, at: 2026-10-28T09:00:00Z}\nl: [a, 'b', \"c\", [d], {e: f}]\n"),
            {"g": {"by": "x/1", "at": "2026-10-28T09:00:00Z"}, "l": ["a", "b", "c", ["d"], {"e": "f"}]},
        )

    def test_flow_mapping_entry_without_value(self):
        self.assertEqual(value("m: {a, b: }\n"), {"m": {"a": None, "b": None}})

    def test_multi_line_flow_with_closing_bracket_at_parent_indent(self):
        self.assertEqual(value("l: [\n  a,\n  b # note\n]\nn: 1\n"), {"l": ["a", "b"], "n": 1})

    def test_continuation_line_must_be_indented(self):
        self.assertEqual(
            failure("a:\n  l: [x,\n  y]\n"), (3, "a continuation line must be indented more than its parent")
        )

    def test_unclosed_flow_collections(self):
        self.assertEqual(failure("l: [a, b\n"), (1, "flow sequence is not closed"))
        self.assertEqual(failure("m: {a: b\n"), (1, "flow mapping is not closed"))

    def test_empty_entries(self):
        self.assertEqual(failure("l: [a, , b]\n"), (1, "flow sequence has an empty entry"))
        self.assertEqual(failure("m: {a: 1, , b: 2}\n"), (1, "flow mapping has an empty entry"))

    def test_missing_separator(self):
        self.assertEqual(failure("l: ['a' 'b']\n"), (1, "expected , or ] in a flow sequence"))
        self.assertEqual(failure("m: {a: 'x' 'y'}\n"), (1, "expected , or } in a flow mapping"))

    def test_flow_pair_is_unsupported(self):
        self.assertEqual(failure("l: [a: b]\n"), (1, "unsupported YAML construct: flow pair"))

    def test_block_constructs_inside_flow(self):
        self.assertEqual(failure("l: [|]\n"), (1, "a block scalar cannot appear inside a flow collection"))
        self.assertEqual(failure("l: [- a]\n"), (1, "a block sequence entry cannot appear inside a flow collection"))

    def test_flow_key_errors(self):
        self.assertEqual(failure("m: {[a]: b}\n"), (1, "unsupported YAML construct: complex key"))
        self.assertEqual(failure("m: {: b}\n"), (1, "a mapping entry has no key"))

    def test_flow_duplicate_key(self):
        self.assertEqual(failure("m: {a: 1, a: 2}\n"), (1, 'duplicate key "a"'))

    def test_expected_a_value(self):
        self.assertEqual(failure("l: [: ]\n"), (1, "expected a value"))


class Errors(unittest.TestCase):
    def test_unsupported_constructs(self):
        cases = {
            "a: &x 1\n": "anchor",
            "a: *x\n": "alias",
            "a: !tag 1\n": "tag",
            "? a\n": "complex key",
            "[a]: b\n": "complex key",
            "&a b: 1\n": "anchor",
            "a: 1\n---\nb: 2\n": "multi-document stream",
            "a: 1\n...\n": "multi-document stream",
            "%YAML 1.2\na: 1\n": "directive",
        }
        for text, name in cases.items():
            with self.subTest(text=text):
                self.assertEqual(failure(text)[1], f"unsupported YAML construct: {name}")

    def test_tab_in_indentation(self):
        self.assertEqual(failure("a:\n\tb: 1\n"), (2, "tab character in indentation"))

    def test_duplicate_block_key(self):
        self.assertEqual(failure("a: 1\nb: 2\na: 3\n"), (3, 'duplicate key "a"'))

    def test_unexpected_indentation(self):
        self.assertEqual(failure("a:\n  b:\n    c: 1\n   d: 2\n"), (4, "unexpected indentation"))

    def test_block_sequence_on_the_line_of_its_key(self):
        self.assertEqual(failure("a: - b\n"), (1, "a block sequence cannot start on the line of its key"))

    def test_sequence_entry_in_a_mapping(self):
        self.assertEqual(failure("a: 1\n- b\n"), (2, "expected a mapping entry, found a sequence entry"))

    def test_line_that_is_no_mapping_entry(self):
        self.assertEqual(failure("a: 1\nplain\n"), (2, "expected a mapping entry"))

    def test_mapping_entry_without_key(self):
        self.assertEqual(failure(": x\n"), (1, "a mapping entry has no key"))

    def test_content_after_the_document(self):
        self.assertEqual(failure("- a\nb: 1\n"), (2, "unexpected content after the document"))


class Locations(unittest.TestCase):
    def test_pointers_map_to_lines_from_the_first_line(self):
        result = pb.parse_yaml("a: 1\nb:\n  - x\n  - {c: d}\n'e/f~': 2\n", 2)
        self.assertEqual(
            dict(result.lines),
            {"": 2, "/a": 2, "/b": 3, "/b/0": 4, "/b/1": 5, "/b/1/c": 5, "/e~1f~0": 6},
        )

    def test_error_line_counts_from_the_first_line(self):
        self.assertEqual(pb.parse_yaml("a: 1\na: 2\n", 2), pb.YamlFail(3, 'duplicate key "a"'))


class Frontmatter(unittest.TestCase):
    def test_present(self):
        block = pb.split_frontmatter("---\na: 1\n---\n\n# Body\n")
        self.assertEqual(block, pb.FrontmatterBlock("present", "a: 1\n", 2, "\n# Body\n", 4))

    def test_absent(self):
        self.assertEqual(pb.split_frontmatter("# Title\n").kind, "absent")

    def test_unclosed(self):
        self.assertEqual(pb.split_frontmatter("---\na: 1\n").kind, "unclosed")

    def test_marker_with_trailing_space_and_crlf(self):
        block = pb.split_frontmatter("\ufeff--- \r\na: 1\r\n---\r\nx")
        self.assertEqual((block.kind, block.yaml, block.body), ("present", "a: 1\n", "x"))


if __name__ == "__main__":
    unittest.main()
