# code-align: Research Code Alignment Between Similar Implementations

Analyze two similar implementations to identify structural differences, inconsistencies, and opportunities for alignment. Generates a research report. Does **not** create plans or make changes — run `plan-create` afterward.

## Core Workflow

**Delegate codebase analysis to read-only search agents where available; otherwise use grep/find/file-read directly. Stay in research mode.**

1. **Discovery & loading** — resolve files from paths or globs, read both file contents, detect code type (API route, processor, etc.)
2. **Structural analysis** — compare:
   - Imports & dependencies (matching imports, missing imports, import order)
   - Interfaces & types (definitions, properties, type consistency)
   - Functions & exports (signatures, helpers, parameters)
   - Code structure (control flow, error handling, logging, naming)
   - Logic & algorithms (processing logic, edge cases, bugs)
   - Configuration (default values, constants)
3. **Alignment report** — summary, critical differences, structural differences, missing features, prioritized recommendations
4. **Interactive mode** (if requested) — ask about reference implementation, control-flow preference, default values, missing features

## Example Output

```
=== Alignment Report ===
Overall Alignment: 72% aligned

Critical Differences:
  [X] File2 missing error handling in process()
  [X] File1 uses deprecated API vs File2 modern approach

Structural Differences:
  [WARN] Different control flow (if/else vs switch)
  [WARN] Different logging (console.log vs logger.info)
  [WARN] Inconsistent defaults (maxItems: 100 vs 200)

Missing Features:
  [+] File1 has metrics tracking
  [+] File2 has deduplication logic

Recommendations:
  1. Add error handling to File2
  2. Align logging approach
  3. Standardize control flow
  4. Sync configuration values
```

## Error Handling

- File not found → verify paths
- Invalid glob → check pattern syntax
- Parse errors → ensure valid source language
- Validation failures → review changes, fix lint/test errors

## Gotchas

- Aligning two implementations doesn't mean making them identical — preserve intentional differences that reflect genuinely different inputs
- "Recommendations" without a designated reference implementation produces unprioritized noise — pick a reference first
- An alignment >90% usually means the two implementations should be one — propose extraction in `plan-create`
- Different control flow (if/else vs switch) is often stylistic; flag but don't force a change unless the project has a stated convention
