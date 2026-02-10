---
name: general-oop-guidelines
description: Trigger on object-oriented patterns, class hierarchies, encapsulation. Use for OOP-style coding principles. Apply when using inheritance, polymorphism, encapsulated state. Keywords: object-oriented, classes, inheritance, polymorphism, encapsulation, SOLID, design patterns, interfaces.
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
