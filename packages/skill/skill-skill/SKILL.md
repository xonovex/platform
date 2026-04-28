---
name: skill-guidelines
description: "Guidelines for authoring SKILL.md files with progressive disclosure. Apply when editing files under `.claude/skills/` or creating new skills. Use for pattern extraction, skill assimilation, simplification, bullet format. Keywords: skill creation, SKILL.md, progressive disclosure, reference files, pattern extraction, skill assimilation, bullet format, guideline skill."
---

# Skill Guidelines Management

## Core Principles

- **Progressive Disclosure** - SKILL.md contains essentials, reference/\* contains examples
- **Project Independence** - Remove project-specific paths, names, domains
- **Bullet Format** - `- **Rule** - Brief 5-10 word how-to, see [reference/file.md](reference/file.md)`
- **Style Consistency** - Match existing skill patterns in structure and voice

## Skill Structure

- **SKILL.md** - Frontmatter, essentials (3-7 bullets), one example, progressive disclosure links
- **reference/\*.md** - Guideline, rationale, how to apply, examples, counter-examples

## Operations

- **Create from document** - Extract guidelines from URLs or files, see [reference/create.md](reference/create.md)
- **Extract from codebase** - Analyze code patterns and project instructions, see [reference/extract.md](reference/extract.md)
- **Assimilate skills** - Merge elements from one skill into another, see [reference/assimilate.md](reference/assimilate.md)
- **Simplify skills** - Condense verbose SKILL.md to bullet list, see [reference/simplify.md](reference/simplify.md)
- **Simplify references** - Merge overlapping sections in reference files, see [reference/simplify-reference.md](reference/simplify-reference.md)

## Progressive Disclosure

- Read [reference/create.md](reference/create.md) - When creating skill from documentation or URL
- Read [reference/extract.md](reference/extract.md) - When extracting patterns from codebase
- Read [reference/assimilate.md](reference/assimilate.md) - When augmenting skill with another skill's elements
- Read [reference/simplify.md](reference/simplify.md) - When condensing verbose skill to bullet format
- Read [reference/simplify-reference.md](reference/simplify-reference.md) - When condensing verbose reference files
