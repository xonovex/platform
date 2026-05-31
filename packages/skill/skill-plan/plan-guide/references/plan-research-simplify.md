# plan-research-simplify: Research Code Simplification Opportunities

Analyze code complexity to identify consolidation, dead code removal, and simplification opportunities. Generates a research report. Does **not** create plans or make changes — run `plan-create` afterward.

## Goal

- Consolidate duplicated logic into shared utilities
- Remove unused code, imports, dependencies
- Flatten over-engineered abstractions
- Simplify complex interfaces
- Centralize scattered configuration

## Core Workflow

**Delegate codebase analysis to read-only search agents where available; otherwise use grep/find/file-read directly. Stay in research mode.**

1. **Discovery** — scan files, build a signature index, analyze dependency trees
2. **Analysis** — check each aspect for issues via focused read-only searches
3. **Report** — group by priority, estimate impact, generate a detailed report

## Analysis Aspects

| Aspect           | What to look for                                                                                       |
| ---------------- | ------------------------------------------------------------------------------------------------------ |
| **Duplicates**   | Same function in multiple files, identical logic with different names, copy-paste blocks (>10 lines)   |
| **Utilities**    | Helpers used in multiple places, repeated validation / transformation logic                            |
| **Types**        | Same interface in multiple files, similar types to unify                                               |
| **Constants**    | Repeated hardcoded values, duplicated config, copied regex patterns                                    |
| **Patterns**     | Duplicated error handling, logging, response formatting                                                |
| **Dead code**    | Uncalled functions, unreachable branches, unused exports / variables                                   |
| **Dependencies** | Unused imports, packages not referenced, transitive deps, duplicate functionality                      |
| **Abstractions** | Single-implementation interfaces, wrapper classes, single-type factories, deep inheritance (>3 levels) |
| **Interfaces**   | Functions with >5 params, deeply nested config, rarely-used optionals, inconsistent signatures         |
| **Config**       | Scattered env access, duplicate defaults, duplicated feature flags / URLs                              |

## Output Format

```
=== Code Simplification Report ===
Path: packages/example/

DUPLICATES (8)
  HIGH: makeNullable() - 6 implementations → utils/type-modifiers.ts
  HIGH: inferType() - 4 implementations → utils/inference.ts

DEAD CODE (12)
  HIGH: legacyParser() never called - src/parsers/legacy.ts:45
  MEDIUM: unreachable else - src/handlers/process.ts:89

UNUSED DEPENDENCIES (5)
  HIGH: lodash → use optional chaining
  MEDIUM: moment.js → already have date-fns

OVER-ENGINEERING (3)
  HIGH: IRepository (1 impl) → inline implementation
  MEDIUM: ConfigFactory (1 type) → use constructor

Impact: ~450 lines removed, 3 dependencies pruned, 25% complexity reduction
```

## Implementation Details

**Finding issues:** grep for signatures, delegate semantic analysis to read-only search agents where available, trace exports, check call graphs

**Location strategy:**

- Same package: `utils/` subdirectory
- Cross-package: existing `shared-*` or propose new
- Follow project conventions

**Safe to recommend:** identical implementations, unused imports, single-impl interfaces, constant consolidation

**Sequencing:** group findings so the downstream plan can apply them incrementally, validating after each (typecheck / lint / test)

## Error Handling

- **False positives:** check dynamic imports, reflection, external entry points
- **Breaking changes:** ensure no external consumers
- **Circular deps:** restructure or create an intermediate module

## Gotchas

- "Dead code" detection misses code reachable via dynamic dispatch, reflection, or external entry points — verify before deleting
- Single-implementation interfaces aren't always over-engineering — they may exist for testability or planned variants
- Consolidating into a "shared utility" without a clear owner produces a junk drawer — propose a clear home in the report
- Removing a dependency that's used transitively for typing only is harmless but looks alarming in diffs — call this out explicitly
