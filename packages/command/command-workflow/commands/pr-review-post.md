---
description: Publish a structured code review to a pull request — summary plus labelled inline comments, blocking tasks, and cross-links — on whichever host the remote points to
allowed-tools:
  - Bash
  - Read
  - Skill
argument-hint: >-
  [branch] [--pr <id>] [--findings <file>] [--no-tasks] [--yes] [--dry-run]
---

# /xonovex-workflow:pr-review-post – Publish a Structured PR Review

The publish stage of the pr-\* review pipeline: takes an already-formed review (one summary plus labelled, line-anchored findings, typically from a same-session `pr-review-analyze` / `pr-review-refine`) and posts it to the PR on whatever host the remote points to, ready for `pr-review-resolve` to close the threads.

## Arguments

`/pr-review-post [branch] [--pr <id>] [--findings <file>] [--no-tasks] [--yes] [--dry-run]`

- `branch` (optional): Source branch whose open PR to target (defaults to current branch)
- `--pr <id>` (optional): Target PR id directly, skipping branch lookup
- `--findings <file>` (optional): Read findings from a JSON file. Omit it to post the findings already in the session context (e.g. from a same-session `pr-review-analyze` / `pr-review-refine` run)
- `--no-tasks` (optional): Post blocking items as labelled comments only, do not promote them to a blocking review
- `--yes` (optional): Skip the preview confirmation and post immediately
- `--dry-run` (optional): Print the preview and stop, post nothing

## Delegation

Load these skills via the `Skill` tool and perform their operations; they are the source of truth for the procedure, format, validation, and gotchas — do not restate them:

- `code-review-guide` (plugin `xonovex-skill-code-review`) — the **review craft** (label vocabulary, blocking / non-blocking / if-minor decorations, summary-plus-inline structure, cross-linking instead of "see comment 3") and the **`findings-schema`** contract the input must satisfy.
- your **host skill** — detect the host from the git remote and load the matching one: `github-guide` (plugin `xonovex-skill-github`) on GitHub, `gitlab-guide` (plugin `xonovex-skill-gitlab`) on GitLab, or another installed `skill-<host>` — and perform its **review-post** op: auth, anchored inline comments, blocking-review promotion, comment deep-links, and id write-back.

Command-level orchestration only: route the target from `--pr` or the open PR for `branch`; detect the host and stop, naming the host skill to install, if none is loaded; source findings from `--findings` else the in-session set, and write the returned `commentId`s back to both the session and that file; `--no-tasks` suppresses blocker promotion, `--dry-run` stops after the preview, `--yes` skips the confirm.

## Examples

```bash
# Preview the review for the current branch's PR, post nothing
/xonovex-workflow:pr-review-post --dry-run

# Post the in-session findings to whatever host the remote points to
/xonovex-workflow:pr-review-post feat/x

# Post findings from a file, target a PR id, skip the prompt
/xonovex-workflow:pr-review-post --pr 42 --findings review.json --yes
```
