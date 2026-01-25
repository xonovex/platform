---
description: >-
  Simplify code by consolidating duplicates, removing dead code, and flattening
  abstractions
model: sonnet
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
argument-hint: >-
  [path] [--aspects
  <duplicates,dead-code,dependencies,abstractions,interfaces,config>]
  [--dry-run] [--fix]
---

# /code-simplify – Research Code Simplification Opportunities

Analyzes code complexity to identify consolidation, dead code removal, and simplification opportunities. Generates a detailed research report. Does NOT create plans or make changes - run `/plan-create` afterward to create an implementation plan.

## Goal

- Consolidate duplicated logic into shared utilities
- Remove unused code, imports, and dependencies
- Flatten over-engineered abstractions
- Simplify complex interfaces
- Centralize scattered configuration

## Usage

```bash
/code-simplify packages/my-package/
/code-simplify src/ --aspects duplicates,dead-code
/code-simplify . --dry-run
```

## Arguments

- `path` (required): Directory to analyze
- `--aspects` (optional): Aspects to check (comma-separated):
  - `duplicates` - Identical/near-identical functions
  - `utilities` - Logic for shared utilities
  - `types` - Redundant type definitions
  - `constants` - Scattered magic values
  - `patterns` - Repeated code patterns
  - `dead-code` - Unused functions, unreachable branches
  - `dependencies` - Unused imports, redundant packages
  - `abstractions` - Over-engineered patterns
  - `interfaces` - Complex APIs (>5 params, nested config)
  - `config` - Scattered configuration
  - `all` (default) - All aspects
- `--dry-run` (optional): Report without changes

## Core Workflow

**Use Task agents with subagent_type=Explore and model=haiku for codebase analysis. Do NOT use EnterPlanMode.**

1. **Discovery** - Use Task/Explore agents to scan files, build signature index, analyze dependency trees
2. **Analysis** - Check each aspect for issues using Explore agents
3. **Report** - Group by priority, estimate impact, generate detailed report

## Analysis Aspects

**Duplicates**: Same function in multiple files, identical logic with different names, copy-paste blocks (>10 lines)

**Utilities**: Helpers used in multiple places, repeated validation/transformation logic

**Types**: Same interface in multiple files, similar types to unify

**Constants**: Repeated hardcoded values, duplicated config, copied regex patterns

**Patterns**: Duplicated error handling, logging, response formatting

**Dead Code**: Uncalled functions, unreachable branches, unused exports/variables

**Dependencies**: Unused imports, packages not referenced, transitive deps, duplicate functionality

**Abstractions**: Single-implementation interfaces, wrapper classes, single-type factories, deep inheritance (>3 levels)

**Interfaces**: Functions with >5 params, deeply nested config, rarely-used optionals, inconsistent signatures

**Config**: Scattered env access, duplicate defaults, duplicated feature flags/URLs

## Output Format

```
=== Code Simplification Report ===
Path: packages/example/

DUPLICATES (8)
  HIGH: makeNullable() - 6 implementations -> utils/type-modifiers.ts
  HIGH: inferType() - 4 implementations -> utils/inference.ts

DEAD CODE (12)
  HIGH: legacyParser() never called - src/parsers/legacy.ts:45
  MEDIUM: unreachable else - src/handlers/process.ts:89

UNUSED DEPENDENCIES (5)
  HIGH: lodash -> use optional chaining
  MEDIUM: moment.js -> already have date-fns

OVER-ENGINEERING (3)
  HIGH: IRepository (1 impl) -> inline implementation
  MEDIUM: ConfigFactory (1 type) -> use constructor

Impact: ~450 lines removed, 3 dependencies pruned, 25% complexity reduction
```

## Implementation Details

**Finding issues**: Grep for signatures, Task agent with subagent_type=Explore and model=haiku for semantic analysis, trace exports, check call graphs

**Location strategy**:

- Same package: `utils/` subdirectory
- Cross-package: Existing `shared-*` or propose new
- Follow project conventions

**Safe changes**: Identical implementations, unused imports, single-impl interfaces, constant consolidation

**Validation**: Fix incrementally, validate after each (typecheck/lint/test)

## Error Handling

- **False positives**: Check dynamic imports, reflection, external entry points
- **Breaking changes**: Ensure no external consumers
- **Circular deps**: Restructure or create intermediate module

## Examples

```bash
/code-simplify packages/api/ --aspects dead-code --dry-run
/code-simplify services/ --aspects abstractions,interfaces
```

## Next Steps

After running this research command:

1. Review the simplification report for accuracy
2. Run `/plan-create` to create an implementation plan from this research
