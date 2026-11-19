# Close-out record template

Copy the block to `plans/<change>/close-out.md`, replace every angle-bracket placeholder and delete any comment lines. The next run of the skill replaces the file; git keeps the history.

```markdown
---
type: Workflow Close-out Record
schema: workflow/close-out/1
title: <Close-out of the package as a short noun phrase>
description: <One sentence stating the result and what stands between the plan and done>
generated: {by: validating-work-package/<version>, at: <timestamp>}
plan: plan
package: work/<slug>-<suffix>
result: partly-met
criteria:
  - criterion: "<Success criterion of the plan, in its own words>"
    status: met
    evidence: "<The command or check, its output and the revision checked, re-derived in this run>"
  - criterion: "<Success criterion of the plan>"
    status: gated
    evidence: "<The gate and who passes it>"
  - criterion: "<Success criterion of the plan>"
    status: open
    evidence: "<The owner and the work that is missing>"
sources:
  - id: <source-id>
    resource: <repository-name>@<commit>:<path>
    title: <What the file is>
---

# Independence

<One sentence: the assessment ran in a fresh read-only context that implemented none of the work, or it ran in the context that did, and the record is not an independent check.>

# Status

<Two or three sentences, bottom line first: the result and what decides the schedule.>

## Met and evidenced

| Criterion | Evidence |
| --- | --- |
| <criterion> | <what was run or read in this run, the output, the revision>[^<source-id>] |

## Gated

| Criterion | Gate | Who passes it |
| --- | --- | --- |
| <criterion> | <gate, or a delivery prepared and not applied> | <owner> |

## Open

| Criterion | Owner | Missing work |
| --- | --- | --- |
| <criterion> | <owner> | <work> |

# Integration

<What was checked across the work items and what each check showed, or the sentence "The package has one work item, so there is no integration to check.">

# Blockers and gates

<N things stop the plan from being done. Only M are code.>

## Blocker 1: <name>

- Kind: <human gate, design gate or residue>
- Owner: <who>
- Gates: <what waits on it>
- Detail: <For a human gate: the exact steps, who performs each, and which step fails silently if skipped. For a design gate: the sharpest assessment and any workable design.>

## No longer a blocker

- <Closed risk and the evidence that closed it.>

<!-- Include "Decision register" when a decision is open, paused, partly settled or was made and never written. -->

# Decision register

## <Question>?

- Why it is open: <what makes it undecidable mechanically>
- Options: <option, what it costs, what it buys>
- Recommendation: <one option and the reason>
- Owner: <who>
- Waits on it: <work or seeds>
- Call: (unrecorded)

# Open points and residue

- <Item, ending in an owner, a register entry or an explicit acceptance.>

# Follow-up seeds

<One seed per coherent chunk of remaining work, from the follow-up-seed template. State the sequencing table first when there is more than one.>

# Not checked

- <What was not verified in this run, and why.>

[^<source-id>]: <Source title>, lines <from> to <to>
```

## Fill in

- `generated.by`: `validating-work-package/<version>`, and `generated.at`: the time the file is written, both by the frontmatter rules of **recording-handoff**. The skill wrote the file, so leave out `verified`.
- Quoting: a value with a colon followed by a space, a quote or a leading special character is written in single quotes (double an apostrophe inside); the criterion and evidence placeholders above are quoted for that reason (the block shows the quotes the formatter writes, and either form parses).
- `plan`: the concept ID of the plan, `plan`. `package`: the concept ID of the work package file. The key is optional: delete the `package` line when no package file exists, which is the case when the plan's work items have no `parent`; the package is then the plan's work items that are not `deprecated`.
- `result`: `met` only when every entry of `criteria` is `met`; `not-met` when none is; otherwise `partly-met`.
- `criteria`: one entry for every success criterion of the plan, in the plan's order. `status` is `met`, `gated` or `open`, and `evidence` says why in one or two sentences.
- `sources`: one entry for each cited file at the commit it was read at. Delete the key and the footnote lines when nothing is cited. For a file read in an uncommitted worktree the commit part is `tree-<hash>`, the worktree tree hash, for example `project@tree-42253ca2427219e9fd1f8d08ca50cb0e95d11791:src/greeting.mjs`. A source cited whole, not by lines, needs no line range in its footnote.
- `# Independence`: always include it.
- `# Integration`, `# Blockers and gates`, `# Open points and residue`: always include them. Write "None." under a heading with nothing to report.
- A delivery that is prepared and not applied: name it as the gate of the criteria it supports, with the user as the owner who commits and publishes it through **using-git** and **using-github** or **using-gitlab**, and list its directory under `# Open points and residue`.
- `# Decision register`: include only when there is an entry.
- `# Follow-up seeds`: include when any criterion is `gated` or `open` or an open point needs work. Otherwise write "None, the plan is met."
- `# Not checked`: always include it. Write "Nothing was left unchecked." when that is true.
