---
description: >-
  Augment an existing skill with elements from another skill while preserving
  structure and style
model: sonnet
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
  - AskUserQuestion
argument-hint: >-
  [target-skill] [source-skill] [--aspects <aspects>] [--percentage <percent>]
  [--interactive] [--dry-run]
---

# /skill-guidelines-assimilate – Augment Skill with Another Skill

Extracts elements from source skill and integrates into target skill while strictly preserving target's structure, style, and voice.

## Arguments

- `target-skill` (required): Target skill file/directory (augmented)
- `source-skill` (required): Source skill file/directory (provides elements)
- `--aspects <aspects>` (optional): Focus aspects (e.g., "error-handling,validation")
- `--percentage <percent>` (optional): Intensity 10-100 (default: 50)
- `--interactive` (optional): Ask clarifying questions
- `--dry-run` (optional): Preview without modifying

## Core Workflow

1. Use TodoWrite to track steps
2. Read target/source skills (including detail files)
3. Analyze target's DNA (structure, style, voice, formatting, conventions)
4. Extract source elements (guidelines, examples, patterns, terminology)
5. Filter by aspects/percentage
6. Ask questions if --interactive
7. Rewrite source in target's voice, match formatting exactly
8. Update/create detail files
9. Preview or apply
10. Report summary

## Integration Rules

**Preserve (CRITICAL):** Frontmatter, section order, bullet format, voice/tone, code style, detail file structure, spacing, terminology

**Extract from source:** New guidelines, enhanced explanations, code examples, detail topics, complementary patterns

**Style matching:** Match bullet patterns, bold/italic/code usage, sentence structure, vocabulary, whitespace, heading caps

**Approach:** Rewrite in target's voice → insert in existing sections → match format exactly → adapt code style → avoid duplicates

**Percentage scale:** 10-30% critical only, 30-50% important (default), 50-70% comprehensive, 70-100% extensive

**Aspect filtering:** Extract only specified aspects (e.g., `--aspects "validation"` = validation content only)

## Examples

```bash
/skill-guidelines-assimilate typescript-guidelines zod-guidelines --aspects "validation"
/skill-guidelines-assimilate react-guidelines vue-guidelines --percentage 25 --dry-run
/skill-guidelines-assimilate python-guidelines typescript-guidelines --interactive
```

## Implementation

**Discovery:** Accept SKILL.md paths or names (e.g., `typescript-guidelines` → `.claude/skills/typescript-guidelines/SKILL.md`)

**Analysis:** Parse target structure → analyze bullet/formatting patterns → detect voice → extract style rules → build template

**Extraction:** Parse source → extract examples/patterns → read detail files → filter by aspects

**Integration:** Rewrite in target's voice → insert in existing sections → merge detail files → validate consistency

## Error Handling

- File not found, invalid percentage (10-100), no new content, aspect not found, structure conflict, style detection failed

## Safety

Recommend git commit, never modify frontmatter `name:`/`description:`, preserve all target content (add only), use `--dry-run`, warn if >40% added, abort if style confidence <80%
