"""Layer 1 of plan-bundle.py: Open Knowledge Format conformance of concept, index and log files."""

import unittest


import plan_bundle_support as S
pb = S.pb


def okf(files: dict[str, str]):
    return pb.okf_findings(S.bundle(files))


class LoadBundle(unittest.TestCase):
    def test_symbolic_link_is_reported(self):
        loaded = S.bundle({"plan.md": S.plan()}, links=("work/link.md",))
        self.assertEqual(loaded.findings, (pb.warning("work/link.md", 1, "symbolic link is not followed"),))
        self.assertIn("work", loaded.directories)

    def test_walk_order_is_depth_first_by_name(self):
        loaded = S.bundle({"b.md": "x", "a-b/x.md": "x", "a/b/c.md": "x"}, directories=("diagrams",))
        self.assertEqual(loaded.directories, ("", "a", "a/b", "a-b", "diagrams"))
        self.assertEqual([file.path for file in loaded.files], ["a/b/c.md", "a-b/x.md", "b.md"])

    def test_concepts_exclude_reserved_files(self):
        loaded = S.bundle({"plan.md": S.plan(), "index.md": "# X\n", "log/log.md": "# L\n"})
        self.assertEqual(list(loaded.concepts), ["plan"])

    def test_bundle_file_fields(self):
        file = pb.read_bundle_file("work/x-abcd.md", "---\ntype: T\n---\nbody")
        self.assertEqual((file.id, file.directory, file.name, file.body, file.body_line), ("work/x-abcd", "work", "x-abcd.md", "body", 4))
        self.assertEqual(file.data, {"type": "T"})


class ConceptDocuments(unittest.TestCase):
    def test_no_frontmatter(self):
        self.assertEqual(
            okf({"notes.md": "# Notes\n"}),
            (pb.error("notes.md", 1, "no frontmatter: a concept document starts with a --- line"),),
        )

    def test_unclosed_frontmatter(self):
        self.assertEqual(okf({"notes.md": "---\ntype: X\n"}), (pb.error("notes.md", 1, "frontmatter is not closed by a --- line"),))

    def test_yaml_error_carries_its_file_line(self):
        self.assertEqual(okf({"notes.md": "---\ntype: X\na: &b c\n---\n"}), (pb.error("notes.md", 3, "unsupported YAML construct: anchor"),))

    def test_frontmatter_must_be_a_mapping(self):
        self.assertEqual(okf({"notes.md": "---\n- a\n---\n"}), (pb.error("notes.md", 1, "frontmatter must be a YAML mapping"),))

    def test_no_type(self):
        self.assertEqual(okf({"notes.md": "---\ntitle: x\ntype: ~\n---\n"}), (pb.error("notes.md", 1, "frontmatter has no type"),))

    def test_type_must_be_a_string(self):
        self.assertEqual(okf({"notes.md": "---\na: 1\ntype: 3\n---\n"}), (pb.error("notes.md", 3, "type must be a string"),))

    def test_type_must_not_be_empty(self):
        self.assertEqual(okf({"notes.md": "---\ntype: ' '\n---\n"}), (pb.error("notes.md", 2, "type must not be empty"),))

    def test_document_of_unknown_type_passes(self):
        self.assertEqual(okf({"notes.md": "---\ntype: Note\n---\n"}), ())


class IndexFiles(unittest.TestCase):
    def test_entry_without_heading(self):
        self.assertEqual(
            okf({"work/index.md": "* [A](a.md) - b\n# H\n"}),
            (pb.error("work/index.md", 1, "an index entry must follow a # heading"),),
        )

    def test_line_that_is_no_entry(self):
        self.assertEqual(
            okf({"work/index.md": "# H\n\nfree text\n* [A](a b.md)\n"}),
            (
                pb.error("work/index.md", 3, "an index line is a # heading or an entry * [Title](url) - description"),
                pb.error("work/index.md", 4, "an index line is a # heading or an entry * [Title](url) - description"),
            ),
        )

    def test_index_needs_a_heading(self):
        self.assertEqual(okf({"work/index.md": "\n"}), (pb.error("work/index.md", 1, "an index needs a # heading"),))

    def test_entry_with_escaped_brackets_is_valid(self):
        self.assertEqual(okf({"work/index.md": "# H\n\n* [A \\[1\\]](a.md) - text\n- [B](b/)\n"}), ())

    def test_index_with_a_type_is_reserved(self):
        self.assertEqual(
            okf({"index.md": "---\ntype: Note\n---\n# H\n"}),
            (pb.error("index.md", 2, "index.md is a reserved file name and cannot be a concept document"),),
        )

    def test_only_the_root_index_has_frontmatter(self):
        self.assertEqual(
            okf({"work/index.md": '---\nokf_version: "0.2"\n---\n# H\n'}),
            (pb.error("work/index.md", 1, "only the bundle root index.md has frontmatter, work/index.md has none"),),
        )

    def test_root_index_frontmatter_must_be_a_mapping(self):
        self.assertEqual(
            okf({"index.md": "---\n---\n# H\n"}),
            (pb.error("index.md", 1, 'the root index.md frontmatter holds only okf_version: "0.2"'),),
        )

    def test_root_index_holds_only_a_quoted_okf_version(self):
        self.assertEqual(
            okf({"index.md": "---\nokf_version: 0.2\nextra: 1\n---\n# H\n"}),
            (
                pb.error("index.md", 2, 'okf_version is a quoted string such as "0.2"'),
                pb.error("index.md", 3, "the root index.md frontmatter holds only okf_version, not extra"),
            ),
        )

    def test_reserved_file_with_broken_frontmatter(self):
        self.assertEqual(okf({"index.md": "---\na: [\n---\n# H\n"}), (pb.error("index.md", 2, "flow sequence is not closed"),))
        self.assertEqual(okf({"index.md": "---\na: 1\n"}), (pb.error("index.md", 1, "frontmatter is not closed by a --- line"),))


class LogFiles(unittest.TestCase):
    def test_valid_log(self):
        self.assertEqual(okf({"log.md": "# Directory Update Log\n\n## 2026-10-28\n* entry\n\n## 2026-10-27\n- entry\n"}), ())

    def test_log_starts_with_a_title(self):
        self.assertEqual(
            okf({"log.md": "## 2026-10-28\n* x\n"}),
            (pb.error("log.md", 1, "a log starts with a # heading: # Directory Update Log"),),
        )
        self.assertEqual(okf({"log.md": ""}), (pb.error("log.md", 1, "a log starts with a # heading: # Directory Update Log"),))

    def test_date_heading_must_be_a_date(self):
        self.assertEqual(
            okf({"log.md": "# Log\n## 2026-02-30\n"}),
            (pb.error("log.md", 2, "date heading 2026-02-30 is not a YYYY-MM-DD date"),),
        )

    def test_dates_run_newest_first(self):
        self.assertEqual(
            okf({"log.md": "# Log\n## 2026-10-27\n## 2026-10-28\n## 2026-10-28\n"}),
            (
                pb.error("log.md", 3, "date headings run newest first, but 2026-10-28 follows 2026-10-27"),
                pb.error("log.md", 4, "date headings run newest first, but 2026-10-28 follows 2026-10-28"),
            ),
        )

    def test_log_line_kinds(self):
        self.assertEqual(
            okf({"log.md": "# Log\n* early\nprose\n"}),
            (
                pb.error("log.md", 2, "a log entry must follow a ## YYYY-MM-DD heading"),
                pb.error("log.md", 3, "a log line is a ## YYYY-MM-DD heading or an entry starting with *"),
            ),
        )

    def test_log_in_a_subdirectory_follows_the_same_structure(self):
        self.assertEqual(okf({"log/log.md": "text\n"})[0].message, "a log starts with a # heading: # Directory Update Log")


if __name__ == "__main__":
    unittest.main()
