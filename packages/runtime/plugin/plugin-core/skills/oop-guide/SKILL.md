---
name: oop-guide
description: "Use when designing class hierarchies or applying OOP principles. Triggers on prompts about classes, inheritance, polymorphism, encapsulation, SOLID, design patterns, or interfaces, even when the user doesn't say 'OOP'."
---

# General Object-Oriented Programming Guidelines

## Stances

- Prefer composition over inheritance; program to interfaces, reserving abstract classes for shared implementation.
- Keep state private; expose behavior through methods.
- SOLID: single responsibility, open/closed, Liskov substitution, interface segregation, dependency inversion.
- Apply patterns (Factory, Strategy, Observer) only when they simplify code, never speculatively.

## Design smells

Recognizers for common OO-design problems and the principle each violates.

- **God Object / Large Class**: one class with many responsibilities and low cohesion; split by responsibility (single responsibility).
- **Refused Bequest**: a subclass that ignores or throws on inherited behavior breaks the base-class contract (Liskov substitution); prefer composition over that inheritance.
- **Downcasting**: a cast that asks "which subtype is this really?" breaks the abstraction and usually signals a Liskov violation; dispatch through the interface instead.
- **Divergent Change**: one class edited for many unrelated reasons; separate the reasons into focused classes (single responsibility).
- **Parallel Inheritance Hierarchies**: every new subclass in one tree forces a mirrored subclass in another; collapse or fold the hierarchies.
- **Alternative Classes with Different Interfaces**: two classes do the same job with unswappable signatures; unify the interface so callers don't choose sides.
- **Temporary Field**: a field populated only in some circumstances, empty otherwise; extract the conditional behavior (often into a method object) so the field isn't left dangling.

## Gotchas

- Liskov substitution is about behavior, not just signatures: `Square extends Rectangle` is the classic violation despite matching types
- Mutable shared state plus method calls = order-dependent behavior; making fields `final`/`readonly` is the cheapest correctness lever
