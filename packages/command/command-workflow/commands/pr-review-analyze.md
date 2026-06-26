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

Reviews the changes on a branch and produces structured findings — a top-level summary plus line-anchored findings, each with a Conventional Comments label and a blocking flag. The findings are the contract consumed by `/xonovex-workflow:pr-review-refine` and `/xonovex-workflow:pr-review-post`; this stage analyzes only and posts nothing.

Findings are an in-session artifact by default — when these commands run together the next stage reads them straight from context. Pass `--out` to persist a `findings.json` only to cross sessions, inspect, or hand-edit.

## Arguments

`/pr-review-analyze [branch] [--base <ref>] [--out <file>] [--effort <low|medium|high>] [--diff <file>] [--since <file>]`

- `branch` (optional): Branch to review (defaults to current branch)
- `--base <ref>` (optional): Base ref to diff against (defaults to `main`)
- `--out <file>` (optional): Persist the findings to a JSON file (`-` for stdout). Omit it when chaining in one session — the findings stay in context for the next stage
- `--effort <level>` (optional): Review depth. `low` / `medium` favour fewer high-confidence findings, `high` widens coverage and may include uncertain ones (use `question`)
- `--diff <file>` (optional): Review a pre-saved unified diff instead of computing one
- `--since <file>` (optional): A prior findings file (or in-session prior findings) to compare against. Tags each fresh finding `recurring` or `new`, and lists prior findings now `gone`. Compares findings only, never the PR — stays platform-independent

## Delegation

Load `code-review-guide` (plugin `xonovex-skill-code-review`) via the `Skill` tool and perform its **`review-analyze`** operation; the findings shape and new-file line anchoring come from its **`findings-schema`** operation, the label vocabulary and decorations from **`conventional-comments`**. They are the source of truth for the review craft and the JSON contract — do not restate them. Also load relevant domain skills (design-system, accessibility, language) so findings are grounded in the code, not just its diff.

Command-level orchestration only: this stage analyzes and posts nothing; `--since` runs `review-analyze`'s recurrence comparison against findings only (never the PR, keeping analyze platform-independent); findings stay in session for the next stage by default, while `--out` persists a `findings.json` (falling back to stdout if the path is not writable); an empty `base...branch` diff yields an empty findings list and says so, and an unknown base ref asks for `--base`.

## Examples

```bash
# Review the current branch against main, findings kept in session
/xonovex-workflow:pr-review-analyze

# Deeper pass against a release base
/xonovex-workflow:pr-review-analyze feat/load-indicator --base release/9 --effort high

# Persist to a file to inspect or hand-edit before posting
/xonovex-workflow:pr-review-analyze --out review.json

# Cross-session pipeline via a file
/xonovex-workflow:pr-review-analyze feat/x --out review.json
/xonovex-workflow:pr-review-post feat/x --findings review.json
```
