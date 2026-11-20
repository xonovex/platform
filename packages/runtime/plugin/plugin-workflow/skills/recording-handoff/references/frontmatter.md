# Frontmatter contract

## Contents

- The common header
- Writing the header
- Actors and the human rule
- Timestamps
- Quoting string values
- Sources and footnotes
- Relation keys
- Status mapping
- Preserving unknown keys
- Rules the schemas cannot express

## The common header

Every Workflow document starts with the keys in `assets/schemas/common.schema.json`, and each type schema extends them. A schema in another skill extends the header by its schema id, `{"allOf": [{"$ref": "workflow/common#/$defs/header"}]}`; [validator.md](validator.md) lists the reference forms.

| Key | Required | Value |
| --- | --- | --- |
| `type` | yes | `Workflow <Name>`, exactly as in the type table |
| `schema` | yes | `workflow/<kebab-name>/<integer>`, for example `workflow/decision/1` |
| `title` | yes | A non-empty display name; it is the link text in generated indexes |
| `description` | yes | One sentence; it follows the title in generated indexes |
| `generated` | yes | `{ by, at }` |
| `verified` | no | A list of `{ by, at }`, or one bare mapping |
| `status` | no | `draft`, `stable` or `deprecated`; absent means stable |
| `stale_after` | no | A timestamp after which the content is stale |
| `sources` | no | A list of source entries |
| `tags` | no | A list of short strings |
| `resource` | no | One URI of the underlying asset |
| `skills_to_consult` | no | Installed skill names; on a Workflow Plan, a Workflow Work Item and a Workflow Research |

There is no `id` key: the file path is the identity.

## Writing the header

Write frontmatter in canonical block form: block mappings and block sequences, with a flow mapping only for `{ by, at }`. Order the keys as in the table, then the type-specific keys. The validator parses a strict YAML subset, so anchors, aliases, tags and multiple documents are errors; use none of them.

```yaml
generated: {by: settling-questions/0.1.0, at: 2026-10-27T09:30:00+01:00}
verified:
  - {by: human:jdoe, at: 2026-10-27T10:05:00+01:00}
```

## Actors and the human rule

An actor is `<producer>/<version>` for a skill or tool (`settling-questions/0.1.0`), `human:<id>` for a person, or `process:<id>` for an automated process (`process:nightly-sync`). For a skill, `<version>` is the `version` field of the `plugin.json` at the root of the plugin that ships the skill (the directory that holds `skills/`): read it, never guess it.

Readers derive trust from the `human:` prefix, so it is reserved:

- Content a person wrote by hand has `generated.by` of the form `human:<id>`.
- Content a skill drafted and a person confirmed keeps the skill as `generated.by`, and gains a `verified` entry with the person as `human:<id>`. The id comes from the invocation: the person says who they are or gives the id. When it is missing and the skill needs it, ask for it, and never invent one.
- A skill never adds a `verified` entry for content it wrote itself, and never writes `human:` for a person who has not confirmed the content in this session.
- An automated check records itself as `process:<id>`.

## Timestamps

Every timestamp (`generated.at`, `verified[].at`, `last_synced`) comes from the system clock: run `date -u +%Y-%m-%dT%H:%M:%SZ` and use its output. Never invent, round or copy a time from another file.

Every timestamp is ISO 8601 with an explicit UTC offset: `2026-10-27T09:30:00+01:00` or `2026-10-27T08:30:00Z`; the command above gives the `Z` form. A date alone or a datetime without an offset fails validation. `generated.at` marks the last meaningful change of the content, so every edit that changes content updates it; `verified[].at` marks a confirmation and is independent of it.

## Quoting string values

A string value that contains a colon followed by a space, a quote, or starts with a special character (`[`, `{`, `*`, `&`, `!`, `|`, `>`, `%`, `@`, a backtick or a leading `-`, `?` or `:` followed by a space) is written in single quotes, with a quote inside it doubled. This matters most for free text such as a log entry's `authorization` and for titles and descriptions: `authorization: 'The invoker said: go ahead'`. The validator rejects an unquoted value it cannot parse.

## Sources and footnotes

Each `sources` entry needs `resource`: an absolute URL, a bundle-relative path starting with `/`, a relative path, or a scope descriptor string. `id` (kebab-case, unique in the file), `title`, `author` (an actor), `last_modified` (a timestamp) and `usage_count` (an integer) are optional.

A body attributes a claim with a footnote whose label is the source `id`. The label, not a list position, is the join key, because a reordered list would silently misattribute:

```markdown
The service exposes one endpoint for each region.[^region-api]

[^region-api]: Regional API reference
```

Give a source an `id` whenever the body cites it, and start the footnote definition text with the source `title`. The definition may add what the source supports, such as the line ranges of a cited file.

## Relation keys

Every relation key holds concept IDs: the path inside the bundle without `.md` and without a leading slash, such as `decisions/target-identifier-7k2f`. The relation keys are `depends_on`, `blocks`, `supersedes`, `superseded_by`, `waits_on`, `decisions`, `parent`, `decision` and `touched`. Use the exact key from the type schema, and never a file name, a title or a display number. Body links use the bundle-relative form with `.md`, for example `/decisions/target-identifier-7k2f.md`.

## Status mapping

`status` is the lifecycle of the document. It is separate from domain fields such as a decision's `state` and a work item's `readiness`.

| Status | Meaning here |
| --- | --- |
| `draft` | A skill wrote the content and the invoker has not confirmed it. Used on files that are edited in place: the plan and work items |
| `stable` (or absent) | Confirmed. Every append-only file is written with this value, because it is written only after confirmation and cannot be edited into it later |
| `deprecated` | No longer current. A decision gets it only when another decision supersedes it, together with `superseded_by` and nothing else changed. Other files get it only through the skill that owns them, with a log entry |

When a decision supersedes another, write the new file with `supersedes`, then set `status: deprecated` and `superseded_by` on the old one. The old file keeps its `generated`, so its position in generated lists stays put.

## Preserving unknown keys

A key that a schema does not name is valid. Keep it, with its value and its position, whenever you edit a file; a producer that drops keys breaks the next consumer. Add keys only through a schema change in the skill that owns the type.

## Rules the schemas cannot express

The schema vocabulary has no conditionals on absent keys, no `not` and no object-value checks. These rules hold anyway, and the validator checks them in code:

- `superseded_by` is present only when `status` is `deprecated`. The schema covers the other direction: `deprecated` requires `superseded_by`.
- A deprecated decision's `superseded_by` target lists it in `supersedes`.
- No relation key refers to its own file, and `depends_on` between work files has no cycle.
- `parent` points at a file of kind `work-package`. `decision` points at a Workflow Decision. A resolving path is not enough.
- `criteria[].id` is unique in the file, and `source` ids are unique in the file.
- Every footnote label used in the body matches a source `id`, and every source cited has a footnote definition.
- `readiness: ready` carries no `waits_on`.
- The file name matches `<slug>-<suffix>`, the directory matches the type, and `schema` names the integer the type schema declares.
- Generated files are current: see generated-files.md.
