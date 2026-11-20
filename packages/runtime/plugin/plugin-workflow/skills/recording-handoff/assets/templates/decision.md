# Decision template

Copy the block to `plans/<change>/decisions/<slug>-<suffix>.md`, replace every angle-bracket placeholder and delete any comment lines. A decision file is never edited after it is written.

```markdown
---
type: Workflow Decision
schema: workflow/decision/1
title: <Decision title as a short noun phrase>
description: <One sentence stating the decision or the question to settle>
generated: {by: <skill>/<version>, at: <timestamp>}
verified: {by: human:<id>, at: <timestamp>}
state: settled
owner: <person or team who settles it>
depends_on:
  - decisions/<slug>-<suffix>
blocks:
  - decisions/<slug>-<suffix>
supersedes:
  - decisions/<slug>-<suffix>
sources:
  - id: <source-id>
    resource: <https://example.org/page>
    title: <Source title>
---

# Decision

<One testable sentence.>

# Reason

<Why this option wins, citing the evidence.>[^<source-id>]

# Options considered

| Option            | Gains           | Costs           |
| ----------------- | --------------- | --------------- |
| <Chosen option>   | <What it gives> | <What it costs> |
| <Rejected option> | <What it gives> | <What it costs> |

<!-- Include "Holds either way" when some outcomes are the same whichever option is chosen. -->

# Holds either way

- <Outcome that does not depend on the choice>

[^<source-id>]: <Source title>
```

## Fill in

- `state` is `settled`, `open`, `paused` or `partly-settled`. For a state other than `settled`, `# Decision` states the question to settle in one sentence and `# Reason` says why it matters.
- `owner`: include when one person or team settles the question.
- `depends_on` and `blocks`: include when another decision must be settled first, or waits on this one. Values are concept IDs without `.md`.
- `supersedes`: include only when this file replaces earlier decisions. Write the new file first, then set `status: deprecated` and `superseded_by` on each old file and change nothing else in it.
- `sources`: include when a claim in the body rests on evidence. Each `id` is a kebab-case key that a footnote label repeats. Delete the key and the footnote lines when nothing is cited.
- `verified`: include when a person settled or confirmed the decision in this session. The skill stays `generated.by`, and the person is `human:<id>`, taken from the invoker; ask for it when the invocation does not give it and never invent one. Delete the key for a decision nobody confirmed, such as an open question the skill raised.
- A decision file needs no `status`, because it is written only after the invoker confirms it.
- Every `<timestamp>` comes from the system clock: run `date -u +%Y-%m-%dT%H:%M:%SZ` and use its output; never invent, round or copy a time from another file. `<version>` in `generated.by` is the `version` field of the `plugin.json` of the plugin that ships the skill.
- Write a `title`, `description` or other string value in single quotes when it contains a colon followed by a space, a quote or a leading special character.
- The suffix of the file name is four random characters from `a-z` and `2-7` only (never `0`, `1`, `8` or `9`).
