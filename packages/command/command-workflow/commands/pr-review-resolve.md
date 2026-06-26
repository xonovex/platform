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

Closes the loop after the author fixes a review. For each finding posted by `/xonovex-workflow:pr-review-post`, it verifies the issue is actually addressed on the current branch, and if so resolves the finding's blocking thread (optionally replying with the fixing commit). Still-broken findings are left open. It is the symmetric end of the pipeline: `post` opens the threads, `resolve` closes them.

It is host-agnostic by **detecting the host from the remote and driving it through that host's native CLI / API**:

- **the code review skill** (always) — judge whether a finding is genuinely addressed, not just moved.
- **the host's CLI / API for the detected host** — reading PR comments, resolving a thread, and replying on a thread (see the Host Mapping below). GitHub via `gh` is the realized host today.

## Goal

- Confirm each posted finding is fixed on the branch before touching its thread
- Resolve only the blocking threads whose findings are genuinely addressed
- Leave still-present findings open, and report fixed vs still-open
- Never blanket-resolve — verification gates every state change

## Arguments

`/pr-review-resolve [branch] [--pr <id>] [--findings <file>] [--reply] [--yes] [--dry-run]`

- `branch` (optional): Branch the fixes landed on (defaults to current branch)
- `--pr <id>` (optional): Target PR id directly, skipping branch lookup
- `--findings <file>` (optional): Posted findings carrying `commentId`s. Omit to use the in-session findings enriched by `pr-review-post`
- `--reply` (optional): Post a short reply on each resolved thread (e.g. "Resolved in `<commit>`")
- `--yes` (optional): Skip the confirmation and resolve immediately
- `--dry-run` (optional): Print the verdicts and stop, change nothing

## Host Mapping

| Neutral concept   | GitHub (`gh`, realized)                  | Other hosts                   |
| ----------------- | ---------------------------------------- | ----------------------------- |
| blocking thread   | review thread from a `--request-changes` | host's blocking-task / thread |
| resolve a thread  | `resolveReviewThread` (GraphQL)          | host's resolve mechanism      |
| reply on a thread | comment with `in_reply_to` id            | comment appended to thread    |

GitHub via `gh` is the realized host today; another host needs its own CLI / API conventions wired in. Derive the host from `git remote get-url origin`, never hardcode one.

## Core Workflow

1. **Detect host + load skills**: read the git remote, load that host's CLI / API plus the code review skill. If the host has no supported resolve path, stop and say so. Resolve the PR from `--pr` or `branch`, verify auth.
2. **Gather posted findings**: take the in-session findings (with `commentId`s) or read `--findings`. If neither has ids, read the PR's open blocking threads via the host's CLI and match them to findings by anchor.
3. **Re-check each finding**: against the current branch, decide `fixed` or `still-open`. Re-read the code at the anchor and judge whether the issue is genuinely addressed, do not assume a changed line means it is fixed.
4. **Preview**: print a table of finding -> verdict (fixed / still-open) and which threads will be resolved. Stop if `--dry-run`. Require confirmation unless `--yes`.
5. **Resolve fixed threads**: for each `fixed` finding, resolve its blocking thread via the host's CLI. Leave `still-open` ones untouched.
6. **Reply (if `--reply`)**: via the host's CLI, post a short reply on each resolved thread naming the fixing commit.
7. **Verify and report**: report resolved-this-run, still-open, and the remaining open blocking-thread total.

## Implementation Details

- **The host's CLI owns the REST calls** — reading threads, the resolve mechanism, threaded replies. This command orchestrates only.
- **Match findings to threads** by `commentId` first. Only fall back to `path` + body similarity when an id is missing — line numbers shift after fixes, so do not match on `line`.
- **Verdict needs evidence**: tie each `fixed` call to what changed at the anchor, surface it in the preview so the user can sanity-check before resolving.
- **Build JSON with a serializer** (`python3` + `json`).

## Error Handling

- Host has no supported resolve path → stop, name the host, point to wiring its CLI / API
- No `commentId`s anywhere and anchors do not match open threads → ask for `--findings`, or resolve manually in the UI
- Thread already resolved → skip and note it
- Finding still present → leave the thread open, list it as still-open (a normal outcome, not an error)
- Resolve lacks write permission → the token needs the host's PR-write scope

## Gotchas

- Resolving threads is normally the PR author's action — this command suits a reviewer verification re-pass or a self-review flow. Make the verdicts visible so the author can object.
- Verify, do not trust. A finding whose line merely moved is not fixed, judge the code, not the diff.
- This changes real PR state — keep the verify-and-confirm default, only bypass with `--yes`.
- A resolved thread only unblocks the merge where the host enforces it, do not imply the merge is now unblocked.

## Examples

```bash
# Verify fixes and preview what would resolve, change nothing
/xonovex-workflow:pr-review-resolve --dry-run

# Resolve the fixed findings' threads on the current branch's PR
/xonovex-workflow:pr-review-resolve

# Resolve and reply with the fixing commit on each thread
/xonovex-workflow:pr-review-resolve --pr 42 --findings review.json --reply

# Re-review then close the loop (cross-session)
/xonovex-workflow:pr-review-analyze feat/x --since review.json --out review.json
/xonovex-workflow:pr-review-resolve feat/x --findings review.json
```
