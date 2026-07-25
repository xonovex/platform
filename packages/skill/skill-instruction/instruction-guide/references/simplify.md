# simplify: Condense Verbose AGENTS.md

Reduce AGENTS.md verbosity by 30-60% (default target 45%) while preserving structure, workflows, and project-specific technology names.

## Core Workflow

Measure baseline → analyze sections (Structure, Subdirectories, Workflow, Integration Points) → apply simplification rules → preview or write → report metrics. Skip files <15 lines.

## Simplification Rules

### Remove

verbose descriptions, explanatory prose, redundant introductions, @docs references, duplicates, code blocks (convert to inline arrow notation)

### Condense

multi-line bullets → single line with inline details; workflows → essential steps; task delegation → arrow notation; directory structures → inline patterns with parentheses

### Keep

section headings, dependencies, command examples with actual tool names, file/directory patterns, integration points, technology names (`moon`, `Terraform`, `Flux`, `npm`, etc.)

## Example

### Before

"This directory contains environment-specific configurations for different deployment targets. Each environment has its own main.tf and tfvars files."

### After

Environment configs (`main.tf`, vars, `backend.sh`) - local, staging

## Gotchas

- Tech names look like noise but are essential — don't strip `moon` / `Terraform` / `Flux` even if a generic verb works
- Removing a code block looks like progress until you realize it was the only place a command was documented
- "Already minimal" usually means <15 lines, but a 30-line file dense with gotchas is also already minimal — measure signal, not just length
