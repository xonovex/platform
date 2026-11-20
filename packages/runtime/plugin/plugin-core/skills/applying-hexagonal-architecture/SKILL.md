---
name: applying-hexagonal-architecture
description: "Use when separating business logic from the database, UI, framework or other I/O behind interfaces, or structuring an app as hexagonal, ports-and-adapters, clean or onion architecture. Triggers on ports, adapters, driving vs driven adapters, dependency inversion, the dependency rule, a composition root, swapping a real adapter for a fake in tests, or business logic tangled with infrastructure, even when the user doesn't say 'hexagonal' and says 'ports and adapters' or 'clean architecture'."
---

# Hexagonal Architecture (Ports and Adapters)

Isolate the domain core behind ports it owns, so every I/O and delivery mechanism is an interchangeable adapter on the outside.

## Essentials

- **Driving in, driven out** - primary adapters call into the core; secondary adapters are called out by it, see [references/ports-and-adapters.md](references/ports-and-adapters.md)
- **Dependencies point inward** - the core names no adapter; the dependency rule + dependency inversion, see [references/dependency-inversion.md](references/dependency-inversion.md)
- **Wiring lives in one composition root** - see [references/composition-root.md](references/composition-root.md)
- **Testability is the payoff** - fake the driven adapters, drive the core in a test, no real I/O, see [references/testability.md](references/testability.md)

## Gotchas

- The hexagon isn't six of anything. It just signals "many ports", not a top/bottom layering.
- Ports-and-adapters is not the microkernel pattern: hexagonal isolates the domain behind a usually-fixed set of ports; an open, _registered_ plug-in set is **applying-microkernel-architecture**. They share the adapter mechanism, nothing more.
- If a unit test needs a real database or network, the boundary leaked: narrow the port or invert the dependency.

## Example

```go
// core owns the port; it never imports a driver
type Repository interface{ Save(o Order) error }

type OrderService struct{ repo Repository } // depends on the port, not Postgres

// driven adapter (outside): a PostgresRepository implementation of the port
// driving adapter (outside): an HTTP handler that calls OrderService
// test: an InMemoryRepository fake, no database
```

## Progressive Disclosure

- Read [references/ports-and-adapters.md](references/ports-and-adapters.md) - Load when defining ports and driving/driven adapters around a core
- Read [references/dependency-inversion.md](references/dependency-inversion.md) - Load when deciding which way dependencies point or detecting a core leak
- Read [references/composition-root.md](references/composition-root.md) - Load when wiring concrete adapters to ports in one place
- Read [references/testability.md](references/testability.md) - Load when testing a core in isolation with fakes/test doubles
