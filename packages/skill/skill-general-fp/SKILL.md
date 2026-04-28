---
name: general-fp-guidelines
description: "Functional programming guidelines: pure functions, immutability, composition over inheritance. Apply when writing FP-style code, module-level functions, or explicit context passing. Use for stateless designs, composition, function chaining. Keywords: functional programming, pure functions, immutability, composition, module functions, explicit context, stateless, FP."
---

# General Functional Programming Guidelines

## Core principles

- Modular design: Prefer module-level functions; pass explicit context; avoid classes.
- Immutability: Favor immutable data; use pure functions without side effects.
- Composition: Build complex behavior by composing simple functions.
- Type safety: Use types consistently; derive from generated/parent types; enable strict modes.
- Clear structure: Split logic into small, focused files/functions.

## Best practices

- Readability: Prefer clarity over cleverness; name things well.
- Errors: Handle and propagate explicitly; never swallow silently.
- Tests: Add/maintain tests to prevent regressions.
- State: Pass state explicitly; avoid global or shared mutable state.

## Code quality

- Linting: Fix root causes of warnings; never suppress with disable comments.
- Validation: Run typecheck, lint, build, test after each major change (all must pass).
- Organization: Order pattern matching from most specific to most general.
