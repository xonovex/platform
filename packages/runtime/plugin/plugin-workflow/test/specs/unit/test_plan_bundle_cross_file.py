"""Layer 3 of plan-bundle.py: rules that span files, file locations and footnotes."""

import unittest


import plan_bundle_support as S
pb = S.pb

OLD = "decisions/old-abcd"
NEW = "decisions/new-abcd"


def cross(files: dict[str, str]) -> list[tuple[str, int, str, str]]:
    return [(f.path, f.line, f.severity, f.message) for f in pb.cross_file_findings(S.bundle(files))]


def located(path: str, text: str) -> list[str]:
    return [f.message for f in pb.location_findings(pb.read_bundle_file(path, text))]


class Relations(unittest.TestCase):
    def test_relation_refers_to_its_own_file(self):
        files = {"work/w-abcd.md": S.work(extra="depends_on: [work/w-abcd]\nparent: work/w-abcd\n")}
        self.assertEqual(
            cross(files),
            [
                ("work/w-abcd.md", 10, "error", "depends_on refers to this file itself"),
                ("work/w-abcd.md", 11, "error", "parent refers to this file itself"),
            ],
        )

    def test_refinement_relation_keys_are_not_relations(self):
        files = {"work/w-abcd.md": S.work(extra="session: work/w-abcd\nfolded_into: [work/w-abcd]\n")}
        self.assertEqual(cross(files), [])

    def test_unknown_lists_are_not_checked(self):
        extra = "links:\n  - {service: s, key: K, fields: {title: ''}}\n  - {service: s, key: K}\n"
        self.assertEqual(cross({"work/w-abcd.md": S.work(extra=extra)}), [])

    def test_decision_relation_must_name_a_decision(self):
        files = {
            "work/w-abcd.md": S.work(),
            "verdicts/w-abcd.md": S.doc(S.header("Workflow Verdict", "workflow/verdict/1", extra="decision: work/w-abcd\n")),
        }
        self.assertEqual(
            cross(files),
            [("verdicts/w-abcd.md", 7, "error", "decision does not resolve to a Workflow Decision: work/w-abcd")],
        )


class Decisions(unittest.TestCase):
    def test_consistent_supersession_passes(self):
        files = {
            f"{OLD}.md": S.decision(extra=f"status: deprecated\nsuperseded_by: {NEW}\n"),
            f"{NEW}.md": S.decision(extra=f"supersedes: [{OLD}]\n"),
        }
        self.assertEqual(cross(files), [])

    def test_deprecated_without_replacement(self):
        self.assertEqual(
            cross({f"{OLD}.md": S.decision(extra="status: deprecated\n")}),
            [(f"{OLD}.md", 8, "error", "a deprecated decision names its replacement in superseded_by")],
        )

    def test_replacement_without_deprecation(self):
        files = {f"{OLD}.md": S.decision(extra=f"superseded_by: {NEW}\n"), f"{NEW}.md": S.decision(extra=f"supersedes: [{OLD}]\n")}
        found = cross(files)
        self.assertIn((f"{OLD}.md", 8, "error", "superseded_by is set but status is not deprecated"), found)
        self.assertIn((f"{NEW}.md", 8, "error", f"supersedes {OLD}, which is not deprecated"), found)

    def test_replacement_is_not_a_decision(self):
        files = {f"{OLD}.md": S.decision(extra="status: deprecated\nsuperseded_by: work/w-abcd\n"), "work/w-abcd.md": S.work()}
        self.assertEqual(cross(files), [(f"{OLD}.md", 9, "error", "superseded_by does not resolve to a decision: work/w-abcd")])

    def test_replacement_does_not_list_it(self):
        files = {f"{OLD}.md": S.decision(extra=f"status: deprecated\nsuperseded_by: {NEW}\n"), f"{NEW}.md": S.decision()}
        self.assertEqual(cross(files), [(f"{OLD}.md", 9, "error", f"{NEW} does not list this decision in supersedes")])

    def test_supersedes_a_file_that_is_no_decision(self):
        files = {f"{NEW}.md": S.decision(extra="supersedes: [work/w-abcd]\n"), "work/w-abcd.md": S.work()}
        self.assertEqual(cross(files), [(f"{NEW}.md", 8, "error", "supersedes does not resolve to a decision: work/w-abcd")])

    def test_supersedes_a_decision_replaced_by_another(self):
        files = {
            f"{OLD}.md": S.decision(extra="status: deprecated\nsuperseded_by: decisions/third-abcd\n"),
            f"{NEW}.md": S.decision(extra=f"supersedes: [{OLD}]\n"),
            "decisions/third-abcd.md": S.decision(extra=f"supersedes: [{OLD}]\n"),
        }
        self.assertEqual(cross(files), [(f"{NEW}.md", 8, "error", f"supersedes {OLD}, whose superseded_by is not this decision")])

    def test_unresolved_targets_are_left_to_the_schema(self):
        self.assertEqual(cross({f"{NEW}.md": S.decision(extra="supersedes: [decisions/none-abcd]\n")}), [])


class WorkItems(unittest.TestCase):
    def test_waiting_without_waits_on(self):
        self.assertEqual(
            cross({"work/w-abcd.md": S.work(readiness="waiting", extra="waits_on: []\n")}),
            [("work/w-abcd.md", 9, "error", "readiness is waiting but waits_on names no decision")],
        )

    def test_ready_with_waits_on_is_a_warning(self):
        files = {"work/w-abcd.md": S.work(extra=f"waits_on: [{OLD}]\n"), f"{OLD}.md": S.decision()}
        self.assertEqual(cross(files), [("work/w-abcd.md", 10, "warning", "readiness is ready but waits_on still names decisions")])

    def test_parent_must_be_a_work_package(self):
        files = {
            "work/w-abcd.md": S.work(extra="parent: work/other-abcd\n"),
            "work/other-abcd.md": S.work(),
            "work/p-abcd.md": S.work(kind="work-package"),
            "work/ok-abcd.md": S.work(extra="parent: work/p-abcd\n"),
        }
        self.assertEqual(cross(files), [("work/w-abcd.md", 10, "error", "parent is not a work package: work/other-abcd")])

    def test_dependency_cycle(self):
        files = {
            "work/a-abcd.md": S.work(extra="depends_on: [work/b-abcd]\n"),
            "work/b-abcd.md": S.work(extra="depends_on: [work/c-abcd]\n"),
            "work/c-abcd.md": S.work(extra="depends_on: [work/a-abcd, work/d-abcd]\n"),
            "work/d-abcd.md": S.work(),
        }
        self.assertEqual(
            cross(files),
            [("work/a-abcd.md", 10, "error", "depends_on forms a cycle: work/a-abcd, work/b-abcd, work/c-abcd")],
        )

    def test_unique_ids(self):
        extra = "criteria:\n  - {id: c1, text: a}\n  - {id: c1, text: b}\nsources:\n  - {resource: r, id: s}\n  - {resource: r, id: s}\n"
        found = cross({"work/w-abcd.md": S.work(extra=extra)})
        self.assertIn(("work/w-abcd.md", 12, "error", "id c1 is already used by criteria[0]"), found)
        self.assertIn(("work/w-abcd.md", 15, "error", "id s is already used by sources[0]"), found)


class Footnotes(unittest.TestCase):
    def file(self, sources: str, body: str) -> dict[str, str]:
        return {"research/r-abcd.md": S.doc(S.header("Workflow Research", "workflow/research/1", extra=f"sources:\n{sources}"), body)}

    def test_cited_and_defined_source_passes(self):
        files = self.file("  - {resource: r, id: api}\n", "Claim.[^api]\n\n[^api]: Reference\n")
        self.assertEqual(cross(files), [])

    def test_footnote_without_source(self):
        files = self.file("  - {resource: r, id: api}\n", "Claim.[^api][^other]\n\n[^api]: R\n[^other]: R\n")
        self.assertEqual(
            cross(files),
            [
                ("research/r-abcd.md", 10, "error", "footnote [^other] matches no sources id"),
                ("research/r-abcd.md", 13, "error", "footnote [^other] matches no sources id"),
            ],
        )

    def test_footnote_without_definition(self):
        files = self.file("  - {resource: r, id: api}\n", "Claim.[^api]\n")
        self.assertEqual(cross(files), [("research/r-abcd.md", 10, "error", "footnote [^api] has no definition")])

    def test_source_never_cited(self):
        files = self.file("  - {resource: r}\n  - {resource: r, id: api}\n", "No claim.\n")
        self.assertEqual(cross(files), [("research/r-abcd.md", 9, "warning", "source api is never cited in the body")])

    def test_code_is_not_prose(self):
        files = self.file("  - {resource: r, id: api}\n", "Claim.[^api] `[^x]`\n\n```\n[^y]\n```\n\n[^api]: R\n")
        self.assertEqual(cross(files), [])

    def test_prose_lines_blank_fences_and_spans(self):
        self.assertEqual(pb.prose_lines("a `b` c\n~~~\nx\n~~~~\nd"), ("a     c", "", "", "", "d"))


class Locations(unittest.TestCase):
    def test_fixed_file_in_the_wrong_directory(self):
        self.assertEqual(located("work/plan.md", S.plan()), ["a Workflow Plan belongs at plan.md"])

    def test_fixed_file_with_the_wrong_name(self):
        self.assertEqual(located("plan2.md", S.plan()), ["a Workflow Plan belongs at plan.md"])

    def test_type_in_the_wrong_directory(self):
        self.assertEqual(located("research/w-abcd.md", S.work()), ["a Workflow Work Item belongs at work/"])

    def test_slug_suffix_name(self):
        rule = pb.SLUG_SUFFIX_RULE
        self.assertEqual(located("work/Bad-abcd.md", S.work()), [f"file name Bad-abcd.md does not match {rule}"])
        self.assertEqual(located("work/w-abc1.md", S.work()), [f"file name w-abc1.md does not match {rule}"])
        self.assertEqual(located(f"work/{'a' * 49}-abcd.md", S.work()), [f"file name {'a' * 49}-abcd.md does not match {rule}"])
        self.assertEqual(located(f"work/{'a' * 48}-abcd.md", S.work()), [])

    def test_log_name(self):
        self.assertEqual(located("log/2026-10-28-writing-plan-abcd.md", S.log_entry()), [])
        self.assertEqual(
            located("log/2026-02-30-writing-plan-abcd.md", S.log_entry()),
            ["file name 2026-02-30-writing-plan-abcd.md does not match <YYYY-MM-DD>-<skill>-<suffix>"],
        )

    def test_work_item_directory(self):
        notes = S.doc(S.header("Workflow Implementation Notes", "workflow/implementation-notes/1"))
        self.assertEqual(located("implementation/w-abcd/notes.md", notes), [])
        self.assertEqual(located("implementation/bad/notes.md", notes), [f"work item directory bad does not match {pb.SLUG_SUFFIX_RULE}"])
        self.assertEqual(
            located("implementation/w-abcd/other.md", notes),
            ["a Workflow Implementation Notes belongs at implementation/<work-id>/notes.md"],
        )

    def test_inline_comment_name(self):
        comment = S.doc(S.header("Workflow Inline Comment", "workflow/inline-comment/1"))
        self.assertEqual(located("deliveries/w-abcd/comments/abcd.md", comment), [])
        self.assertEqual(
            located("deliveries/w-abcd/comments/abc1.md", comment),
            ["file name abc1.md does not match <suffix>: 4 characters from a-z and 2-7"],
        )

    def test_removed_types_have_no_location(self):
        for kind, path in (
            ("Workflow Refinement Session", "refinement/2026-10-28-x/pack.md"),
            ("Workflow Outcome", "refinement/bad/outcomes/abcd.md"),
            ("Workflow Sync Operation", "operations/run/op.md"),
        ):
            with self.subTest(kind):
                self.assertEqual(located(path, S.doc(S.header(kind, "workflow/x/1"))), [])


if __name__ == "__main__":
    unittest.main()
