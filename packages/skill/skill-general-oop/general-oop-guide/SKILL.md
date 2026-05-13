---
name: general-oop-guide
description: "Use when designing class hierarchies or applying OOP principles. Triggers on prompts about classes, inheritance, polymorphism, encapsulation, SOLID, design patterns, or interfaces, even when the user doesn't say 'OOP'. Skip functional-style designs (use general-fp-guide) and language-specific guidance covered by typescript-guide / python-guide."
---

# General Object-Oriented Programming Guidelines

## Core principles

- Encapsulation: Keep state private; expose behavior through methods.
- Inheritance: Use for shared behavior; prefer composition when appropriate.
- Polymorphism: Program to interfaces; use abstract classes for shared implementation.
- Type safety: Use types consistently; derive from generated/parent types; enable strict modes.
- Clear structure: Split logic into focused classes with single responsibilities.

## Best practices

- Readability: Prefer clarity over cleverness; name things well.
- Errors: Handle and propagate explicitly; never swallow silently.
- Tests: Add/maintain tests to prevent regressions.
- SOLID: Single responsibility, open/closed, Liskov substitution, interface segregation, dependency inversion.

## Code quality

- Linting: Fix root causes of warnings; never suppress with disable comments.
- Validation: Run typecheck, lint, build, test after each major change (all must pass).
- Design patterns: Apply appropriate patterns (Factory, Strategy, Observer) when they simplify code.

## Gotchas

- Liskov substitution is about behavior, not just signatures — `Square extends Rectangle` is the classic violation despite matching types
- Equality (`equals` / `==`) and hashing must agree — overriding one and not the other breaks `HashSet`/`HashMap` silently
- Premature class hierarchies are harder to refactor than premature composition — start flat, extract a base only when ≥3 implementations exist
- Mutable shared state plus method calls = order-dependent behavior; making fields `final`/`readonly` is the cheapest correctness lever
