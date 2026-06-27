# sync: Refresh AGENTS.md to Match Current State

Update AGENTS.md files to reflect current directory structure, files, and configuration state. Maintain project-specific technology names.

## Core Workflow

1. Track each file in a task list
2. **Discover:** scan subdirectories, read `package.json` / config files, identify patterns
3. **Analyze:** parse Subdirectories section, extract workflows, preserve Integration Points
4. **Sync:** add new directories, remove deleted ones, update patterns
5. **Update workflows** (if requested): sync operations, update delegation chains, keep technology names
6. **Validate and report:** check broken references, show diff

## Discovery

### Subdirectories

scan 1 level deep, exclude `node_modules`, `.git`, `build`, `dist`, `coverage`, `.artifacts`; detect `<name>/` patterns

### Files

identify common files (`package.json`, `moon.yml`, `CMakeLists.txt`, `Dockerfile`), list configs in parentheses

### Commands

extract from `package.json` / config files, format with backticks, preserve tool names

### Git history

check recent commits for directory/file additions not reflected in AGENTS.md

## Sync Strategy

- **Add:** new directories → create bullet with pattern and file examples
- **Update:** existing directories → refresh patterns with current state
- **Remove:** deleted directories → remove bullet, warn
- **Preserve:** Integration Points, headings, descriptions, technology names
- **Verify commands:** build / test / lint commands listed in AGENTS.md will be **executed** by agents — confirm they still work, update or remove stale ones
- **Respect nested scope:** each AGENTS.md syncs against its own directory tree, not the whole repo — the closest one wins, so don't pull subproject details into the root

## Example

### Before

`` `environments/` ``: Environment configs

### After

`` `environments/<name>/` ``: Environment configs (`main.tf`, vars, `backend.sh`) - local, staging, production

## Implementation

- Concise format: inline parentheses for files, dash-separated examples
- Pattern detection: recognize `<name>/` for similar subdirectories
- Technology preservation: keep actual names (`moon`, `npm`, `Terraform`, `Flux`)
- Diff output: show added (`+`), removed (`-`), updated (`~`)

## Error Handling

- File not found: verify path
- No AGENTS.md (`--all`): report missing locations
- No changes: report "Already synchronized"
- Parse errors: warn, preserve manual content

## Safety

Preview before writing; preserve manual descriptions, Integration Points, and section structure; keep technology references.

## Gotchas

- Removed directories that come back later look like a sync churn loop — confirm git history before deleting bullets
- Manual descriptions in Subdirectories often encode information the filesystem can't tell you (purpose, owner, status) — preserve verbatim
- A directory that exists in the filesystem but isn't worth documenting (e.g. tooling cache) shouldn't get auto-added — match the existing skip list
- A stale `npm test` / `cargo build` / `pytest` command will be executed by agents and waste turns — re-run listed commands during sync and update or remove broken ones
- Nested AGENTS.md files sync independently — pulling a subproject's content into the root violates closest-wins precedence
