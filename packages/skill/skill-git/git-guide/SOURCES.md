# Sources

Most of this skill encodes this repo's own git workflow (conventional commits, the
worktree create / commit / validate / merge / cleanup / abandon lifecycle) and is
repo-original — those reference files have no upstream and are expected to show as
"uncovered" in the source audit.

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Core Principles, Strategy, Gotchas
  - The branching-strategy guidance only (trunk-based vs long-lived branches)
- **Aspects extracted:**
  - "Moving away from GitFlow" — the failure modes of long-lived branches (growing merge conflicts that deter refactors, ever-larger batched PRs, late discovery of design flaws, no shared runnable view of the product, review latency gating throughput); trunk-based alternative: everyone commits small chunks to trunk at least daily, squash + rebase each push for linear searchable history, local branches as ephemeral undo, async post-merge review (fix-forward or revert), feature flags to develop a replacement in parallel and ship it dark, and periodic system-level reviews → `references/branching-strategy.md`

## Trunk-based development prior art

- **URLs:**
  - Trunk-based development — https://trunkbaseddevelopment.com/
  - Martin Fowler, "Patterns for Managing Source Code Branches" — https://martinfowler.com/articles/branching-patterns.html
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Strategy
  - Corroborating the integration-frequency / short-lived-branch / feature-flag model
- **Aspects extracted:**
  - Continuous integration to a shared trunk, short-lived branches, keep-trunk-releasable → `references/branching-strategy.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
