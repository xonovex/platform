# Log entry template

Copy the block to `plans/<change>/log/<YYYY-MM-DD>-<skill>-<suffix>.md`, replace every angle-bracket placeholder and delete any comment lines. Every skill writes one for each run that changes the bundle.

```markdown
---
type: Workflow Log Entry
schema: workflow/log-entry/1
title: <Short label for the run>
description: <One sentence stating what the run did>
generated: {by: <skill>/<version>, at: <timestamp>}
skill: <skill-name>
action: Creation
authorization: <Authorization as the skill received it>
account: <Account the effect was applied under>
touched:
  - decisions/<slug>-<suffix>
---

<One sentence stating what changed and why.>
```

## Fill in

- `action` is `Creation` when the run created a concept file other than this log entry, otherwise `Update`; `Deprecation` and `Initialization` keep their own meaning. It becomes the bold word on the generated log line.
- The first paragraph of the body is the summary: one sentence that the generator copies into `log.md`, so write it for a reader who sees only that line. Further content, such as a table, may follow after a blank line and stays out of `log.md`.
- `authorization`: include when the run applied something outward. Record the words the skill received, such as the invoker's approval, never a secret. Write the value in single quotes when it contains a colon followed by a space, a quote or a leading special character, for example `authorization: 'The invoker said: go ahead'`. The same holds for `title` and `description`.
- `account`: include when an outward effect ran under a named account.
- A log entry carries no `verified` key: the skill wrote it and nobody else confirms it.
- `touched` lists the concept files the run created, changed or deprecated, and never this log entry itself. Omit the key when the list is empty.
- Every `<timestamp>` comes from the system clock: run `date -u +%Y-%m-%dT%H:%M:%SZ` and use its output; never invent, round or copy a time from another file. `<version>` in `generated.by` is the `version` field of the `plugin.json` of the plugin that ships the skill.
- The suffix in the file name is four random characters from `a-z` and `2-7` only: `0`, `1`, `8` and `9` are never used. The same holds for run ids.
- A run that changed nothing writes no log entry.
