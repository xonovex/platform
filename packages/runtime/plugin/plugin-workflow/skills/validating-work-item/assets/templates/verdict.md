# Verdict template

Copy the block to `plans/<change>/verdicts/<work-id>.md`, where `<work-id>` is the work file's name without `.md`, replace every angle-bracket placeholder and delete any comment lines. The next validation of the same work item replaces the file.

````markdown
---
type: Workflow Verdict
schema: workflow/verdict/1
title: <Verdict for the work item, as a short noun phrase>
description: <One sentence stating how many criteria passed, failed or were blocked>
generated: {by: validating-work-item/<version>, at: <timestamp>}
work: work/<slug>-<suffix>
worktree_hash: <tree hash printed by the command in the tree-hash reference>
source: independent
results:
  - criterion: c1
    result: pass
    evidence: "<Command or anchor, and what it showed>"
  - criterion: c2
    result: fail
    evidence: "<Command or anchor, and the measured value>"
  - criterion: c3
    result: blocked
    evidence: "<What could not run, and what would unblock it>"
---

# Result

<Counts by result>, tree <first 12 characters of the hash>, <independent or self-check>.

Diff checked:

```text
<output of git diff --stat HEAD <hash>>
```

# Criteria

## c1 <criterion text>

- Command: <the command as run, or the anchor>
- Observed: <short excerpt, measured number or file and line>
- Limits: <what this check does not cover>

## c2 <criterion text>

- Command: <the command as run>
- Observed: <the measured value against the expected one>
- Limits: <baseline unknown, platform not tried, and similar>

## c3 <criterion text>

- Command: <the command that could not run>
- Observed: <the error or the missing input>
- Limits: <what would unblock it and who can provide it>

# Could not check

- <What was not verified, and why.>
````

## Fill in

- `generated.by`: `validating-work-item/<version>`, and `generated.at`: the time the file is written, both by the frontmatter rules of **recording-handoff**. The skill wrote the file, so leave out `verified`.
- `work`: the concept ID of the work file.
- `worktree_hash`: the hash taken at the worktree root before the first check, exactly as the command prints it.
- `source`: `independent` when a context that did not build the work ran the checks, else `self-check`; head the `# Result` line with the same label.
- `results`: one entry for each `criteria[].id` of the work file, in its order, and no other id. Write `evidence` in single quotes (double an apostrophe inside it; the block above shows the quotes the formatter writes, and either form parses), because a colon followed by a space in an unquoted value breaks the frontmatter; the same holds for a `title` or `description` that holds one. Keep it to one sentence; the detail goes in the body.
- `# Result`: the counts, the short hash and the diff stat. When the invoker confirmed a run on an empty diff, say so instead of the stat.
- `# Criteria`: one section for each criterion, in the same order, with the three fields.
- `# Could not check`: always include it. Write "Nothing was left unchecked." when that is true.
