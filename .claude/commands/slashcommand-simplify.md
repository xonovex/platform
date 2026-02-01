---
description: >-
  Automatically simplify slash command files by reducing verbosity while
  maintaining functionality
model: sonnet
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - TodoWrite
argument-hint: "[command-file] [--dry-run] [--target-reduction <percent>]"
---

# /slashcommand-simplify – Simplify Slash Command Documentation

Automatically reduces verbosity in slash command files while preserving all functional content. Removes duplication, simplifies examples, and makes content generic for cross-project use.

## Goal

- Reduce file length by 40-60% while maintaining full functionality
- Remove duplicate and redundant sections
- Simplify verbose examples and explanations
- Convert project-specific content to generic equivalents
- Preserve essential functionality, arguments, and implementation guidance

## Usage

```bash
# Simplify a command file
/slashcommand-simplify .claude/commands/my-command.md

# Preview changes without modifying
/slashcommand-simplify .claude/commands/my-command.md --dry-run

# Target specific reduction percentage
/slashcommand-simplify .claude/commands/my-command.md --target-reduction 50
```

## Arguments

- `command-file` (required): Path to slash command file to simplify
- `--dry-run` (optional): Preview changes without writing file
- `--target-reduction` (optional): Target reduction percentage (default: 50, range: 30-70)

## Core Workflow

1. **Setup** - Use TodoWrite to track simplification steps
2. **Read File** - Load the command file using Read tool
3. **Analyze Structure** - Identify sections:
   - Essential (keep): Front matter, goal, arguments, core workflow, implementation details, error handling
   - Simplifiable (reduce): Examples, explanations, output samples
   - Removable (delete): Advanced features, best practices, version control, technical notes
4. **Identify Patterns** - Scan for project-specific content: paths (`packages/myapp/`), domain clusters (3+ related terms), API/service names, industry terminology, redundant sections
5. **Simplify Content**: Merge duplicates, reduce examples (4+ → 2-3), condense explanations, shorten output samples (60-70%)
6. **Make Generic**: Replace specific paths/domain terms/API names, remove industry context
7. **Preview or Apply** - Show diff (--dry-run) or write simplified file (Edit/Write tool)
8. **Report** - Display line count reduction and sections modified

## Simplification Rules

**Remove entirely:** Advanced features, best practices, version control integration, technical notes, troubleshooting (if redundant)

**Merge:** "What this does" + "Workflow" → "Core Workflow"; Usage + Examples; multiple example subsections → 2-3 cases

**Simplify:** Reduce examples (4+ → 2-3), condense output samples (60-70%), convert paragraphs to bullets

**Make generic:** Replace specific paths (`packages/myapp/` → `packages/example/`), domain terms (`products` → `items`), project names (`MyProjectAPI` → `API`)

**Keep (essential):** Front matter, goal, arguments, core workflow, implementation details, error handling, safety guidelines

## Implementation Details

**Analysis:** Count lines, calculate target (current × reduction%), parse sections with `/^##?\s+(.+)$/gm`, categorize as Essential/Simplifiable/Removable

**Simplification:** Remove categorized sections, merge duplicates, condense examples (2-3 kept, 10-15 lines), convert paragraphs to bullets

**Generalization:**

1. **Paths**: Replace `/packages\/[\w-]+\//g` → `packages/example/`
2. **Domain terms**: Find 3+ related specialized terms (e.g., users+orders+payments), replace with generic equivalents
3. **Project names**: Replace `/(\w+)(API|Service|Database|Client)/g` → `MyProject$2`
4. **Business context**: Remove industry-specific workflows, use generic CRUD examples

**Validation:** Verify front matter/arguments/workflow preserved, measure reduction

**Dry-run output:**

```
=== Preview ===
File: example.md | 503 → 252 lines (50%)
Remove: Advanced Features (45), Best Practices (32)
Merge: "What this does" + "Workflow" → "Core Workflow"
Simplify: Examples 4→3 (-60), Output samples (-45)
Generic: 8 paths, 15 domain terms
Result: 208 lines (59%)
```

## Examples

```bash
/slashcommand-simplify .claude/commands/my-command.md
/slashcommand-simplify .claude/commands/my-command.md --dry-run
/slashcommand-simplify .claude/commands/my-command.md --target-reduction 60
```

## Error Handling

**Errors:** File not found (verify path), not a slash command (check front matter), already simplified (lower target), cannot achieve target (reduce %)

**Safety:** Commit to git first, use `--dry-run`, test after changes, skip if <150 lines
