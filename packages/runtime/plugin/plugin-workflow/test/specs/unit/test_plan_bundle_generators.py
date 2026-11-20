"""The index and log generators of plan-bundle.py, the stale check and the complete validation."""

import unittest


import plan_bundle_support as S
pb = S.pb


def indexes(files: dict[str, str], directories: tuple[str, ...] = ()):
    return pb.generate_indexes(S.bundle(files, directories))


def index_files(files: dict[str, str], directories: tuple[str, ...] = ()) -> dict[str, str]:
    generated = indexes(files, directories)
    assert generated.findings == (), generated.findings
    return dict(generated.files)


def at(day: int, hour: int = 9) -> str:
    return f"2026-10-{day:02d}T{hour:02d}:00:00Z"


BASE = {
    "plan.md": S.plan(),
    "log/2026-10-28-writing-plan-abcd.md": S.log_entry(),
}


class Indexes(unittest.TestCase):
    def test_root_index(self):
        files = {
            **BASE,
            "close-out.md": S.doc(S.header("Workflow Close-Out", "workflow/close-out/1", "Close out", at(29))),
            "notes-abcd.md": S.doc(S.header("Note", "x", "Notes", at(27))),
            "decisions/d-abcd.md": S.decision(),
            "diagrams/flow.md": S.doc("type: Note"),
        }
        self.assertEqual(
            index_files(files, ("diagrams",))["index.md"],
            '---\nokf_version: "0.2"\n---\n\n'
            "# Documents\n\n"
            "* [The plan](plan.md) - A description.\n"
            "* [Close out](close-out.md) - A description.\n"
            "* [Notes](notes-abcd.md) - A description.\n\n"
            "# Directories\n\n"
            "* [decisions](decisions/) - Decisions that shape the plan\n"
            "* [log](log/) - One entry for each skill run\n",
        )

    def test_diagrams_get_no_index_and_other_directories_do(self):
        files = {**BASE, "diagrams/flow.md": S.doc("type: Note"), "sketches/run.md": S.doc("type: Note\ntitle: Run")}
        generated = index_files(files)
        self.assertNotIn("diagrams/index.md", generated)
        self.assertEqual(generated["sketches/index.md"], "# Documents\n\n* [Run](run.md)\n")
        self.assertIn("* [sketches](sketches/)\n", generated["index.md"])

    def test_refinement_is_a_generic_directory(self):
        files = {**BASE, "refinement/2026-10-28-x/pack.md": S.doc("type: Workflow Refinement Session\ntitle: Pack")}
        generated = index_files(files)
        self.assertEqual(generated["refinement/index.md"], "# Directories\n\n* [2026-10-28-x](2026-10-28-x/)\n")
        self.assertIn("* [refinement](refinement/)\n", generated["index.md"])

    def test_decisions_group_by_state_and_number_oldest_first(self):
        files = {
            **BASE,
            "decisions/b-abcd.md": S.decision("open", "Second open", at(20)),
            "decisions/a-abcd.md": S.decision("open", "First open", at(10)),
            "decisions/c-abcd.md": S.decision("settled", "Settled one", at(25)),
            "decisions/d-abcd.md": S.decision("settled", "Old", at(5), extra="status: deprecated\n"),
            "decisions/e-abcd.md": S.decision("partly-settled", "Partly", at(1)),
            "decisions/notes.md": S.doc("type: Note\ntitle: Notes"),
        }
        self.assertEqual(
            index_files(files)["decisions/index.md"],
            "# Settled\n\n* [1. Settled one](c-abcd.md) - A description.\n\n"
            "# Open\n\n* [2. First open](a-abcd.md) - A description.\n* [3. Second open](b-abcd.md) - A description.\n\n"
            "# Partly settled\n\n* [4. Partly](e-abcd.md) - A description.\n\n"
            "# Superseded\n\n* [5. Old](d-abcd.md) - A description.\n\n"
            "# Documents\n\n* [Notes](notes.md)\n",
        )

    def test_decision_without_a_known_state(self):
        generated = indexes({**BASE, "decisions/a-abcd.md": S.decision("maybe")})
        self.assertNotIn("decisions/index.md", dict(generated.files))
        self.assertEqual(
            generated.findings,
            (pb.error("decisions/a-abcd.md", 7, "state must be settled, open, paused or partly-settled to group the decision"),),
        )

    def test_work_runs_in_dependency_order_then_creation_time(self):
        files = {
            **BASE,
            "work/a-abcd.md": S.work("First made", at=at(1), extra="depends_on: [work/c-abcd]\n"),
            "work/b-abcd.md": S.work("Second made", at=at(2)),
            "work/c-abcd.md": S.work("Third made", at=at(3)),
            "work/w-abcd.md": S.work("Waiting", "waiting", at(4), "waits_on: [decisions/x-abcd]\n"),
            "work/x-abcd.md": S.work("No time", at="never"),
        }
        self.assertEqual(
            index_files(files)["work/index.md"],
            "# Ready now\n\n"
            "* [Second made](b-abcd.md) - A description.\n"
            "* [Third made](c-abcd.md) - A description.\n"
            "* [First made](a-abcd.md) - A description.\n"
            "* [No time](x-abcd.md) - A description.\n\n"
            "# Waiting on a decision\n\n* [Waiting](w-abcd.md) - A description.\n",
        )

    def test_work_cycle_blocks_the_work_index(self):
        files = {
            **BASE,
            "work/a-abcd.md": S.work(extra="depends_on: [work/b-abcd]\n"),
            "work/b-abcd.md": S.work(extra="depends_on: [work/a-abcd]\n"),
        }
        generated = indexes(files)
        self.assertNotIn("work/index.md", dict(generated.files))
        self.assertEqual(S.messages(generated.findings), ["depends_on forms a cycle: work/a-abcd, work/b-abcd"])

    def test_work_item_without_readiness(self):
        generated = indexes({**BASE, "work/a-abcd.md": S.work(readiness="later")})
        self.assertEqual(
            generated.findings,
            (pb.error("work/a-abcd.md", 9, "readiness must be ready or waiting to group the work item"),),
        )

    def test_files_that_cannot_be_listed(self):
        generated = indexes({**BASE, "research/a.md": "---\na: [\n---\n", "research/b.md": "---\ntitle: x\n---\n"})
        self.assertNotIn("research/index.md", dict(generated.files))
        self.assertEqual(
            generated.findings,
            (
                pb.error("research/a.md", 1, "an index cannot list this file until its frontmatter parses"),
                pb.error("research/b.md", 1, "an index cannot list this file until it has a type"),
            ),
        )

    def test_roles_pin_files_and_describe_directories(self):
        files = {
            **BASE,
            "work/w-abcd.md": S.work("Ship the API"),
            "implementation/w-abcd/zeta.md": S.doc("type: Note\ntitle: Zeta"),
            "implementation/w-abcd/notes.md": S.doc("type: Workflow Implementation Notes\ntitle: Notes"),
            "deliveries/w-abcd/pull-request.md": S.doc("type: Workflow Pull Request\ntitle: PR"),
            "deliveries/w-abcd/commit-message.md": S.doc("type: Workflow Commit Message\ntitle: Commit"),
            "deliveries/w-abcd/comments/abcd.md": S.doc("type: Workflow Inline Comment\ntitle: C"),
        }
        generated = index_files(files)
        self.assertEqual(
            generated["implementation/w-abcd/index.md"],
            "# Implementation notes\n\n* [Notes](notes.md)\n* [Zeta](zeta.md)\n",
        )
        self.assertEqual(
            generated["deliveries/w-abcd/index.md"],
            "# Delivery\n\n* [Commit](commit-message.md)\n* [PR](pull-request.md)\n\n"
            "# Directories\n\n* [comments](comments/) - Inline comments on the pull request\n",
        )
        self.assertEqual(generated["deliveries/index.md"], "# Directories\n\n* [w-abcd](w-abcd/) - Ship the API\n")
        self.assertEqual(generated["deliveries/w-abcd/comments/index.md"], "# Comments\n\n* [C](abcd.md)\n")
        self.assertEqual(generated["log/index.md"], "# Log entries\n\n* [Entry](2026-10-28-writing-plan-abcd.md) - A description.\n")

    def test_documents_run_newest_first_and_links_are_escaped(self):
        files = {
            **BASE,
            "research/old-abcd.md": S.doc(S.header("Workflow Research", "workflow/research/1", "Old [1]", at(1))),
            "research/new abcd.md": S.doc(S.header("Workflow Research", "workflow/research/1", "New", at(2))),
            "research/untimed.md": S.doc("type: Note\ndescription: '  spaced\n  out  '"),
        }
        self.assertEqual(
            index_files(files)["research/index.md"],
            "# Research\n\n* [New](new%20abcd.md) - A description.\n"
            "* [Old \\[1\\]](old-abcd.md) - A description.\n* [untimed](untimed.md) - spaced out\n",
        )

    def test_obsolete_indexes(self):
        files = {**BASE, "empty/index.md": "# X\n", "diagrams/index.md": "# X\n"}
        self.assertEqual(indexes(files, ("empty", "diagrams")).obsolete, ("empty/index.md",))

    def test_empty_directories_get_no_index(self):
        self.assertNotIn("empty/index.md", index_files(BASE, ("empty", "empty/deeper")))


class Log(unittest.TestCase):
    def test_log_groups_by_date_newest_first(self):
        files = {
            "log/2026-10-27-a-abcd.md": S.log_entry("First.\n\nMore.", at(27, 9), "Initialization"),
            "log/2026-10-28-b-abcd.md": S.log_entry("Second `x`\nline.", at(28, 9), "Update"),
            "log/2026-10-28-c-abcd.md": S.log_entry("Third.", at(28, 10)),
        }
        self.assertEqual(
            dict(pb.generate_log(S.bundle(files)).files)["log.md"],
            "# Directory Update Log\n\n## 2026-10-28\n"
            "* **Creation**: Third. [2026-10-28-c-abcd](/log/2026-10-28-c-abcd.md)\n"
            "* **Update**: Second line. [2026-10-28-b-abcd](/log/2026-10-28-b-abcd.md)\n\n"
            "## 2026-10-27\n* **Initialization**: First. [2026-10-27-a-abcd](/log/2026-10-27-a-abcd.md)\n",
        )

    def test_empty_log(self):
        self.assertEqual(pb.generate_log(S.bundle({})).files, (("log.md", "# Directory Update Log\n"),))

    def test_log_entry_problems(self):
        files = {
            "log/2026-10-28-a-abcd.md": "---\na: [\n---\n",
            "log/2026-10-28-b-abcd.md": S.log_entry("", "2026-10-28", " "),
        }
        generated = pb.generate_log(S.bundle(files))
        self.assertEqual(generated.files, (("log.md", "# Directory Update Log\n"),))
        self.assertEqual(
            generated.findings,
            (
                pb.error("log/2026-10-28-a-abcd.md", 1, "log.md cannot list this entry until its frontmatter parses"),
                pb.error("log/2026-10-28-b-abcd.md", 6, "generated.at must be a datetime with a UTC offset to date the log entry"),
                pb.error("log/2026-10-28-b-abcd.md", 8, "action must name the log entry's action"),
                pb.error("log/2026-10-28-b-abcd.md", 10, "the body must hold the one-sentence summary"),
            ),
        )

    def test_first_paragraph(self):
        self.assertEqual(pb.first_paragraph("\n\n  One\n two  \n\n| table |\n"), "One two")
        self.assertEqual(pb.first_paragraph("\n \n"), "")


class Stale(unittest.TestCase):
    def test_missing_stale_and_obsolete(self):
        files = {"work/w-abcd.md": S.work(), "work/index.md": "# Ready now\n\n* [Old](w-abcd.md)\n", "x/index.md": "# X\n"}
        loaded = S.bundle(files)
        self.assertEqual(
            pb.stale_findings(loaded, pb.generate_indexes(loaded), "index"),
            (
                pb.error("index.md", 1, "generated file is missing; run the index command with --write"),
                pb.error("work/index.md", 3, "generated file is stale; run the index command with --write"),
                pb.error("x/index.md", 1, "generated file is not needed in a directory without concept documents; run the index command with --write"),
            ),
        )

    def test_stale_line_past_the_end(self):
        loaded = S.bundle({"log.md": "# Directory Update Log"})
        generated = pb.Generated((("log.md", "# Directory Update Log\n"),), (), ())
        self.assertEqual(pb.stale_findings(loaded, generated, "log"), (pb.error("log.md", 2, "generated file is stale; run the log command with --write"),))

    def test_current_files_are_not_stale(self):
        files = S.with_generated(BASE)
        loaded = S.bundle(files)
        self.assertEqual(pb.stale_findings(loaded, pb.generate_indexes(loaded), "index"), ())


class Validate(unittest.TestCase):
    def test_complete_bundle_is_clean(self):
        files = S.with_generated(
            {
                **BASE,
                "decisions/d-abcd.md": S.decision(),
                "work/w-abcd.md": S.work(extra="waits_on: []\n"),
                "research/r-abcd.md": S.doc(
                    S.header("Workflow Research", "workflow/research/1", extra="sources:\n  - {resource: r, id: api}\n"),
                    "Claim.[^api]\n\n[^api]: R\n",
                ),
            }
        )
        found = pb.validate_bundle(S.bundle(files), S.registry())
        self.assertEqual(found, (pb.warning("research/r-abcd.md", 3, "schema file for workflow/research/1 not found"),))

    def test_generator_finding_is_dropped_for_a_failing_file(self):
        files = S.with_generated(BASE)
        files["decisions/d-abcd.md"] = S.decision("maybe")
        found = pb.validate_bundle(S.bundle(files), S.registry())
        self.assertEqual(S.lines(found, "state"), [("decisions/d-abcd.md", 7, "error")])
        self.assertEqual(S.messages(found, "decisions/index.md"), [])
        self.assertEqual(S.messages(found, "index.md"), ["generated file is stale; run the index command with --write"])

    def test_registry_and_bundle_findings_come_first_in_path_order(self):
        broken = S.schema_sources({**S.SHARED_SCHEMAS, "extra.schema.json": {"bogus": 1}})
        files = S.with_generated(BASE)
        found = pb.validate_bundle(S.bundle(files, links=("link.md",)), S.registry(broken))
        self.assertEqual(
            [(f.path, f.message) for f in found],
            [
                (f"{S.SCHEMA_DIRECTORY}/extra.schema.json", 'unsupported schema keyword "bogus"'),
                ("link.md", "symbolic link is not followed"),
            ],
        )


if __name__ == "__main__":
    unittest.main()
