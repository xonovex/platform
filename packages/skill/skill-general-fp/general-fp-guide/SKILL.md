---
name: general-fp-guide
description: "Use when writing functional-style code or reviewing for FP cleanliness. Triggers on prompts about pure functions, immutability, function composition, module-level functions, explicit context passing, avoiding inheritance, or stateless designs, even when the user doesn't say 'FP' or 'functional'."
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

## Gotchas

- Most language `map`/`filter` are eager, not lazy — chaining many of them allocates intermediate arrays; use generators / streams for large data
- Closures over mutable state look pure but aren't — referential transparency requires both inputs and captured environment to be immutable
- Partial application order matters — `flip` exists because curry direction is opinionated, not arbitrary
- `Promise`/IO/effects compose differently from pure values — naïve `map` over them runs the effect, not the value
