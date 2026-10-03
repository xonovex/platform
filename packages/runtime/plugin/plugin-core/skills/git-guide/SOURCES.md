# Sources

## Git reference documentation

- **URLs:**
  - https://git-scm.com/docs
  - https://git-scm.com/docs/git-worktree
  - https://git-scm.com/docs/git-rebase
- **Last reviewed:** 2026-07-19
- **Used for:** `SKILL.md` and all `references/`; commits, merges, rebases, conflict handling, worktree lifecycle, and repository-state inspection.
- **References:** all
- **Aspects extracted:** Command semantics and flags come from the Git reference. Conventional-commit wording and safety defaults are Xonovex conventions.

## Xonovex worktree operation boundaries

- **Provenance:** Repository-original preview/apply, separation, and recovery conventions for worktree operations
- **Last reviewed:** 2026-07-24
- **References:** references/worktree-create.md, references/worktree-merge.md, references/worktree-abandon.md, references/worktree-cleanup.md
- **Aspects extracted:** Create, merge, and cleanup default to preview; abandon is read-only; merge preserves workspace resources; cleanup resolves exact targets and requires separate force or remote-deletion authority.
