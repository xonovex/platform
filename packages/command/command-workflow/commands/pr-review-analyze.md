---
description: Review a branch diff into structured findings (Conventional Comments labels, blocking flags) — kept in session, or written to a file
allowed-tools:
  - Bash
  - Read
  - Grep
  - Skill
argument-hint: >-
  [branch] [--base <ref>] [--out <file>] [--effort <low|medium|high>] [--diff <file>]
---

# /xonovex-workflow:pr-review-analyze – Review a Diff into a Findings File

Reviews the changes on a branch and produces structured findings — a top-level summary plus line-anchored findings, each with a Conventional Comments label and a blocking flag. The findings are the contract consumed by `/xonovex-workflow:pr-review-refine` and `/xonovex-workflow:pr-review-post`. This command does the analysis only, it posts nothing.

The findings are an in-session artifact by default: when these commands run together in one session, the next stage reads them straight from context, no file needed. Pass `--out` to persist a `findings.json` only when you need to cross sessions, inspect, or hand-edit it.

It composes with the **code review skill** for the review craft: the label vocabulary, blocking / non-blocking / if-minor decorations, summary-plus-inline structure, self-contained comments, and the verify-before-asserting discipline. Load any relevant domain skills too (design-system, accessibility, language) — finding quality depends on understanding the code, not just its diff.

## Goal

- Turn a branch diff into structured, labelled, line-anchored findings
- Separate correctness issues from quality suggestions, and blocking from non-blocking
- Anchor every finding to a real new-file line so it cannot orphan when posted
- Hand the findings to the next stage in context, or persist them with `--out`

## Arguments

`/pr-review-analyze [branch] [--base <ref>] [--out <file>] [--effort <low|medium|high>] [--diff <file>] [--since <file>]`

- `branch` (optional): Branch to review (defaults to current branch)
- `--base <ref>` (optional): Base ref to diff against (defaults to `main`)
- `--out <file>` (optional): Persist the findings to a JSON file (`-` for stdout). Omit it when chaining in one session — the findings stay in context for the next stage
- `--effort <level>` (optional): Review depth. `low` / `medium` favour fewer high-confidence findings, `high` widens coverage and may include uncertain ones (use `question`)
- `--diff <file>` (optional): Review a pre-saved unified diff instead of computing one
- `--since <file>` (optional): A prior findings file (or in-session prior findings) to compare against. Tags each fresh finding `recurring` or `new`, and lists prior findings now `gone`. Compares findings only, never the PR — stays platform-independent

## Core Workflow

1. **Load craft**: invoke the code review skill for labels, decorations, and structure. Load relevant domain skills so claims are grounded.
2. **Compute the diff**: `git diff <base>...<branch>` (or read `--diff`). Track new-file line numbers from each hunk header so findings can anchor to `ADDED` / `CONTEXT` lines.
3. **Read for correctness first**: bugs, regressions, broken edge cases, accessibility and contract violations. Re-read surrounding code to confirm each claim, do not assert from the diff alone.
4. **Then quality**: reuse, simplification, efficiency, naming, consistency. Right-size to the change — a small diff does not need a long list.
5. **Label each finding**: pick a Conventional Comments `label` and a `decoration`. Set `blocking: true` only for must-fix-before-merge items. If unsure a thing is real, use `question`, not `issue`.
6. **Anchor each finding**: a specific new-file `path` + `line` + `lineType` that exists in the diff.
7. **Draft the summary**: lead with what works, then number the priority (blocking) points so cross-links can attach later.
8. **Diff against prior (if `--since`)**: match fresh findings to the prior set by `path` plus body similarity (not line number, which shifts), tag each `new` or `recurring`, and report prior findings with no match as `gone`. Carry the prior `commentId` onto a recurring finding so it stays linked.
9. **Hand off and validate**: keep the findings in context for the next stage, writing a `findings.json` only if `--out` is given. Confirm every `path`/`line` is a real diff line, then print a short table for review.

## Findings Schema

The same shape whether held in context or written to a file:

```jsonc
{
  "summary": "Markdown body. Lead with positives. Number the priority points.",
  "findings": [
    {
      "path": "packages/.../module.ts", // repo-relative path in the new file version
      "line": 420, // new-file line number present in the diff
      "lineType": "ADDED", // ADDED | CONTEXT
      "label": "issue", // praise | nitpick | suggestion | issue | question | thought | chore
      "decoration": "blocking", // blocking | non-blocking | if-minor
      "blocking": true, // must-fix-before-merge
      "body": "Self-contained markdown: the problem, why, and a suggested fix.",
      "status": "new", // new | recurring — only set when --since is used
      "commentId": 101, // carried from the prior findings when recurring
    },
  ],
}
```

## Implementation Details

- **New-file line numbers**: parse hunk headers `@@ -a,b +c,d @@` and walk the `+`/context lines so each anchor uses the post-change line (matches the poster's new-file side). Added lines are `ADDED`, unchanged in-hunk lines are `CONTEXT`.
- **Build JSON with a serializer** (`python3` + `json`) so bodies with backticks, quotes, and newlines stay valid.
- **Self-contained bodies**: no "see the other comment" — each finding stands alone, since they may be posted, edited, or resolved independently.
- **Effort dial**: lower effort returns only what you are confident about, higher effort trades precision for recall and leans on `question` for the uncertain ones.

## Error Handling

- `base...branch` empty → nothing changed to review, write an empty findings list and say so
- Base ref not found → list candidate refs, or ask for `--base`
- A finding line is not in the diff → it would orphan when posted, re-anchor to an `ADDED` / `CONTEXT` line in the changed hunk
- `--out` path not writable → fall back to stdout and keep the findings in context

## Gotchas

- Anchors must be **new-file** line numbers, not old-file or absolute — otherwise the poster orphans them.
- Do not invent findings to fill space. Zero issues on a clean diff is a valid result, right-size the review.
- Lead the summary with what works, a review that is all negatives hides the signal.
- Match the repo's prose rules (commit / PR conventions often apply to review comments too) — check the project instructions.
- This emits findings only. Publishing, labels-as-tasks, and cross-links are `/xonovex-workflow:pr-review-post`'s job.

## Examples

```bash
# Review the current branch against main, findings kept in session
/xonovex-workflow:pr-review-analyze

# Review a branch against a release base, deeper pass
/xonovex-workflow:pr-review-analyze feat/load-indicator --base release/9 --effort high

# Persist to a file to inspect or hand-edit before posting
/xonovex-workflow:pr-review-analyze --out review.json

# Same-session pipeline, no file needed
/xonovex-workflow:pr-review-analyze feat/x
/xonovex-workflow:pr-review-post feat/x

# Cross-session pipeline via a file
/xonovex-workflow:pr-review-analyze feat/x --out review.json
/xonovex-workflow:pr-review-post feat/x --findings review.json
```
