---
name: skill-guidelines
description: >-
  Trigger on `.claude/skills/` directory work, skill documentation creation. Use when creating or managing guideline skills. Apply for extracting patterns, assimilating skills, simplifying verbose skills. Keywords: skill creation, progressive disclosure, SKILL.md, details, pattern extraction, skill assimilation, bullet format.
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

## Progressive Disclosure

- Read [reference/create.md](reference/create.md) - When creating skill from documentation or URL
- Read [reference/extract.md](reference/extract.md) - When extracting patterns from codebase
- Read [reference/assimilate.md](reference/assimilate.md) - When augmenting skill with another skill's elements
- Read [reference/simplify.md](reference/simplify.md) - When condensing verbose skill to bullet format
