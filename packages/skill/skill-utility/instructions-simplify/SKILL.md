---
description: "Reduce AGENTS.md/CLAUDE.md verbosity while preserving structure and technology names. Use when the user asks to simplify, condense, or shorten project instructions. Keywords: AGENTS.md, CLAUDE.md, simplify, condense, reduce verbosity, instructions cleanup."
---

# /xonovex-utility:instructions-simplify – Simplify project instruction files

Reduce AGENTS.md/CLAUDE.md verbosity by 40-50% while preserving structure, workflows, and project-specific technology names.

## Core Workflow

1. Use TodoWrite to track steps
2. Read file and measure baseline
3. Analyze sections (Structure, Subdirectories, Workflow, Integration Points)
4. Apply simplification rules
5. Preview or write changes
6. Report metrics

## Simplification Rules

**Remove:** Verbose descriptions, explanatory prose, redundant introductions, @docs references, duplicates, code blocks (convert to inline arrow notation)

**Condense:** Multi-line bullets → single line with inline details; workflows → essential steps; task delegation → arrow notation; directory structures → inline patterns with parentheses

**Keep:** Section headings, dependencies, command examples with actual tool names, file/directory patterns, integration points, technology names (moon, Terraform, Flux, npm)

## Example

**Before:** "This directory contains environment-specific configurations for different deployment targets. Each environment has its own main.tf and tfvars files."

**After:** Environment configs (main.tf, vars, backend.sh) - local, staging

## Implementation

- Skip files <15 lines
- Preserve section hierarchy
- Convert verbose descriptions to inline parenthetical details
- Use → for workflow chains
- Keep actual technology names

## Error Handling

- File not found: verify path
- Already minimal: skip if <15 lines
- Invalid reduction: must be 30-60%
- Not AGENTS.md/CLAUDE.md: warn and confirm

## Safety

Preserve headings/patterns/commands, preview before writing, never remove integration points or technology references. Only modify CLAUDE.md when explicitly specified.
