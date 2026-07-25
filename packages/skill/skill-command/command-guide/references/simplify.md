# simplify: Condense a Verbose Prompt

Reduce verbosity in a prompt file while preserving all functional content. Removes duplication, simplifies examples, makes content generic for cross-project use.

## Goal

- Reduce file length by 30-70% (default target 50%) while maintaining full functionality
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
   - **Usually removable:** advanced features, version control, technical notes; best practices and troubleshooting only after checking they hold no safety rule
4. **Identify Patterns** — scan for project-specific content: paths, domain clusters (3+ related terms), API/service names, industry terminology, redundant sections
5. **Simplify Content** — merge duplicates, reduce examples (4+ → 2-3), condense explanations, shorten output samples (60-70%)
6. **Make Generic** — replace specific paths / domain terms / API names; remove industry context
7. **Preview or Apply** — show diff (`--dry-run`) or write the simplified file
8. **Report** — display line-count reduction and sections modified

## Simplification Rules

### Usually removable

Sections that restate baseline model knowledge, the harness's own documentation, or content already covered elsewhere in the file — typically Advanced Features, Version-Control Integration, Technical Notes, and redundant Troubleshooting. Before deleting Best Practices or Troubleshooting, check whether it is the only home of a safety or destructive-action rule; if so, move that rule into Safety and delete the rest.

### Merge

"What this does" + "Workflow" → "Core Workflow"; Usage + Examples; multiple example subsections → 2-3 cases

### Simplify

reduce examples (4+ → 2-3), condense output samples (60-70%), convert paragraphs to bullets

### Make generic

replace specific paths, domain terms, project names with generic equivalents (`packages/myapp/` → `packages/example/`, `users` → `items`, `MyProjectAPI` → `API`)

### Keep (essential)

metadata block, goal, arguments, core workflow, implementation details, error handling, safety guidelines

## Generalization Patterns

1. **Paths:** neutralize the project segment but keep the path's depth and role (`packages/billing-service/src/` → `packages/example/src/`)
2. **Domain terms:** find 3+ related specialized terms (e.g. `users+orders+payments`) and replace with generic equivalents
3. **Project names:** neutralize the qualifier and keep the role suffix (`StripeClient` → `PaymentClient`, `AcmeUserService` → `UserService`)
4. **Business context:** remove industry-specific workflows, use generic CRUD examples

## Dry-run Output

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

Commit to git first; use `--dry-run`; test after changes; skip if already thin (a Delegation block and no inlined procedure).

## Gotchas

- "Already simplified" usually means <150 lines, but a 200-line file with 60% prose padding still has room — measure prose:bullets ratio
- Generalizing too aggressively (e.g. all paths → `src/`) destroys the example's usefulness — keep enough specificity to learn from
