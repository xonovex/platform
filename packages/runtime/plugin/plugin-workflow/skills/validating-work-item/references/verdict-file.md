# The verdict file

## Contents

- Path and keys
- Body
- Writing and closing the run
- A rerun
- Who reads the verdict

## Path and keys

`verdicts/<work-id>.md`, where `<work-id>` is the work file's name without `.md`. The type is `Workflow Verdict` and the schema is `workflow/verdict/1`, validated by [../assets/schemas/verdict.schema.json](../assets/schemas/verdict.schema.json). Use [../assets/templates/verdict.md](../assets/templates/verdict.md). The common header, the timestamps and the version in `generated.by` follow the frontmatter rules of **recording-handoff**; `generated.by` is `validating-work-item/<version>`, and the file carries no `verified` entry. The keys beyond the header:

| Key | Holds |
| --- | --- |
| `work` | Concept ID of the work file, `work/<slug>-<suffix>`. |
| `worktree_hash` | The tree hash taken at the worktree root before the first check, as the command prints it. |
| `source` | `independent` when a context that did not build the work ran the checks, else `self-check`. |
| `results` | A required list with one `{ criterion, result, evidence }` for each criterion of the work file. |

`criterion` is an id such as `c1` and must match a `criteria[].id` of the work file; the schema cannot check this, so the run does. `result` is `pass`, `fail` or `blocked`. `evidence` is one sentence that names the command or the anchor and what it showed, quoted as the frontmatter rules of **recording-handoff** require.

## Body

In this order:

1. `# Result`: one line with the counts by result, the first 12 characters of the hash, the source, then the diff stat the hash stands for.
2. `# Criteria`: one `##` section for each criterion, named by its id and text, with `Command`, `Observed` and `Limits`.
3. `# Could not check`: what was not verified and why, or the sentence "Nothing was left unchecked."

## Writing and closing the run

1. Report the results inline, outcome first: every criterion with its result in one line, and the one failure or block that matters most.
2. Create `verdicts/` when it is missing, then write or replace `verdicts/<work-id>.md` and announce it in one line: the path, the counts and the short hash. In a non-interactive run nobody reads a reply before the file exists, so the final reply carries the results after the file is written and states in the same reply that the file was written.
3. Write one `Workflow Log Entry` from the template in **recording-handoff** with `skill: validating-work-item`, `action: Creation` when the run created the verdict (`Update` when the verdict already existed) and `verdicts/<work-id>` in `touched`. A run that changed nothing, such as one that stopped at a missing precondition, writes no log entry. The run applies nothing outward, so the entry has no `authorization` or `account`.
4. Regenerate the indexes, `verdicts/index.md` included, and the log with the **recording-handoff** script, validate the bundle, and compare the verdict's `criterion` values with the work file's ids. Report any finding.
5. Do not commit or stage anything, in the plan directory or the worktree. Offer the next step and stop.

## A rerun

The next run of this skill for the same work item replaces the file, because git keeps its history. Rewrite the whole file from the new run: a new hash, new results, new evidence. Update `generated.at` and preserve unknown frontmatter keys. Do not keep results of the earlier run, since they describe other content.

## Who reads the verdict

- **preparing-delivery** marks criteria from it only when `worktree_hash` equals the worktree's hash at delivery. Otherwise the description shows the criteria unmarked. A matching verdict with `source: self-check` marks them with the label `self-check`, never as validated.
- **implementing-work-item** treats a work item as done for its dependents when its verdict has `pass` for every criterion and `source: independent`.
- **validating-work-package** reads the verdicts of the package's items.
