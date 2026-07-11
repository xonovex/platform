# code-shared-extract: Identify Common Patterns for Shared Library Extraction

Find duplicated patterns across packages and propose extraction to a shared library. Read-only: produces a ranked candidate report and phased extraction plan.

## Technique

- Scan for repeated functions, components, hooks, middleware, types, constants (semantic match, not just identical text)
- Group by similarity: identical code · equivalent logic · same interface
- Rank candidates by `occurrences × complexity × cross-package count`
- Choose target: same-package `utils/` · cross-package `shared-*` · new library
- Plan phased extraction minimizing breaking changes; validate with typecheck / lint / tests
