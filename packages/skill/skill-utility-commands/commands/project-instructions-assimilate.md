---
description: >-
  Augment project instructions with elements from another project's instructions
  while preserving structure and style
model: sonnet
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - TodoWrite
  - AskUserQuestion
argument-hint: >-
  [target-instructions] [source-instructions] [--aspects <aspects>]
  [--percentage <percent>] [--interactive] [--dry-run]
---

# /project-instructions-assimilate – Augment Project Instructions

Extracts organizational patterns from source AGENTS.md/CLAUDE.md and integrates into target while strictly preserving target's structure, style, and project-specific context.

## Arguments

- `target-instructions` (required): Target AGENTS.md/CLAUDE.md (augmented)
- `source-instructions` (required): Source AGENTS.md/CLAUDE.md (provides patterns)
- `--aspects <aspects>` (optional): Focus aspects (e.g., "workflow,structure,integration")
- `--percentage <percent>` (optional): Intensity 10-100 (default: 45)
- `--interactive` (optional): Ask clarifying questions
- `--dry-run` (optional): Preview without modifying

## Core Workflow

1. Use TodoWrite to track steps
2. Read target/source instructions
3. Analyze target's DNA (structure, style, voice, formatting, conventions, project context)
4. Extract source patterns (organizational, workflow, integration, structure, dependencies)
5. Filter by aspects/percentage
6. Ask questions if --interactive
7. Rewrite patterns in target's voice with target's technology names
8. Preview or apply
9. Report summary

## Integration Rules

**Preserve (CRITICAL):** Section order, project tech names (moon/npm/Terraform/etc), paths/directories, command syntax, notation style (arrows/parens), spacing, terminology, all project context

**Extract from source:** Organizational patterns only (section grouping, hierarchy, workflow presentation, integration docs, structure styles)

**Style matching:** Match section presentation, command notation, arrow style, inline details format, heading caps, whitespace, tech vocabulary

**Approach:** Extract patterns NOT content → rewrite with target's tech names → insert in existing sections → match formatting exactly → preserve all project-specific elements

**Project preservation:** Never replace tech names, keep paths/directories, preserve commands, maintain integration docs, keep dependencies unchanged

**Percentage scale:** 10-30% critical only, 30-50% important (default: 45), 50-70% comprehensive, 70-100% extensive

**Aspect filtering:** `workflow` (sequences/delegation), `structure` (directory/hierarchy), `integration` (doc styles), `dependencies` (doc approaches), `commands` (notation styles)

## Examples

```bash
/project-instructions-assimilate services/api/AGENTS.md ../template/AGENTS.md --aspects "workflow"
/project-instructions-assimilate infrastructure/AGENTS.md examples/AGENTS.md --aspects "structure"
/project-instructions-assimilate AGENTS.md ../reference/AGENTS.md --percentage 30 --dry-run
```

## Implementation

**Discovery:** Accept AGENTS.md/CLAUDE.md paths or directories (e.g., `services/api` → `services/api/AGENTS.md`)

**Analysis:** Parse target structure → analyze formatting → detect voice → extract conventions → identify project context → build template

**Extraction:** Identify organizational patterns → extract workflow presentation → find integration approaches → filter by aspects

**Integration:** Extract patterns only (not content) → rewrite with target's tech names → insert in existing sections → match formatting → validate project-specificity

## Error Handling

File not found, invalid percentage (10-100), no new patterns, aspect not found, incompatible structure, style detection failed

## Safety

Recommend git commit, never modify project tech names/paths/commands, preserve all target content, use `--dry-run`, warn if >30% added, abort if style confidence <85% or project context at risk, only modify CLAUDE.md when explicitly specified
