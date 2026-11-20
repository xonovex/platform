# The close-out file

## Contents

- Where the file lives
- Frontmatter
- Body
- The result
- Writing and closing the run
- Closing a package again

## Where the file lives

`close-out.md` at the root of the change's plan directory, with `type: Workflow Close-out Record` and `schema: workflow/close-out/1`. Its concept ID is `close-out`. The file has no slug and no suffix, which is the one naming exception besides `plan.md`.

## Frontmatter

The common header, the timestamps and the version in `generated.by` follow the frontmatter rules of **recording-handoff**. The keys this type adds, validated by [../assets/schemas/close-out.schema.json](../assets/schemas/close-out.schema.json):

| Key | Required | Value |
| --- | --- | --- |
| `plan` | yes | The concept ID of the plan, `plan` |
| `package` | no | The concept ID of the work package file, such as `work/fare-capping-k3m7`; leave the key out when no package file exists |
| `result` | yes | `met`, `partly-met` or `not-met` |
| `criteria` | yes | One entry for each success criterion of the plan: `criterion`, `status` (`met`, `gated` or `open`) and `evidence` |
| `sources` | no | One entry for each cited file at a commit, each with an `id` |

`generated.by` is `validating-work-package/<version>`. The skill wrote the file, so it carries no `verified` entry and no `human:` actor. A `criterion` or an `evidence` value often holds a colon followed by a space, so quote it as the frontmatter rules of **recording-handoff** require.

A work package file is optional. A plan whose work items have no `parent` has no package: its package is the set of its work items that are not `deprecated`, and the close-out leaves out `package`. When a package file exists, its package is that file and the work items whose `parent` is that file, and `package` names it; a `package` that names a missing file is a validation error.

## Body

In this order: `# Independence`, `# Status` (met, gated and open as three tables), `# Integration`, `# Blockers and gates`, `# Decision register` (only when there is an entry), `# Open points and residue`, `# Follow-up seeds`, `# Not checked`. Cite evidence with footnotes keyed to the `sources` ids, as research files do.

The `criterion` text in the frontmatter is the plan's own wording, so a reader can match it to `# Success criteria` of the plan. The tables in the body repeat it with the longer evidence.

## The result

Take the result from the entries, never from the work item verdicts: `met` when every entry is met, `not-met` when none is, `partly-met` otherwise. A gated entry is not met.

## Writing and closing the run

1. Report inline, bottom line first: the result, the criteria by state, and the one blocker that decides the schedule.
2. Write `close-out.md` and announce the write in one line: the path, the result and the counts by state.
3. Write one `Workflow Log Entry` from the template in **recording-handoff** with `skill: validating-work-package`, `action: Creation` when the run created the close-out (`Update` when the run replaced an earlier record) and `close-out` in `touched`. A run that stopped on a precondition and changed nothing writes no log entry.
4. Regenerate the indexes and the log with the **recording-handoff** script, validate the bundle and report any finding.
5. Do not commit or stage. Offer the next steps and stop.

## Closing a package again

The file is replaced by the next run, and git keeps the history. A plan with more than one work package has one close-out file, so say in `# Status` which package the record closes and name the others as open or unchecked. Each run starts from the evidence of that day: it does not carry a mark over from the previous record.
