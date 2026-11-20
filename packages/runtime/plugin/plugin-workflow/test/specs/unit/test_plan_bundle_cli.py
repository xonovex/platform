"""The command line core, the family level and the ECMAScript helpers of plan-bundle.py."""

import json
import math
import unittest


import plan_bundle_support as S
pb = S.pb
Entry = pb.DirectoryEntry

BASE = {"plan.md": S.plan(), "log/2026-10-28-writing-plan-abcd.md": S.log_entry()}


def options(command: str, mode: str | None = None) -> object:
    return pb.Options(command, "b", (), "text", mode)


class Arguments(unittest.TestCase):
    def test_valid_commands(self):
        self.assertEqual(
            pb.parse_arguments(["validate", "b", "--schemas", "x", "--schemas", "y", "--format", "json"]),
            pb.Options("validate", "b", ("x", "y"), "json", None),
        )
        self.assertEqual(pb.parse_arguments(["index", "--write", "b", "--write"]), pb.Options("index", "b", (), "text", "write"))
        self.assertEqual(pb.parse_arguments(["log", "b", "--check"]), pb.Options("log", "b", (), "text", "check"))

    def test_help(self):
        self.assertEqual(pb.parse_arguments(["bogus", "-h"]), pb.HelpRequest())
        self.assertEqual(pb.parse_arguments(["--help"]), pb.HelpRequest())

    def test_usage_errors(self):
        cases = {
            (): "missing command",
            ("check",): "unknown command: check",
            ("validate",): "missing bundle directory",
            ("validate", "a", "b"): "unexpected argument: b",
            ("validate", "b", "--schemas"): "--schemas needs a value",
            ("index", "b", "--schemas", "x", "--check"): "--schemas applies only to validate",
            ("validate", "b", "--format", "xml"): "--format is text or json, not xml",
            ("validate", "b", "--check"): "--check applies only to index and log",
            ("index", "b", "--check", "--write"): "use --check or --write, not both",
            ("index", "b", "-x"): "unknown option: -x",
            ("log", "b"): "log needs --check or --write",
        }
        for argv, message in cases.items():
            with self.subTest(argv=argv):
                self.assertEqual(pb.parse_arguments(list(argv)), pb.UsageError(message))

    def test_usage_names_the_python_runner_and_the_family(self):
        self.assertTrue(pb.USAGE.startswith("Usage: uv run plan-bundle.py <command> <bundle-dir> [options]\n"))
        self.assertIn("is a family", pb.USAGE)


class Family(unittest.TestCase):
    def test_members_in_name_order(self):
        listing = (
            Entry("roadmap.md", "file"),
            Entry("notes.md", "file"),
            Entry("zeta", "directory", ("log", "work")),
            Entry("Beta", "directory", ("log",)),
            Entry("alpha", "directory", ("log",)),
            Entry(".hidden", "directory", ("log",)),
            Entry("no-bundle", "directory", ("work",)),
            Entry("linked", "other"),
        )
        self.assertEqual(pb.family_members(listing), ("Beta", "alpha", "zeta"))

    def test_family_without_bundles(self):
        self.assertEqual(pb.family_members((Entry("roadmap.md", "file"), Entry("x", "directory"))), ())

    def test_directories_that_are_no_family(self):
        cases = {
            "no roadmap": (Entry("a", "directory", ("log",)),),
            "roadmap is a directory": (Entry("roadmap.md", "directory"),),
            "a bundle with a plan": (Entry("roadmap.md", "file"), Entry("plan.md", "file")),
            "a bundle with a log": (Entry("roadmap.md", "file"), Entry("log", "directory")),
        }
        for label, listing in cases.items():
            with self.subTest(label):
                self.assertIsNone(pb.family_members(listing))

    def test_a_file_named_log_does_not_make_a_bundle(self):
        self.assertEqual(pb.family_members((Entry("roadmap.md", "file"), Entry("log", "file"))), ())

    def test_prefix_finding(self):
        self.assertEqual(pb.prefix_finding(pb.error("plan.md", 3, "m"), "alpha"), pb.error("alpha/plan.md", 3, "m"))
        self.assertEqual(pb.prefix_finding(pb.error("/s/x.schema.json", 1, "m"), "alpha"), pb.error("/s/x.schema.json", 1, "m"))


class RunBundle(unittest.TestCase):
    def test_validate(self):
        run = pb.run_bundle(options("validate"), S.bundle(S.with_generated(BASE)), S.registry())
        self.assertEqual(run, pb.BundleRun((), (), (), False))

    def test_check_reports_stale_files(self):
        run = pb.run_bundle(options("index", "check"), S.bundle(BASE), S.registry())
        self.assertEqual(S.messages(run.findings), ["generated file is missing; run the index command with --write"] * 2)
        self.assertFalse(run.writing)

    def test_write_returns_only_changed_files_and_obsolete_ones(self):
        files = {**S.with_generated(BASE), "empty/index.md": "# X\n"}
        files["log/index.md"] = "# old\n"
        run = pb.run_bundle(options("index", "write"), S.bundle(files), S.registry())
        self.assertEqual([path for path, _ in run.writes], ["log/index.md"])
        self.assertEqual(run.deletes, ("empty/index.md",))
        self.assertEqual((run.findings, run.writing), ((), True))

    def test_write_with_generator_findings_writes_nothing(self):
        files = {**BASE, "log/2026-10-28-x-abcd.md": "---\na: [\n---\n"}
        run = pb.run_bundle(options("log", "write"), S.bundle(files), S.registry())
        self.assertEqual(run.writes, ())
        self.assertTrue(run.writing)
        self.assertEqual(
            S.messages(run.findings),
            ["generated file is missing; run the log command with --write", "log.md cannot list this entry until its frontmatter parses"],
        )


class Report(unittest.TestCase):
    FINDINGS = (pb.error("/cwd/a/plan.md", 3, "bad"), pb.warning("/elsewhere/x.json", 1, "odd"))

    def test_text(self):
        self.assertEqual(
            pb.render_report(self.FINDINGS, None, "text", "/cwd"),
            "a/plan.md:3: error: bad\n/elsewhere/x.json:1: warning: odd\n2 findings: 1 error, 1 warning\n",
        )

    def test_text_with_written_files(self):
        self.assertEqual(pb.render_report((), ["log.md"], "text", "/cwd"), "0 findings: 0 errors, 0 warnings; wrote 1 file\n")

    def test_json(self):
        report = json.loads(pb.render_report(self.FINDINGS, ["x"], "json", "/cwd", ("fam", ("alpha",))))
        self.assertEqual(
            report,
            {
                "findings": [
                    {"path": "a/plan.md", "line": 3, "severity": "error", "message": "bad"},
                    {"path": "/elsewhere/x.json", "line": 1, "severity": "warning", "message": "odd"},
                ],
                "summary": {"findings": 2, "errors": 1, "warnings": 1},
                "written": ["x"],
                "bundles": ["alpha"],
            },
        )

    def test_json_escapes_a_lone_surrogate(self):
        text = pb.render_report((pb.error("a", 1, "x\ud800y"),), None, "json", "/")
        self.assertIn('"x\\ud800y"', text)
        self.assertTrue(text.endswith("}\n"))

    def test_family_without_bundles(self):
        self.assertEqual(
            pb.render_report((), None, "text", "/cwd", ("plans", ())),
            "plans is a family without bundles: no child directory holds log/\n",
        )
        self.assertEqual(json.loads(pb.render_report((), None, "json", "/cwd", ("plans", ())))["bundles"], [])

    def test_family_with_bundles_prints_the_merged_findings(self):
        self.assertEqual(
            pb.render_report((pb.error("alpha/plan.md", 1, "m"),), None, "text", "/cwd", ("plans", ("alpha",))),
            "alpha/plan.md:1: error: m\n1 finding: 1 error, 0 warnings\n",
        )

    def test_display_path(self):
        self.assertEqual(pb.display_path("rel/x", "/cwd"), "rel/x")
        self.assertEqual(pb.display_path("/cwd/a/b", "/cwd"), "a/b")
        self.assertEqual(pb.display_path("/other/a", "/cwd"), "/other/a")

    def test_exit_status(self):
        self.assertEqual(pb.exit_status(()), 0)
        self.assertEqual(pb.exit_status((pb.warning("a", 1, "w"),)), 0)
        self.assertEqual(pb.exit_status((pb.warning("a", 1, "w"), pb.error("a", 1, "e"))), 1)

    def test_finish_removes_duplicates_and_sorts(self):
        found = [pb.warning("b", 1, "w"), pb.error("a", 2, "e"), pb.error("a", 2, "e"), pb.error("a", 1, "z"), pb.warning("a", 1, "a")]
        self.assertEqual(pb.finish(found), (pb.error("a", 1, "z"), pb.warning("a", 1, "a"), pb.error("a", 2, "e"), pb.warning("b", 1, "w")))


class Helpers(unittest.TestCase):
    def test_compare_text_uses_code_units(self):
        self.assertEqual(sorted(["b", "B", "a", "\U0001f600", "￿"], key=pb.text_key), ["B", "a", "b", "\U0001f600", "￿"])
        self.assertEqual(pb.compare_text("a", "a"), 0)
        self.assertEqual(pb.compare_text("a", "b"), -1)

    def test_trim_and_collapse_use_the_js_white_space_set(self):
        self.assertEqual(pb.js_trim("﻿  x  "), "x")
        self.assertEqual(pb.collapse(" a 　\t b "), "a b")
        self.assertEqual(pb.js_trim("\x85x"), "\x85x")

    def test_js_keys(self):
        self.assertEqual(pb.js_keys({"b": 1, "10": 1, "a": 1, "2": 1, "01": 1}), ["2", "10", "b", "a", "01"])

    def test_js_number(self):
        cases = {1.0: "1", 1e21: "1e+21", 1e20: "100000000000000000000", 0.1: "0.1", 1e-7: "1e-7", 1.5e-7: "1.5e-7",
                 -0.0: "0", 123.456: "123.456", 0.000001: "0.000001", math.inf: "Infinity", -math.inf: "-Infinity"}
        for value, text in cases.items():
            with self.subTest(value=value):
                self.assertEqual(pb.js_number(value), text)
        self.assertEqual(pb.js_number(math.nan), "NaN")

    def test_js_stringify_and_string(self):
        value = {"b": [1.0, None, True, math.nan], "1": "q\"\n\x01"}
        self.assertEqual(pb.js_stringify(value), '{"1":"q\\"\\n\\u0001","b":[1,null,true,null]}')
        self.assertEqual([pb.js_string(v) for v in (None, False, 2.0, [1, None, "x"], {})], ["null", "false", "2", "1,,x", "[object Object]"])

    def test_js_pattern(self):
        self.assertEqual(pb.compile_js_pattern("^a$").search("a\n"), None)
        self.assertIsNone(pb.compile_js_pattern("^a.b$").search("a\nb"))
        self.assertIsNotNone(pb.compile_js_pattern(r"^\s$").search(" "))
        self.assertIsNotNone(pb.compile_js_pattern("^(?<name>x)[$.]$").search("x$"))
        self.assertIsNone(pb.compile_js_pattern(r"^\d$").search("٣"))

    def test_dates(self):
        self.assertTrue(pb.is_calendar_date("2024-02-29"))
        self.assertFalse(pb.is_calendar_date("2023-02-29"))
        # Date.UTC reads the years 0 to 99 as 1900 to 1999: 1900 is no leap year, 1996 is one.
        self.assertFalse(pb.is_calendar_date("0000-02-29"))
        self.assertTrue(pb.is_calendar_date("0096-02-29"))
        self.assertTrue(pb.is_datetime_offset("2026-10-28T09:00Z"))
        self.assertTrue(pb.is_datetime_offset("2026-10-28T09:00:00.123456+05:30"))
        for text in ("2026-10-28T09:00:00", "2026-10-28 09:00Z", "2026-10-28T24:00Z", "2026-10-28T09:00+24:00", 5):
            with self.subTest(text=text):
                self.assertFalse(pb.is_datetime_offset(text))

    def test_instant_ms(self):
        self.assertEqual(pb.instant_ms("1970-01-01T00:00:00Z"), 0)
        self.assertEqual(pb.instant_ms("1970-01-01T01:00:00.9999+01:00"), 999)
        self.assertEqual(pb.instant_ms("1969-12-31T23:00-01:00"), 0)
        self.assertEqual(pb.instant_ms("2026-10-28T09:00:00Z"), 1793178000000)
        self.assertIsNone(pb.instant_ms("2026-10-28"))


if __name__ == "__main__":
    unittest.main()
