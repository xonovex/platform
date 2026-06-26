---
description: >-
  Create a comprehensive, multi-language travel guide for a specified topic or
  location
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
  [topic] [subject] [--path <path>] [--lang <en,nl>] [--research-only] [--slug
  <slug>]
---

# /xonovex-utility:content-travelguide-add — Comprehensive Travel Guide Generator

## Arguments

- `topic` (required): The type of travel guide to create (e.g., "Port", "City", "Museum").
- `subject` (required): The specific subject of the travel guide (e.g., "Barcelona", "The Louvre").
- `--path` (required): The destination directory for the generated files.
- `--lang` (optional): Comma-separated list of languages (e.g., `en,nl`). Defaults to `en`.
- `--research-only` (optional): Conduct research and present findings without creating files.
- `--slug` (optional): Custom slug for the generated files. If not provided, the slug is derived from the subject.

## Delegation

Load the `content-guide` skill (plugin `xonovex-skill-content`) and perform its
**travelguide-add** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
