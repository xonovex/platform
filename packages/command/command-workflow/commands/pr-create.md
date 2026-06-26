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

Opens a pull request for a branch on whatever host the git remote points to: pushes the branch, drafts a structured description from the actual diff and commits (what / why / how, testing, tradeoffs), and creates the PR — optionally with reviewers, a linked work item / issue, and cross-links to sibling PRs. It is the author-side counterpart to the `pr-review-*` pipeline: this opens the PR, those review it.

It is host-agnostic by **detecting the host from the remote and driving it through that host's native CLI / API**, so the workflow stays the same while the create / link specifics live in one place per host:

- **the pull request skill** (always) — the description craft: what / why / how, matching depth to size, one concern per PR, real testing evidence over "tested locally", surfacing tradeoffs early, the lean template, and the self-review gate before assigning.
- **the host's CLI / API for the detected host** — auth, repo coordinates, branch push, `create PR`, draft state, reviewers, and PR / work-item linking (see the Host Mapping below). GitHub via `gh` is the realized host today.
- **the git skill** — branch, commit, the conventional-commit title, and the rebase-onto-base that keeps the PR diff clean.

## Goal

- Push the branch and open a PR against the base, on the detected host
- Draft a description grounded in the real diff and commits, not invented
- Right-size the change (flag an oversized, splittable diff) and gate on a self-review before opening
- Preview before creating (a PR is outward-facing) and report the clickable URL

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

## Host Mapping

The workflow speaks in neutral concepts, the host's CLI / API realizes each:

| Neutral concept           | GitHub (`gh`, realized)               | Other hosts                       |
| ------------------------- | ------------------------------------- | --------------------------------- |
| push the branch           | `git push -u origin <branch>`         | `git push -u origin <branch>`     |
| create the PR             | `gh pr create --base --head`          | host's `create pull request` call |
| set / refresh description | `gh pr edit --body`                   | host's `update --description`     |
| open as draft             | `gh pr create --draft`                | where the host has a draft state  |
| add reviewers             | `gh pr create --reviewer` / `pr edit` | host's reviewer mechanism         |
| link a work item / issue  | `Closes #<id>` / `#<id>` in the body  | work-item ref in the description  |
| cross-link sibling PRs    | PR URL in the body                    | PR URL / ref in the description   |

GitHub via `gh` is the fully-owned host today. Another host needs its own CLI / API conventions wired in (auth, coordinates, create, link) before this command can target it — the neutral concepts above are what each one must realize. Derive the host from `git remote get-url origin`, never hardcode one.

## Core Workflow

1. **Detect host + load skills**: read the git remote, pick the host's CLI / API for that host, and load the pull request skill and the git skill. If the host has no supported create path, stop and say so.
2. **Resolve source + target**: source branch from the argument or the current branch, base from `--base` or `main`. Derive the repo coordinates from the remote (do not hardcode), and verify auth.
3. **Ready the branch**: ensure changes are committed (defer to the git skill), push the branch, and if it is behind the base, `fetch` + `rebase` onto the base so the PR diff is just this change. Confirm checks / CI are green where the host exposes them — the pull request skill's self-review gate.
4. **Right-size**: compute `git diff --stat <base>...<branch>`. If the diff is large and not legitimately big (scaffolding, generated code, dependency bumps), surface that it may want splitting; if it is legitimately big, plan to call it out in the description so the reviewer calibrates.
5. **Draft the content** with the pull request skill: a conventional-commit `--title` (or generate one from the commits), and a description with `## What / ## Why / ## Changes / ## Testing / ## Tradeoffs / ## Related` — grounded in the diff, the commit messages, and any `--work-item`. Match each section's depth to the size of the change. Read `--description` from a file instead if given.
6. **Preview + confirm**: print the title, base, and rendered description. Stop if `--dry-run`. Require confirmation unless `--yes` — creating a PR is outward-facing.
7. **Create** via the host's CLI: the branch is pushed, so create the PR against the base, capture the returned id and URL. `--draft` opens it as work-in-progress where supported.
8. **Enrich**: add `--reviewers`, link `--work-item`, and cross-link `--related` siblings, each via the host's mechanism. In a coordinated set, link the siblings to each other.
9. **Verify**: confirm the PR has no conflicts against its base (the host's merge-status check), then report the clickable URL.

## Description Template

The body the pull request skill fills, matched in depth to the change:

```markdown
## What

<One or two sentences on the net effect of the change.>

## Why

<The engineering or product goal. Link the work item.>

## Changes

- <Key change, grouped by area.>

## Testing

- <Automated tests added or updated, and what they cover.>
- <Manual verification: environment, scenario, observed result, how to reproduce.>

## Tradeoffs / risks

- <Known limitation and why it is acceptable, or an alternative rejected. Omit if none.>

## Related PRs

<Sibling / stacked PRs of the same change set, cross-linked.>
```

## Implementation Details

- **The host's CLI owns the create / link calls** — push, create, draft, reviewers, work-item and sibling links. This command orchestrates, it does not invent a host's flags out of thin air.
- **Build the description with a serializer or heredoc** (`python3` + `json`, or a quoted heredoc), never hand-escape — bodies contain backticks, quotes, and newlines. `gh pr edit --body` replaces the whole body, so an additive edit reads the body back, appends, then re-sets.
- **Ground the body in reality**: derive `## Changes` from the diff and `## Testing` from the checks actually run — do not invent test evidence or risks.
- **Title from the branch**: generate a single conventional-commit title that captures the squashed intent of the branch's commits, not a list of them.
- **Coordinates from the remote each run**: resolve owner / repo (or host / project / repo) from `git remote get-url origin`, so the command works across repos and hosts.

## Error Handling

- Not a git repo, or no commits ahead of the base → nothing to open a PR for, say so
- Branch not pushed → push first, `create` has nothing to open against otherwise
- An open PR already exists for the branch → do not duplicate, report it and offer to refresh its description instead
- Host has no supported create path → stop, name the host, and point to wiring its CLI / API
- Branch behind the base, or conflicts against it → `fetch` + `rebase` onto the base and re-push before creating
- Auth failure → defer to the host CLI's auth guidance (token, scope, sign-in — e.g. `gh auth status`)

## Gotchas

- `create` does not push — push the branch first, then create against it.
- Resolve repo coordinates from the remote, never hardcode — the command must work across repos and hosts unchanged.
- A host's `update --description` (e.g. `gh pr edit --body`) replaces the entire body — for an additive note, read it back, append, then re-set rather than blind-overwriting.
- A large diff that is legitimately big (scaffolding, generated code, dependency bumps) is fine — but call it out in the description so the reviewer calibrates instead of skimming the parts that matter.
- Keep the preview-and-confirm default — opening a PR is outward-facing, only bypass with `--yes`.
- An SSH push prints a "post-quantum key exchange" line — it is informational, not an error.
- Not every host has a draft state — `--draft` is a no-op where it is unsupported.

## Examples

```bash
# Open a PR for the current branch against main, description generated, preview first
/xonovex-workflow:pr-create

# Target a release base, open as a draft
/xonovex-workflow:pr-create feat/load-indicator --base release/9 --draft

# Hand-written description, add reviewers, link a work item, skip the prompt
/xonovex-workflow:pr-create --description pr-body.md --reviewers alice,bob --work-item 1234 --yes

# Coordinated change set: cross-link the sibling PRs
/xonovex-workflow:pr-create feat/local-docker-stack --related 18,19,20

# Preview the title and description, create nothing
/xonovex-workflow:pr-create --dry-run
```
