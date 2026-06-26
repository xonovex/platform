---
description: Auto-curate latest news stories on a topic and generate bilingual content
allowed-tools:
  - WebSearch
  - WebFetch
  - Write
  - Read
  - Bash
  - Glob
  - Grep
  - TodoWrite
  - Skill
argument-hint: >-
  [topic] [--path <path>] [--lang <en,nl>] [--days <days>] [--max <max>] [--slug
  <slug>]
---

# /xonovex-utility:content-news-add — Auto-curate latest news stories

## Arguments

- `topic` (required): The subject to search news for.
- `--path` (required): The destination directory for the generated files.
- `--lang` (optional): Comma-separated list of languages (e.g., `en,nl`). Defaults to `en`.
- `--days` (optional): Number of days to look back for news. Defaults to `7`.
- `--max` (optional): Maximum number of stories to generate. Defaults to `3`.
- `--slug` (optional): Custom slug for the generated files. If not provided, the slug is derived from the title.

## Delegation

Load the `content-guide` skill (plugin `xonovex-skill-content`) and perform its
**news-add** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
