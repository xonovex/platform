---
description: Open a pull request on whichever host the remote points to — push the branch, draft a what/why/how description from the diff, then create it with reviewers, work-item and sibling-PR links
allowed-tools:
  - Bash
  - Read
  - Grep
  - Skill
argument-hint: >-
  [branch] [--base <ref>] [--title <text>] [--description <file>] [--draft]
  [--reviewers <a,b>] [--work-item <id>] [--related <id,...>] [--yes] [--dry-run]
---

# /xonovex-workflow:pr-create – Open a Pull Request with a Drafted Description

Opens a pull request for a branch on whatever host the git remote points to: pushes the branch, drafts a what / why / how description from the real diff and commits, and creates the PR with optional reviewers, a linked work item, and cross-links to sibling PRs. It is the author-side counterpart to the `pr-review-*` pipeline — this opens the PR, those review it.

## Arguments

`/pr-create [branch] [--base <ref>] [--title <text>] [--description <file>] [--draft] [--reviewers <a,b>] [--work-item <id>] [--related <id,...>] [--yes] [--dry-run]`

- `branch` (optional): Source branch to open the PR from (defaults to current branch)
- `--base <ref>` (optional): Target branch the PR merges into (defaults to `main`)
- `--title <text>` (optional): PR title (if omitted, generates a conventional-commit title from the branch's commits)
- `--description <file>` (optional): Read the PR body from a file (`-` for stdin) instead of generating it. The generated body still follows the pull request skill's template
- `--draft` (optional): Open as a draft / work-in-progress where the host supports it
- `--reviewers <a,b>` (optional): Comma-separated reviewers to add after creating
- `--work-item <id>` (optional): Link a work item / issue to the PR
- `--related <id,...>` (optional): Cross-link sibling PRs of a coordinated change set
- `--yes` (optional): Skip the preview confirmation and create immediately
- `--dry-run` (optional): Print the title, target, and rendered description, then stop — create nothing

## Delegation

Load and follow these skills via the `Skill` tool; they are the source of truth — do not restate them:

- `pull-request-guide` (plugin `xonovex-skill-pull-request`) — the **description craft**: what / why / how grounded in the diff, sizing and splitting an oversized diff, testing evidence, tradeoffs, the lean template, and the self-review + CI-green gate.
- `git-guide` (plugin `xonovex-skill-git`) — branch, commit, the conventional-commit title, and the **push + rebase-onto-base** that readies the branch.
- your **host skill** — detect the host from the git remote and load the matching one: `github-guide` (plugin `xonovex-skill-github`) on GitHub, `gitlab-guide` (plugin `xonovex-skill-gitlab`) on GitLab, or another installed `skill-<host>` — and perform its **create** op: auth, PR / MR create, draft, reviewers, work-item and sibling links, additive body edits, and the existing-PR / no-create-path / auth handling.

Command-level glue (not skill craft): route `[branch]` / `--base` to the create op's source / target (defaults: current branch, `main`); `--title` / `--description <file>` override the generated title / body; detect the host before anything and stop, naming the host skill to install, if none is loaded for it; preview the title + body and require confirmation unless `--yes`, stopping after the preview on `--dry-run`.

## Examples

```bash
/xonovex-workflow:pr-create
/xonovex-workflow:pr-create feat/load-indicator --base release/9 --draft
/xonovex-workflow:pr-create --description pr-body.md --reviewers alice,bob --work-item 1234 --yes
```
