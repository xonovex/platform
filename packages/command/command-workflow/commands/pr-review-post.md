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

Publishes an already-formed code review to the pull request on whatever host the git remote points to: one top-level summary plus line-anchored inline comments, each carrying a Conventional Comments label, with blocking items promoted to the host's blocking mechanism and the summary cross-linked to its inline details.

It is host-agnostic by **detecting the host from the remote and driving it through that host's native CLI / API**, so the workflow stays the same while the REST specifics live in one place per host:

- **the code review skill** (always) — label vocabulary (praise / nitpick / suggestion / issue / question / thought / chore), the blocking / non-blocking / if-minor decorations, summary-plus-inline structure, self-contained comments, cross-linking instead of "see comment 3".
- **the host's CLI / API for the detected host** — auth, find-PR, inline anchors, blocking mechanism, and comment deep-links (see the Host Mapping below). GitHub via `gh` is the realized host today.

## Goal

- Publish a labelled, structured review to the PR on the detected host
- Promote blocking findings to the host's blocking mechanism
- Cross-link the summary to each inline detail, and record each comment id onto its finding
- Preview before posting (review is outward-facing) and verify after

## Arguments

`/pr-review-post [branch] [--pr <id>] [--findings <file>] [--no-tasks] [--yes] [--dry-run]`

- `branch` (optional): Source branch whose open PR to target (defaults to current branch)
- `--pr <id>` (optional): Target PR id directly, skipping branch lookup
- `--findings <file>` (optional): Read findings from a JSON file. Omit it to post the findings already in the session context (e.g. from a same-session `pr-review-analyze` / `pr-review-refine` run)
- `--no-tasks` (optional): Post blocking items as labelled comments only, do not promote them to a blocking review
- `--yes` (optional): Skip the preview confirmation and post immediately
- `--dry-run` (optional): Print the preview and stop, post nothing

## Host Mapping

The workflow speaks in neutral concepts, the host's CLI / API realizes each:

| Neutral concept               | GitHub (`gh`, realized)                   | Other hosts                    |
| ----------------------------- | ----------------------------------------- | ------------------------------ |
| inline comment at `file:line` | review comment (`path` + `line` + `side`) | host's anchored-comment call   |
| blocking task                 | `gh pr review --request-changes`          | host's blocking-task mechanism |
| comment deep-link             | comment `html_url`                        | host's discussion-thread link  |

GitHub via `gh` is the realized host today; another host needs its own CLI / API conventions wired in before this command can target it. Derive the host from `git remote get-url origin`, never hardcode one.

## Core Workflow

1. **Detect host + load skills**: read the git remote, pick the host's CLI / API for that host, and load the code review skill. If the host has no supported post path, stop and say so.
2. **Resolve target**: derive the repo coordinates from the remote (do not hardcode), pick the PR from `--pr` or the open PR for `branch`, and verify auth.
3. **Take findings**: read `--findings` if given, else use the findings already in the session context. Each finding needs `path`, `line`, `lineType`, a Conventional Comments `label`, a `blocking` flag, and a `body`. Plus one `summary`.
4. **Validate**: every finding has a known label and an explicit blocking / non-blocking decoration, every `path`/`line` resolves to a diff line, the summary references each blocking finding.
5. **Preview**: print the summary and a table of inline comments (file:line, label, blocking). Stop if `--dry-run`. Require confirmation unless `--yes`.
6. **Post inline comments**: for each finding, prepend `**<label> (<decoration>)**` to the body and post it anchored to its line via the host's CLI. Capture each new comment id, confirm it anchored (not orphaned).
7. **Promote blockers**: unless `--no-tasks`, mark every blocking finding with the host's blocking mechanism (on GitHub, a `--request-changes` review).
8. **Post and cross-link summary**: post the top-level summary, then edit it to add a `([details](<deep-link>))` link to each blocking finding's comment id.
9. **Record ids**: write each finding's new `commentId` (and the summary's) back onto the findings — into the session context, and into the file if `--findings` was given — so `pr-review-resolve` can match findings to threads later.
10. **Verify**: report the posted comments and the open blocking total.

## Findings Schema

The same shape whether passed in context or read from `--findings`:

```jsonc
{
  "summary": "Markdown body. Number the priority points so cross-links can attach.",
  "findings": [
    {
      "path": "packages/.../module.ts", // repo-relative path in the new file version
      "line": 420, // line number in the diff
      "lineType": "ADDED", // ADDED | CONTEXT
      "label": "issue", // a Conventional Comments label
      "decoration": "blocking", // blocking | non-blocking | if-minor
      "blocking": true, // true -> promoted to a blocking review
      "body": "Self-contained markdown with the problem and a suggested fix.",
      "commentId": 101, // written back by this command after posting
    },
  ],
}
```

## Implementation Details

- **The host's CLI owns the REST calls** — anchors, blocking promotion, edit-with-version, deep-link form. This command orchestrates, it does not hardcode one host's API.
- **Build JSON with a serializer** (`python3` + `json`), never hand-escape — bodies contain backticks, quotes, and newlines.
- **Label prefix**: a bold lead-in line `**<label> (<decoration>)**` then a blank line then the body. Skip if the body already starts with the label (idempotent re-runs).
- **Cross-link last**: ids do not exist until posted, so cross-linking is always the final edit.
- **Write back ids**: enrich each finding with its `commentId` so `pr-review-resolve` can act, re-matching by `path:line` is fragile once the code shifts.

## Error Handling

- Host has no supported post path → stop, name the host, and point to wiring its CLI / API
- No PR for branch → list open PRs, or ask for `--pr`
- Auth failure → defer to the host CLI's auth guidance (token, scope — e.g. `gh auth status`)
- Post lacks write permission → the token needs the host's PR-write scope, blocking reviews especially
- Anchor did not attach (orphaned) → the line is not in the diff, pick an ADDED / CONTEXT line in the changed hunk

## Gotchas

- A blocking review only _gates_ the merge where the host enforces it (e.g. GitHub branch protection requiring an approving review). Do not imply the merge is blocked, say it depends on the host's settings.
- Resolve repo coordinates from the remote each run — hardcoding breaks reuse and cross-host portability.
- Cross-linking must come after posting, the ids do not exist until then.
- This posts to a real PR — keep the preview-and-confirm default, only bypass with `--yes`.

## Examples

```bash
# Preview the review for the current branch's PR, post nothing
/xonovex-workflow:pr-review-post --dry-run

# Post the in-session findings to whatever host the remote points to
/xonovex-workflow:pr-review-post feat/x

# Post findings from a file, target a PR id, skip the prompt
/xonovex-workflow:pr-review-post --pr 42 --findings review.json --yes

# Labelled comments only, no blocking review
/xonovex-workflow:pr-review-post --pr 42 --findings review.json --no-tasks
```
