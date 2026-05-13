# simplify: Condense Verbose AGENTS.md

Reduce AGENTS.md verbosity by 40-50% while preserving structure, workflows, and project-specific technology names.

## Core Workflow

1. Track steps in a task list
2. Read file and measure baseline
3. Analyze sections (Structure, Subdirectories, Workflow, Integration Points)
4. Apply simplification rules
5. Preview or write changes
6. Report metrics

## Simplification Rules

**Remove:** verbose descriptions, explanatory prose, redundant introductions, @docs references, duplicates, code blocks (convert to inline arrow notation)

**Condense:** multi-line bullets → single line with inline details; workflows → essential steps; task delegation → arrow notation; directory structures → inline patterns with parentheses

**Keep:** section headings, dependencies, command examples with actual tool names, file/directory patterns, integration points, technology names (`moon`, `Terraform`, `Flux`, `npm`, etc.)

## Example

**Before:** "This directory contains environment-specific configurations for different deployment targets. Each environment has its own main.tf and tfvars files."

**After:** Environment configs (`main.tf`, vars, `backend.sh`) - local, staging

## Implementation

- Skip files <15 lines
- Preserve section hierarchy
- Convert verbose descriptions to inline parenthetical details
- Use `→` for workflow chains
- Keep actual technology names

## Error Handling

- File not found: verify path
- Already minimal: skip if <15 lines
- Invalid reduction: must be 30-60%
- Not AGENTS.md: warn and confirm

## Safety

Preserve headings/patterns/commands; preview before writing; never remove integration points or technology references.

## Gotchas

- Tech names look like noise but are load-bearing — don't strip `moon` / `Terraform` / `Flux` even if a generic verb works
- Removing a code block looks like progress until you realize it was the only place a command was documented
- "Already minimal" usually means <15 lines, but a 30-line file dense with gotchas is also already minimal — measure signal, not just length
