---
description: "Merge useful patterns from one skill into another while preserving the target's structure. Use when the user asks to assimilate, port, or borrow patterns between SKILL.md files. Keywords: SKILL.md, skill assimilate, merge skills, skill porting, guideline reuse."
---

# /xonovex-utility:skill-guidelines-assimilate – Augment Skill with Another Skill

Extracts elements from source skill and integrates into target skill while strictly preserving target's structure, style, and voice.

## Core Workflow

1. Use TodoWrite to track steps
2. Read target/source skills (including reference files)
3. Analyze target's DNA (structure, style, voice, formatting, conventions)
4. Extract source elements (guidelines, examples, patterns, terminology)
5. Filter by aspects/percentage
6. Ask clarifying questions if interactive mode was requested
7. Rewrite source in target's voice, match formatting exactly
8. Update/create reference files
9. Preview or apply
10. Report summary

## Integration Rules

**Preserve (CRITICAL):** Frontmatter, section order, bullet format, voice/tone, code style, reference file structure, spacing, terminology

**Extract from source:** New guidelines, enhanced explanations, code examples, reference topics, complementary patterns

**Style matching:** Match bullet patterns, bold/italic/code usage, sentence structure, vocabulary, whitespace, heading caps

**Approach:** Rewrite in target's voice → insert in existing sections → match format exactly → adapt code style → avoid duplicates

**Percentage scale:** 10-30% critical only, 30-50% important (default), 50-70% comprehensive, 70-100% extensive

**Aspect filtering:** Extract only specified aspects (e.g., "validation" = validation content only)

## Implementation

**Discovery:** Accept SKILL.md paths or names (e.g., `typescript-guidelines` → `.claude/skills/typescript-guidelines/SKILL.md`)

**Analysis:** Parse target structure → analyze bullet/formatting patterns → detect voice → extract style rules → build template

**Extraction:** Parse source → extract examples/patterns → read reference files → filter by aspects

**Integration:** Rewrite in target's voice → insert in existing sections → merge reference files → validate consistency

## Error Handling

- File not found, invalid percentage (10-100), no new content, aspect not found, structure conflict, style detection failed

## Safety

Recommend git commit, never modify frontmatter `name:`/`description:`, preserve all target content (add only), preview before writing, warn if >40% added, abort if style confidence <80%
