---
description: Write or rewrite content to remove AI writing patterns and add human voice
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
  - Skill
argument-hint: "[text-or-file] [--tone <formal|casual|technical>] [--in-place] [--audit]"
---

# /xonovex-utility:content-humanize — Remove AI writing patterns

## Arguments

- `text-or-file` (required): Inline text, a file path, or `-` to read from stdin
- `--tone` (optional): `formal`, `casual`, or `technical`. Default: infer from input
- `--in-place` (optional): Overwrite the source file instead of printing
- `--audit` (optional): Include the "what made it obviously AI" pass in the output

## Delegation

Load the `content-guide` skill (plugin `xonovex-skill-content`) and perform its
**humanize** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
