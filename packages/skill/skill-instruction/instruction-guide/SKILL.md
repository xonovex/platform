---
name: instruction-guide
description: "Use when authoring or reviewing AGENTS.md / CLAUDE.md project-instruction files. Triggers on edits to `AGENTS.md` or `CLAUDE.md` and on prompts about structure, brevity, dry-run review, syncing with directory state, or assimilating instruction patterns across projects, even when the user doesn't say 'instructions'. Skip the focused automations — use instructions-init / instructions-sync / instructions-simplify / instructions-consolidate / instructions-assimilate when the user names a specific step."
---

# Project Instruction Guidelines

## Core Principles

- **Preserve Project Context** - Never modify technology names, paths, or commands
- **Match Style and Voice** - Maintain target file's formatting and terminology
- **Structure Integrity** - Keep section order and hierarchy intact
- **Safe Modifications** - Use dry-run to preview changes before applying

## Common Operations

- **Assimilate** - Extract organizational patterns from source and integrate into target
- **Simplify** - Reduce verbosity while preserving structure and workflows
- **Sync** - Update directory listings to reflect current filesystem state

## Progressive Disclosure

- **Assimilate patterns** - Augment instructions with patterns from another file, see [reference/assimilate.md](reference/assimilate.md)
- **Simplify instructions** - Reduce verbosity in AGENTS.md/CLAUDE.md, see [reference/simplify.md](reference/simplify.md)
- **Sync with filesystem** - Update directory structure to current state, see [reference/sync.md](reference/sync.md)
