# The validator

## Contents

- Run it
- The three layers
- Writing a schema
- Output and exit codes
- Generators
- Without the script

## Run it

`scripts/plan-bundle.py` is one Python file for Python 3.12 or later that uses only the standard library. Run it with `uv run`, which reads its inline metadata, and give it a command and a bundle or family directory:

```bash
uv run scripts/plan-bundle.py validate plans/<change>
uv run scripts/plan-bundle.py validate plans/<change> --format json
uv run scripts/plan-bundle.py validate plans/<change> --schemas path/to/extra/schemas
```

- `--schemas <dir>` adds a schema directory and can repeat. By default it loads `assets/schemas` of every skill directory next to this skill, so a schema in an owner skill needs no registration.
- `--format text|json` selects the output. Text is the default.
- A family directory holds a `roadmap.md` and no `plan.md` or `log/`. Given one, each command runs on every direct subdirectory that holds a `log/` directory, in name order, and ignores the roadmap and every other entry. Each finding path starts with the name of its bundle, and the exit code is the worst of the bundles.

## The three layers

1. **Open Knowledge Format conformance.** Every `.md` file that is not `index.md` or `log.md` has parseable frontmatter and a non-empty `type`. `index.md` and `log.md` follow the structure of an index and a log. The frontmatter parser accepts block and flow mappings and sequences, plain and quoted scalars, block scalars and comments. It rejects anchors, aliases, tags and multiple documents with `unsupported YAML construct: <name>` instead of guessing.
2. **Schemas.** A file whose `type` starts with `Workflow ` is checked against the schema its `schema` key names. Unknown keys pass, and so does a document of an unknown type outside the `Workflow ` family. A Workflow document whose schema file cannot be found is a warning, not an error.
3. **Cross-file rules.** Every concept ID in a relation key resolves to a file. A deprecated decision has `superseded_by`, and its target lists it in `supersedes`. Work `depends_on` is acyclic. `readiness: waiting` has a non-empty `waits_on`. File names match `<slug>-<suffix>`. The generated files are current.

## Writing a schema

A schema is a `*.schema.json` file in an `assets/schemas` directory. A type schema names the document schema it validates in `$id` or as the `const` of its `schema` property, for example `workflow/decision/1`. A `$ref` takes one of three forms, resolved in this order:

- `#/$defs/<name>`: a definition in the same file.
- `<file>.schema.json#/$defs/<name>`: a definition in a file in the same directory, as the shared type schemas use `common.schema.json#/$defs/header`.
- `<schema-id>#/$defs/<name>`: a definition in any loaded schema whose `$id` is `<schema-id>`. This is how an owner skill extends the shared header without a copy of it: `{"allOf": [{"$ref": "workflow/common#/$defs/header"}]}`.

A `$ref` that does not resolve is an error in the schema file, and documents of that schema are not checked until it is fixed. When a `--schemas` directory and a default directory both hold a schema with the same `$id`, the `--schemas` one wins; the same `$id` twice within the defaults, or twice within the `--schemas` directories, is an error. The `$id` of `common.schema.json` is `workflow/common`, which no document can name, because a document's `schema` always ends in an integer.

## Output and exit codes

One finding per line, then a summary line:

```text
decisions/target-identifier-7k2f.md:9: error: superseded_by does not resolve to a decision: decisions/old-name-aaaa
work/pricing-api-q3mn.md:1: warning: schema file for workflow/work-item/2 not found
2 findings: 1 error, 1 warning
```

- Exit 0: no findings, or warnings only.
- Exit 1: at least one error.
- Exit 2: usage error, such as an unknown command or a missing directory.

Fix errors in the file that carries them. For an error about a generated file, run the generator rather than editing the file.

## Generators

`index` and `log` take `--check` or `--write`. `--check` reports a stale or missing generated file and exits 1 when one is stale. `--write` rewrites the files. The output depends on file contents only, so the same bundle always produces the same bytes. The grouping and ordering rules are in [generated-files.md](generated-files.md).

## Without the script

When the script is not available, for example a machine without Python or `uv`, or a copy of the skill without `scripts/`, say so in the entry report and do not claim the bundle is valid.

- Check the files the step reads by hand against [frontmatter.md](frontmatter.md) and the schemas in `assets/schemas/`: the common header, the type's required keys, timestamps with offsets, concept IDs that resolve, the file name pattern. Report what you find in the same `path:line: severity: message` form.
- Leave `index.md` and `log.md` untouched and report them as not regenerated. Read the directory and the frontmatter of the files directly instead of the index.
- Never write a generated file by hand to make a check pass.
