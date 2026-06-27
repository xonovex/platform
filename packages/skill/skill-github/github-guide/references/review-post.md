# review-post: publish a structured review on GitHub

## Guideline

Batch the summary body, every line-anchored inline comment, and the verdict into ONE `POST .../pulls/{n}/reviews` object — anchored to the PR HEAD sha. This realizes the findings authored per **`code-review-guide`** on GitHub; that skill owns the labels, severity, and blocking decoration — this file only anchors and submits them.

## Rationale

A GitHub review is a first-class atomic object: one call carries summary + inline comments + the APPROVE / REQUEST_CHANGES / COMMENT verdict. `gh pr review` posts only the review-level body (no inline support, cli/cli#12396), and the standalone `.../pulls/{n}/comments` endpoint frequently 422s on `line`/`side` (cli/cli#13358) — so build the inline comments inside the review object.

## The one review call

```bash
HEAD=$(gh pr view 123 --json headRefOid -q .headRefOid)
gh api --method POST repos/{owner}/{repo}/pulls/123/reviews \
  -f commit_id="$HEAD" \
  -f event=REQUEST_CHANGES \
  -f body=$'## Summary\nOne blocking issue inline.' \
  -f 'comments[][path]=src/app.ts' \
  -F 'comments[][line]=42' \
  -f 'comments[][side]=RIGHT' \
  -f 'comments[][body]=**issue (blocking):** guard against null here.'
```

For many comments, build the JSON and pipe it instead of repeating `comments[]`:

```bash
jq -n --arg c "$HEAD" '{commit_id:$c, event:"REQUEST_CHANGES", body:"## Summary\n…",
  comments:[{path:"src/a.ts",line:42,side:"RIGHT",body:"…"},
            {path:"src/b.ts",start_line:10,start_side:"RIGHT",line:14,side:"RIGHT",body:"…"}]}' \
  | gh api --method POST repos/{owner}/{repo}/pulls/123/reviews --input -
```

- `event` = `APPROVE` | `REQUEST_CHANGES` | `COMMENT`. `body` is **required** for REQUEST_CHANGES and COMMENT.
- Omit `event` entirely → the review stays **PENDING** (draft) until submitted.

## Inline anchor model (exact)

Each `comments[]` item needs `path` + `body`, plus the anchor:

- **`line`** = the line number **in the file**, NOT a diff position. The legacy `position` field is deprecated — never compute it.
- **`side`** = `RIGHT` (head / added / unchanged-context, the default) or `LEFT` (base / deleted).
- **Multi-line range:** add `start_line` + `start_side` describing the FIRST line; they must precede `line`/`side` and sit in the same diff hunk. They are required (not optional) when a comment spans multiple lines.
- **`commit_id`** = the PR HEAD sha. Omitting it defaults to the latest commit, but passing the HEAD sha explicitly is the documented correct usage; a stale sha renders the comment "outdated" once new commits land.
- Use `-F` for integer `line`/`start_line`, `-f` for string fields.

## What you cannot do this way

- `gh pr review --approve/--request-changes/--comment --body/--body-file` posts ONLY the review-level body — no inline. Reach for `gh api` whenever inline comments are needed.
- You **cannot** APPROVE or REQUEST_CHANGES your **own** PR (HTTP 422) — self-reviews are COMMENT-only.

## Blocking mechanism

REQUEST_CHANGES hard-blocks merge **only** when branch protection / a ruleset has "Require a pull request before merging" with required approvals enabled — otherwise it is advisory. It clears in exactly two ways:

1. the **same** reviewer switches to APPROVE, or
2. someone with write access **dismisses** it: `PUT .../pulls/{n}/reviews/{id}/dismissals`.

A second person's approval does NOT override an outstanding REQUEST_CHANGES. Check gating with `gh pr view N --json reviewDecision,mergeStateStatus,mergeable` (`mergeStateStatus=BLOCKED`).

## Deep links

Take them from the response `html_url`; never reconstruct anchors by hand:

- inline comment → `#discussion_r<id>`
- review → `#pullrequestreview-<id>`
- top-level PR comment (`gh pr comment`, an issue comment) → `#issuecomment-<id>` — and it never carries a verdict.

### Related

[review-resolve.md](./review-resolve.md), [auth.md](./auth.md)
