---
description: >-
  Augment an existing slash command with elements from another slash command
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
  [target-command] [source-command] [--aspects <aspects>] [--percentage
  <percent>] [--interactive] [--dry-run]
---

# /slashcommand-assimilate – Augment Slash Command

Extracts elements from source command and integrates into target command while strictly preserving target's structure, style, and voice.

## Arguments

- `target-command` (required): Target command file (augmented)
- `source-command` (required): Source command file (provides elements)
- `--aspects <aspects>` (optional): Focus aspects (e.g., "workflow,validation,error-handling")
- `--percentage <percent>` (optional): Intensity 10-100 (default: 50)
- `--interactive` (optional): Ask clarifying questions
- `--dry-run` (optional): Preview without modifying

## Core Workflow

1. Use TodoWrite to track steps
2. Read target/source commands
3. Analyze target's DNA (structure, style, voice, formatting, conventions)
4. Extract source elements (workflow, arguments, validation, error handling, examples)
5. Filter by aspects/percentage
6. Ask questions if --interactive
7. Rewrite source in target's voice, match formatting exactly
8. Preview or apply
9. Report summary

## Integration Rules

**Preserve (CRITICAL):** Frontmatter, section order, formatting (bullets/numbers), voice/tone, code style, argument format, example structure, spacing

**Extract from source:** Workflow steps, arguments/flags, validation, error handling, safety guidelines, examples

**Style matching:** Match workflow format, argument style, example format, heading caps, whitespace, vocabulary

**Approach:** Rewrite in target's voice → insert in existing sections → match format exactly → adapt examples → avoid duplicates

**Percentage scale:** 10-30% critical only, 30-50% important (default), 50-70% comprehensive, 70-100% extensive

**Aspect filtering:** `workflow` (steps/sequence), `arguments` (patterns/validation), `error-handling` (cases/messages), `examples` (use cases), `validation` (rules/checks), `safety` (guidelines/warnings)

## Examples

```bash
/slashcommand-assimilate git-commit.md git-feature-validate.md --aspects "validation"
/slashcommand-assimilate plan-create.md plan-tdd-create.md --aspects "workflow"
/slashcommand-assimilate code-align-check.md code-quality-improve.md --percentage 25 --dry-run
```

## Implementation

**Discovery:** Accept .md paths or names (e.g., `git-commit` → `.claude/commands/git-commit.md`)

**Analysis:** Parse target structure → analyze formatting → detect voice → extract conventions → build template

**Extraction:** Parse source → extract workflow/arguments/validation → collect examples → filter by aspects

**Integration:** Rewrite in target's voice → insert in existing sections → merge workflows → validate consistency

## Error Handling

File not found, invalid percentage (10-100), no new content, aspect not found, incompatible commands, style detection failed

## Safety

Recommend git commit, never modify frontmatter without confirmation, preserve all target content (add only), use `--dry-run`, warn if >40% added, abort if style confidence <80%
