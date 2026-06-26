---
description: Close the loop on a posted review — verify each finding is fixed on the branch, resolve its blocking thread, optionally reply — on whichever host the remote points to
allowed-tools:
  - Bash
  - Read
  - Grep
  - Skill
argument-hint: >-
  [branch] [--pr <id>] [--findings <file>] [--reply] [--yes] [--dry-run]
---

# /xonovex-workflow:pr-review-resolve – Close Out a Posted Review

Symmetric end of the review pipeline: `pr-review-post` opens the blocking threads, this closes them. For each posted finding it verifies the issue is genuinely fixed on the branch, then resolves only those findings' threads (optionally replying with the fixing commit), leaving still-broken findings open.

## Arguments

`/pr-review-resolve [branch] [--pr <id>] [--findings <file>] [--reply] [--yes] [--dry-run]`

- `branch` (optional): Branch the fixes landed on (defaults to current branch)
- `--pr <id>` (optional): Target PR id directly, skipping branch lookup
- `--findings <file>` (optional): Posted findings carrying `commentId`s. Omit to use the in-session findings enriched by `pr-review-post`
- `--reply` (optional): Post a short reply on each resolved thread (e.g. "Resolved in `<commit>`")
- `--yes` (optional): Skip the confirmation and resolve immediately
- `--dry-run` (optional): Print the verdicts and stop, change nothing

## Delegation

Load these skills via the `Skill` tool; they are the source of truth for the procedure, verdict craft, matching, and gotchas — do not restate them:

- your **host skill** — detect the host from the git remote and load the matching one: `github-guide` (plugin `xonovex-skill-github`) on GitHub, `gitlab-guide` (plugin `xonovex-skill-gitlab`) on GitLab, or another installed `skill-<host>` — and perform its **review-resolve** op: auth, reading threads, matching findings to threads by `commentId`, resolving, replying.
- `code-review-guide` (plugin `xonovex-skill-code-review`) — the **verify-addressed judgment**: decide whether a finding is genuinely fixed, not just moved.

Command-level glue: route `--pr` / `branch` to PR lookup; detect the host and stop, naming the host skill to install, if none is loaded; source findings from `--findings` else the in-session set carried forward by `pr-review-post`.

## Examples

```bash
# Verify fixes and preview what would resolve, change nothing
/xonovex-workflow:pr-review-resolve --dry-run

# Resolve the fixed findings' threads, replying with the fixing commit
/xonovex-workflow:pr-review-resolve --pr 42 --findings review.json --reply
```
