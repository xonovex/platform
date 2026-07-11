---
name: fp-guide
description: "Use when writing functional-style code or reviewing for FP cleanliness. Triggers on prompts about pure functions, immutability, function composition, module-level functions, explicit context passing, avoiding inheritance, or stateless designs, even when the user doesn't say 'FP' or 'functional'."
---

# General Functional Programming Guidelines

## Core principles

- Prefer module-level functions over classes; pass context explicitly; avoid inheritance.
- Favor immutable data and pure functions; no global or shared mutable state.
- Compose complex behavior from small, focused functions.
- Order pattern matching from most specific to most general.

## Gotchas

- Most language `map`/`filter` are eager, not lazy — chaining many of them allocates intermediate arrays; use generators / streams for large data
- Closures over mutable state look pure but aren't — referential transparency requires both inputs and captured environment to be immutable
- Partial application order matters — `flip` exists because curry direction is opinionated, not arbitrary
- `Promise`/IO/effects compose differently from pure values — naïve `map` over them runs the effect, not the value
