"""Layer 2 of plan-bundle.py: the schema registry and the schema check of Workflow documents."""

import unittest


import plan_bundle_support as S
pb = S.pb
DIR = S.SCHEMA_DIRECTORY


def registry_messages(schemas: dict[str, object]) -> list[tuple[str, int, str]]:
    found = S.registry(S.schema_sources({**S.SHARED_SCHEMAS, **schemas})).findings
    return [(f.path, f.line, f.message) for f in found]


def node_problem(node: object) -> list[str]:
    """The registry findings of one extra schema file holding node."""
    return [message for path, _, message in registry_messages({"extra.schema.json": node}) if path.endswith("extra.schema.json")]


def schema_messages(files: dict[str, str], registry=None) -> list[tuple[str, int, str, str]]:
    loaded = S.bundle(files)
    found = pb.schema_findings(loaded, registry or S.registry())
    return [(f.path, f.line, f.severity, f.message) for f in found]


class Registry(unittest.TestCase):
    def test_shared_schemas_load_clean(self):
        registry = S.registry()
        self.assertEqual(registry.findings, ())
        self.assertEqual(
            sorted(registry.by_id),
            ["workflow/common", "workflow/decision/1", "workflow/log-entry/1", "workflow/plan/1", "workflow/work-item/1"],
        )

    def test_invalid_json(self):
        found = registry_messages({"bad.schema.json": '{\n  "a": 1,\n  oops\n}\n'})
        self.assertEqual(len(found), 1)
        path, line, message = found[0]
        self.assertEqual((path, line), (f"{DIR}/bad.schema.json", 3))
        self.assertTrue(message.startswith("invalid JSON: "))

    def test_non_json_constants_are_invalid(self):
        self.assertTrue(registry_messages({"bad.schema.json": '{"a": NaN}'})[0][2].startswith("invalid JSON: "))

    def test_duplicate_id_in_one_tier(self):
        self.assertEqual(
            registry_messages({"twin.schema.json": '{\n  "$id": "workflow/plan/1"\n}'}),
            [(f"{DIR}/twin.schema.json", 2, f"schema workflow/plan/1 is also declared by {DIR}/plan.schema.json")],
        )

    def test_duplicate_id_by_const(self):
        found = registry_messages({"zz.schema.json": {"properties": {"schema": {"const": "workflow/plan/1"}}}})
        self.assertEqual([m for _, _, m in found], [f"schema workflow/plan/1 is also declared by {DIR}/plan.schema.json"])

    def test_extra_directory_wins_over_a_default(self):
        override = S.schema_sources({"plan.schema.json": {"$id": "workflow/plan/1", "required": ["owners"]}}, "/extra", 0)
        registry = S.registry(override, S.schema_sources())
        self.assertEqual(registry.findings, ())
        self.assertEqual(registry.by_id["workflow/plan/1"].file, "/extra/plan.schema.json")

    def test_same_file_twice_loads_once(self):
        registry = S.registry(S.schema_sources(), S.schema_sources())
        self.assertEqual(registry.findings, ())
        self.assertEqual(len(registry.files), 5)

    def test_unsupported_keyword(self):
        self.assertEqual(node_problem({"not": {}}), ['unsupported schema keyword "not"'])

    def test_schema_must_be_an_object(self):
        self.assertEqual(node_problem({"items": True}), ["a schema must be a JSON object"])

    def test_keyword_values(self):
        cases = {
            "type": ({"type": "date"}, "type must name string, integer, number, boolean, array, object, null"),
            "empty type": ({"type": []}, "type must name string, integer, number, boolean, array, object, null"),
            "required": ({"required": [1]}, "required must be an array of key names"),
            "properties": ({"properties": []}, "must be an object of schemas"),
            "allOf": ({"allOf": []}, "must be a non-empty array of schemas"),
            "oneOf": ({"oneOf": {}}, "must be a non-empty array of schemas"),
            "enum": ({"enum": []}, "enum must be a non-empty array"),
            "pattern": ({"pattern": "("}, "pattern is not a valid regular expression: ("),
            "pattern type": ({"pattern": 5}, "pattern must be a string"),
            "minLength": ({"minLength": -1}, "minLength must be a non-negative integer"),
            "minItems": ({"minItems": 1.5}, "minItems must be a non-negative integer"),
            "uniqueItems": ({"uniqueItems": 1}, "uniqueItems must be true or false"),
            "additionalProperties": ({"additionalProperties": {}}, "additionalProperties must be true or false"),
            "format": ({"format": "date"}, 'format supports only "datetime-offset"'),
            "x-concept": ({"x-concept": "work"}, 'x-concept is "" or a directory prefix ending in /, such as "decisions/"'),
            "$ref": ({"$ref": "#/definitions/x"}, "$ref must be #/$defs/<name>, <file>.schema.json#/$defs/<name> or <schema-id>#/$defs/<name>"),
            "$id": ({"$id": 3}, "$id must be a string"),
            "missing ref": ({"$ref": "#/$defs/none"}, "$ref target #/$defs/none does not exist"),
        }
        for label, (node, message) in cases.items():
            with self.subTest(label):
                self.assertEqual(node_problem(node), [message])

    def test_annotations_and_integral_float_pass(self):
        self.assertEqual(node_problem({"title": "x", "$comment": "c", "examples": [], "minLength": 1.0}), [])

    def test_finding_line_points_at_the_keyword(self):
        found = registry_messages({"extra.schema.json": '{\n  "title": "x",\n  "bogus": 1\n}'})
        self.assertEqual(found, [(f"{DIR}/extra.schema.json", 3, 'unsupported schema keyword "bogus"')])

    def test_references_resolve_by_sibling_file_and_by_id(self):
        sibling = {"$ref": "common.schema.json#/$defs/actor"}
        by_id = {"$ref": "workflow/common#/$defs/actor"}
        self.assertEqual(node_problem(sibling), [])
        self.assertEqual(node_problem(by_id), [])
        other = S.schema_sources({"x.schema.json": {"$ref": "common.schema.json#/$defs/actor"}}, "/other", 1)
        self.assertEqual(
            [f.message for f in S.registry(S.schema_sources(), other).findings],
            ["$ref target common.schema.json#/$defs/actor does not exist"],
        )


class DocumentSchemas(unittest.TestCase):
    def test_valid_documents_pass(self):
        files = {
            "plan.md": S.plan(),
            "decisions/a-decision-abcd.md": S.decision(),
            "work/a-work-item-abcd.md": S.work(extra="criteria:\n  - {id: c1, text: t}\n"),
            "log/2026-10-28-writing-plan-abcd.md": S.log_entry(extra="touched: [plan]\n"),
        }
        self.assertEqual(schema_messages(files), [])

    def test_schema_must_name_a_versioned_schema(self):
        self.assertEqual(
            schema_messages({"plan.md": S.doc("type: Workflow Plan\nschema: workflow/plan")}),
            [("plan.md", 3, "error", 'schema must name workflow/<kebab-name>/<integer>, found "workflow/plan"')],
        )
        self.assertEqual(
            schema_messages({"plan.md": S.doc("title: x\ntype: Workflow Plan")}),
            [("plan.md", 3, "error", "schema must name workflow/<kebab-name>/<integer>, found nothing")],
        )
        self.assertEqual(
            schema_messages({"plan.md": S.doc("type: Workflow Plan\nschema: {v: 1.0}")})[0][3],
            'schema must name workflow/<kebab-name>/<integer>, found {"v":1}',
        )

    def test_missing_schema_is_a_warning(self):
        self.assertEqual(
            schema_messages({"research/r-abcd.md": S.doc(S.header("Workflow Research", "workflow/research/1"))}),
            [("research/r-abcd.md", 3, "warning", "schema file for workflow/research/1 not found")],
        )

    def test_broken_schema_is_not_applied(self):
        broken = S.registry(S.schema_sources({**S.SHARED_SCHEMAS, "common.schema.json": {**S.COMMON, "bogus": 1}}))
        self.assertEqual(
            schema_messages({"plan.md": S.plan()}, broken),
            [("plan.md", 3, "warning", "schema file for workflow/plan/1 has errors and was not applied")],
        )

    def test_non_workflow_documents_are_not_checked(self):
        self.assertEqual(schema_messages({"notes.md": S.doc("type: Note\nschema: nope")}), [])

    def test_unknown_keys_pass(self):
        self.assertEqual(schema_messages({"plan.md": S.plan(extra="extra_key: {nested: [1]}\n")}), [])

    def test_missing_required_key(self):
        self.assertEqual(
            schema_messages({"decisions/d-abcd.md": S.doc(S.header("Workflow Decision", "workflow/decision/1"))}),
            [("decisions/d-abcd.md", 2, "error", "missing required key state")],
        )

    def test_type_mismatch(self):
        self.assertEqual(
            schema_messages({"plan.md": S.plan(extra="tags: x\n")}),
            [("plan.md", 7, "error", "tags: expected array, got string")],
        )

    def test_const(self):
        files = {"plan.md": S.doc(S.header("Workflow Plan", "workflow/plan/1").replace("type: Workflow Plan", "type: Workflow  Plan"))}
        self.assertIn(("plan.md", 2, "error", 'type: must be "Workflow Plan"'), schema_messages(files))

    def test_enum(self):
        self.assertEqual(
            schema_messages({"decisions/d-abcd.md": S.decision(state="maybe")}),
            [("decisions/d-abcd.md", 7, "error", 'state: must be one of "settled", "open", "paused", "partly-settled"')],
        )

    def test_pattern(self):
        self.assertEqual(
            schema_messages({"plan.md": S.plan().replace("tester/1.0.0", "nobody")}),
            [
                (
                    "plan.md",
                    6,
                    "error",
                    'generated.by: "nobody" does not match the pattern ^(human|process):[A-Za-z0-9][A-Za-z0-9._-]*$|^[A-Za-z][A-Za-z0-9._-]*/[A-Za-z0-9][A-Za-z0-9._+-]*$',
                )
            ],
        )

    def test_pattern_dollar_does_not_match_before_a_final_line_break(self):
        found = schema_messages({"plan.md": S.plan(extra='sources:\n  - {resource: r, id: "a\\n"}\n')})
        self.assertEqual(found, [("plan.md", 8, "error", 'sources[0].id: "a\\n" does not match the pattern ^[a-z0-9]+(-[a-z0-9]+)*$')])

    def test_min_length(self):
        self.assertEqual(
            schema_messages({"plan.md": S.plan().replace("title: The plan", "title: ''")}),
            [("plan.md", 4, "error", "title: must not be empty")],
        )
        extra = S.schema_sources({"x.schema.json": {"$id": "workflow/x/1", "properties": {"t": {"minLength": 3}}}}, "/x", 0)
        found = schema_messages({"x.md": S.doc("type: Workflow X\nschema: workflow/x/1\nt: ab")}, S.registry(extra))
        self.assertEqual(found, [("x.md", 4, "error", "t: must be at least 3 characters long")])

    def test_datetime_offset(self):
        self.assertEqual(
            schema_messages({"plan.md": S.plan(at="2026-10-28T09:00:00")}),
            [("plan.md", 6, "error", 'generated.at: "2026-10-28T09:00:00" is not an ISO 8601 datetime with an explicit UTC offset')],
        )

    def test_concept_id_problems(self):
        files = {
            "decisions/kept-abcd.md": S.decision(),
            "work/w-abcd.md": S.work(
                extra="depends_on:\n  - work/none-abcd\n  - decisions/kept-abcd\n  - work/w-abcd.md\n  - /work/w-abcd\n  - 3\n"
            ),
        }
        found = [message for *_, message in schema_messages(files)]
        self.assertEqual(
            found,
            [
                "depends_on[0]: does not resolve to a file in the bundle: work/none-abcd",
                "depends_on[1]: must be a concept in work/: decisions/kept-abcd",
                'depends_on[2]: "work/w-abcd.md" does not match the pattern ^[a-z0-9][a-z0-9-]*(/[a-z0-9][a-z0-9-]*)*$',
                "depends_on[2]: is a file name, a concept ID has no .md: work/w-abcd.md",
                'depends_on[3]: "/work/w-abcd" does not match the pattern ^[a-z0-9][a-z0-9-]*(/[a-z0-9][a-z0-9-]*)*$',
                "depends_on[3]: is not a concept ID, a bundle path without a leading / or .md: /work/w-abcd",
                "depends_on[4]: expected string, got integer",
                "depends_on[4]: must be a concept ID, got integer",
            ],
        )

    def test_min_items_and_unique_items(self):
        found = schema_messages({"plan.md": S.plan(extra="verified: []\ntags: [a, b, a]\n")})
        self.assertEqual(
            [m for *_, m in found],
            [
                "verified: matches none of the allowed shapes (shape 1: expected object, got array; shape 2: must not be empty)",
                "tags[2]: repeats item [0]",
            ],
        )
        extra = S.schema_sources({"x.schema.json": {"$id": "workflow/x/1", "properties": {"l": {"minItems": 2}}}}, "/x", 0)
        found = schema_messages({"x.md": S.doc("type: Workflow X\nschema: workflow/x/1\nl: [a]")}, S.registry(extra))
        self.assertEqual([m for *_, m in found], ["l: must have at least 2 items"])

    def test_additional_properties(self):
        extra = S.schema_sources(
            {"x.schema.json": {"$id": "workflow/x/1", "properties": {"type": {}, "schema": {}}, "additionalProperties": False}}, "/x", 0
        )
        found = schema_messages({"x.md": S.doc("type: Workflow X\nschema: workflow/x/1\n'2': a\nb: c")}, S.registry(extra))
        self.assertEqual(found, [("x.md", 4, "error", "2: is not an allowed key"), ("x.md", 5, "error", "b: is not an allowed key")])

    def test_one_of_matching_more_than_one_shape(self):
        extra = S.schema_sources({"x.schema.json": {"$id": "workflow/x/1", "oneOf": [{"type": "object"}, {"required": ["type"]}]}}, "/x", 0)
        found = schema_messages({"x.md": S.doc("type: Workflow X\nschema: workflow/x/1")}, S.registry(extra))
        self.assertEqual(found, [("x.md", 2, "error", "matches more than one allowed shape")])

    def test_one_of_shape_text_is_relative_to_the_value(self):
        found = schema_messages({"plan.md": S.plan(extra="verified: {by: nobody, at: 2026-10-28T09:00:00Z}\n")})
        self.assertEqual(
            found[0][3],
            'verified: matches none of the allowed shapes (shape 1: by: "nobody" does not match the pattern '
            "^(human|process):[A-Za-z0-9][A-Za-z0-9._-]*$|^[A-Za-z][A-Za-z0-9._-]*/[A-Za-z0-9][A-Za-z0-9._+-]*$; "
            "shape 2: expected array, got object)",
        )

    def test_reference_loop(self):
        extra = S.schema_sources({"x.schema.json": {"$id": "workflow/x/1", "$defs": {"a": {"$ref": "#/$defs/a"}}, "$ref": "#/$defs/a"}}, "/x", 0)
        found = schema_messages({"x.md": S.doc("type: Workflow X\nschema: workflow/x/1")}, S.registry(extra))
        self.assertEqual(found, [("x.md", 2, "error", "schema references #/$defs/a in a loop")])

    def test_integer_type_accepts_an_integral_float(self):
        extra = S.schema_sources({"x.schema.json": {"$id": "workflow/x/1", "properties": {"n": {"type": "integer"}, "f": {"type": "integer"}}}}, "/x", 0)
        found = schema_messages({"x.md": S.doc("type: Workflow X\nschema: workflow/x/1\nn: 1.0\nf: 1.5")}, S.registry(extra))
        self.assertEqual(found, [("x.md", 5, "error", "f: expected integer, got number")])

    def test_issue_path_text(self):
        self.assertEqual(pb.issue_text(pb.SchemaIssue(("criteria", 0, "id"), "m")), "criteria[0].id: m")
        self.assertEqual(pb.issue_text(pb.SchemaIssue((), "m")), "m")


if __name__ == "__main__":
    unittest.main()
