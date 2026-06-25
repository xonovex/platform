---
name: instruction-guide
description: "Use when authoring, reviewing, initializing, syncing, simplifying, consolidating, or assimilating AGENTS.md project-instruction files. Triggers on edits to `AGENTS.md` and on prompts about bootstrapping fresh instructions, refreshing stale docs to match the filesystem, condensing verbose files, deduping across the repo, or porting organizational patterns from another project — even when the user doesn't say 'instructions' or 'AGENTS.md'."
---

# Project Instruction Guidelines

Author, refresh, simplify, consolidate, and port AGENTS.md project-instruction files while preserving technology names, paths, and project context.

## Core Principles

- **Open standard** — AGENTS.md is the cross-harness format read by many agent harnesses; keep content harness-neutral
- **README for agents** — operational detail (commands, conventions, gotchas) belongs in AGENTS.md; human-onboarding prose belongs in README.md
- **Teammate heuristic** — "anything you'd tell a new teammate" belongs here; anything you'd tell a casual visitor belongs in README
- **Nested precedence** — for monorepos, the AGENTS.md closest to the edited file wins; place subproject-specific guidance in subdirectories
- **Commands are executable** — agents auto-run programmatic checks listed in AGENTS.md; treat command examples as instructions, not just docs
- **Preserve Project Context** — never modify technology names, paths, or commands
- **Match Style and Voice** — maintain the target file's formatting and terminology
- **Structure Integrity** — keep section order and hierarchy intact
- **Safe Modifications** — preview changes (dry-run) before applying
- **Signal Over Volume** — keep only non-obvious content; anything inferable from `package.json` + source belongs in the code, not the doc
- **Living documentation** — update alongside project changes (sync operation)

## Gotchas

- AGENTS.md files that restate code are worse than no doc — they rot when the code changes
- Tech names look like noise but are load-bearing; never substitute generics for `moon` / `Terraform` / `npm`
- Root AGENTS.md is the project entry-point doc — never auto-delete even if it looks thin
- Manual descriptions in Subdirectories often encode info the filesystem can't (purpose, owner, status) — preserve verbatim during sync
- Nested AGENTS.md aren't duplicates — they're scope-specific overrides; never merge a subproject's file into the root
- Build / test commands listed in AGENTS.md will be **executed** by agents; broken commands waste agent turns

## Operations

- **Init** — bootstrap a fresh AGENTS.md for a directory — see [references/init.md](references/init.md)
- **Sync** — refresh AGENTS.md to match current directory state — see [references/sync.md](references/sync.md)
- **Simplify** — reduce verbosity while preserving structure and workflows — see [references/simplify.md](references/simplify.md)
- **Consolidate** — dedupe across the repo and standardize bullet-list format — see [references/consolidate.md](references/consolidate.md)
- **Merge / assimilate** — port organizational patterns from another project — see [references/merge.md](references/merge.md)

## Progressive Disclosure

- Read [references/init.md](references/init.md) - Load when bootstrapping a fresh AGENTS.md for a directory
- Read [references/sync.md](references/sync.md) - Load when refreshing AGENTS.md to match current filesystem state
- Read [references/simplify.md](references/simplify.md) - Load when condensing a verbose AGENTS.md
- Read [references/consolidate.md](references/consolidate.md) - Load when cleaning up scattered AGENTS.md across a repo
- Read [references/merge.md](references/merge.md) - Load when porting organizational patterns from another project's AGENTS.md
