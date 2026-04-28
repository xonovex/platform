---
description: "Sync AGENTS.md files with the actual current directory structure and codebase state. Use when the user asks to refresh AGENTS.md, update instructions to match the codebase, or fix stale documentation. Keywords: AGENTS.md, sync, refresh, stale instructions, directory structure update, instructions sync."
---

# /xonovex-utility:instructions-sync – Sync AGENTS.md with Current State

Update AGENTS.md files to reflect current directory structure, files, and configuration state. Maintain project-specific technology names.

## Core Workflow

1. Use TodoWrite to track each file
2. Discover: scan subdirectories, read package.json/config files, identify patterns
3. Analyze: parse Subdirectories section, extract workflows, preserve Integration Points
4. Sync: add new directories, remove deleted ones, update patterns
5. Update workflows (if requested): sync operations, update delegation chains, keep technology names
6. Validate and report: check broken references, show diff

## Discovery

**Subdirectories:** Scan 1 level deep, exclude (node_modules, .git, build, dist, coverage, .artifacts), detect `<name>/` patterns

**Files:** Identify common files (package.json, moon.yml, CMakeLists.txt, Dockerfile), list configs in parentheses

**Commands:** Extract from package.json/config files, format with backticks, preserve tool names

**Git History:** Check recent commits for directory/file additions not reflected in AGENTS.md

## Sync Strategy

- **Add:** New directories → create bullet with pattern and file examples
- **Update:** Existing directories → refresh patterns with current state
- **Remove:** Deleted directories → remove bullet, warn
- **Preserve:** Integration Points, headings, descriptions, technology names

## Example

**Before:** `environments/`: Environment configs

**After:** `environments/<name>/`: Environment configs (main.tf, vars, backend.sh) - local, staging, production

## Implementation

- Concise format: inline parentheses for files, dash-separated examples
- Pattern detection: recognize `<name>/` for similar subdirectories
- Technology preservation: keep actual names (moon, npm, Terraform, Flux)
- Diff output: show added (+), removed (-), updated (~)

## Error Handling

- File not found: verify path
- No AGENTS.md (--all): report missing locations
- No changes: report "Already synchronized"
- Parse errors: warn, preserve manual content

## Safety

Preview before writing, preserve manual descriptions/Integration Points/section structure, keep technology references. Only sync AGENTS.md files; never modify CLAUDE.md unless explicitly specified.
