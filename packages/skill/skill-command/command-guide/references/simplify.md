# simplify: Condense a Verbose Prompt

Reduce verbosity in a prompt file while preserving all functional content. Removes duplication, simplifies examples, makes content generic for cross-project use.

## Goal

- Reduce file length by 40-60% while maintaining full functionality
- Remove duplicate and redundant sections
- Simplify verbose examples and explanations
- Convert project-specific content to generic equivalents
- Preserve essential functionality, arguments, and implementation guidance

## Arguments

- `prompt-file` (required) — path to the prompt file to simplify
- `--dry-run` (optional) — preview without writing
- `--target-reduction` (optional) — target reduction percentage (default 50, range 30-70)

## Core Workflow

1. **Setup** — track steps in a task list
2. **Read File** — load the prompt file
3. **Analyze Structure** — classify each section:
   - **Essential (keep):** metadata block, goal, arguments, core workflow, implementation details, error handling
   - **Simplifiable (reduce):** examples, explanations, output samples
   - **Removable (delete):** advanced features, best practices, version control, technical notes
4. **Identify Patterns** — scan for project-specific content: paths, domain clusters (3+ related terms), API/service names, industry terminology, redundant sections
5. **Simplify Content** — merge duplicates, reduce examples (4+ → 2-3), condense explanations, shorten output samples (60-70%)
6. **Make Generic** — replace specific paths / domain terms / API names; remove industry context
7. **Preview or Apply** — show diff (`--dry-run`) or write the simplified file
8. **Report** — display line-count reduction and sections modified

## Simplification Rules

**Remove entirely:** advanced features, best practices, version-control integration, technical notes, troubleshooting (if redundant)

**Merge:** "What this does" + "Workflow" → "Core Workflow"; Usage + Examples; multiple example subsections → 2-3 cases

**Simplify:** reduce examples (4+ → 2-3), condense output samples (60-70%), convert paragraphs to bullets

**Make generic:** replace specific paths, domain terms, project names with generic equivalents (`packages/myapp/` → `packages/example/`, `users` → `items`, `MyProjectAPI` → `API`)

**Keep (essential):** metadata block, goal, arguments, core workflow, implementation details, error handling, safety guidelines

## Implementation Details

**Analysis:** count lines, calculate target (current × reduction%), parse sections, categorize as Essential/Simplifiable/Removable

**Simplification:** remove categorized sections, merge duplicates, condense examples (2-3 kept, 10-15 lines), convert paragraphs to bullets

**Generalization:**

1. **Paths:** replace `/packages\/[\w-]+\//g` → `packages/example/`
2. **Domain terms:** find 3+ related specialized terms (e.g. `users+orders+payments`) and replace with generic equivalents
3. **Project names:** replace `(\w+)(API|Service|Database|Client)` → `MyProject$2`
4. **Business context:** remove industry-specific workflows, use generic CRUD examples

**Validation:** verify metadata block / arguments / workflow preserved; measure reduction

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

## Error Handling

- File not found → verify path
- Not a prompt (no metadata block or wrong format for the harness) → check
- Already simplified → lower target
- Cannot achieve target → reduce %

## Safety

Commit to git first; use `--dry-run`; test after changes; skip if <150 lines.

## Gotchas

- "Already simplified" usually means <150 lines, but a 200-line file with 60% prose padding still has room — measure prose:bullets ratio
- Stripping "Best Practices" sections sometimes removes the only place a critical safety rule lives — diff carefully
- Generalizing too aggressively (e.g. all paths → `src/`) destroys the example's usefulness — keep enough specificity to learn from
