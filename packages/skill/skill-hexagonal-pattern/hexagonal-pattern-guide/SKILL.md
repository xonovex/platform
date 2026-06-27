---
name: hexagonal-pattern-guide
description: "Use when isolating an application or domain core from its I/O and delivery mechanisms behind interfaces — hexagonal / ports-and-adapters / clean / onion architecture. Triggers on decoupling the domain from the database/UI/framework, driving (primary) vs driven (secondary) adapters, dependency inversion / the dependency rule ('depend on an interface, not the implementation'), a composition root wiring adapters to ports, swapping a real adapter for a fake in tests, or keeping business logic free of infrastructure — even when the user says 'ports and adapters' or 'clean architecture'."
---

# Hexagonal Architecture (Ports and Adapters)

Isolate the domain core behind ports it owns, so every I/O and delivery mechanism is an interchangeable adapter on the outside.

## Essentials

- **Core owns the ports; the outside is adapters** - interfaces defined by the core, implemented outside, see [references/ports-and-adapters.md](references/ports-and-adapters.md)
- **Driving in, driven out** - primary adapters call into the core; secondary adapters are called out by it, see [references/ports-and-adapters.md](references/ports-and-adapters.md)
- **Dependencies point inward** - the core names no adapter; the dependency rule + dependency inversion, see [references/dependency-inversion.md](references/dependency-inversion.md)
- **One composition root** - the single place that wires concrete adapters to ports, see [references/composition-root.md](references/composition-root.md)
- **Testability is the payoff** - fake the driven adapters, drive the core in a test, no real I/O, see [references/testability.md](references/testability.md)

## Gotchas

- The hexagon isn't six of anything — it just signals "many ports", not a top/bottom layering.
- A port belongs to the _core_, not the adapter — if the adapter defines the interface, the dependency points the wrong way.
- Ports-and-adapters is not the microkernel pattern — hexagonal isolates the domain behind a usually-fixed set of ports; an open, _registered_ plug-in set is microkernel-pattern-guide. They share the adapter mechanism, nothing more.
- Anemic ports leak infrastructure types (ORM entities, HTTP request objects) into the core — keep port types domain-owned.
- If a unit test needs a real database or network, the boundary leaked — narrow the port or invert the dependency.

## Example

```go
// core owns the port; it never imports a driver
type Repository interface{ Save(o Order) error }

type OrderService struct{ repo Repository } // depends on the port, not Postgres

// driven adapter (outside): a PostgresRepository implementation of the port
// driving adapter (outside): an HTTP handler that calls OrderService
// test: an InMemoryRepository fake — no database
```

## Progressive Disclosure

- Read [references/ports-and-adapters.md](references/ports-and-adapters.md) - Load when defining ports and driving/driven adapters around a core
- Read [references/dependency-inversion.md](references/dependency-inversion.md) - Load when deciding which way dependencies point or detecting a core leak
- Read [references/composition-root.md](references/composition-root.md) - Load when wiring concrete adapters to ports in one place
- Read [references/testability.md](references/testability.md) - Load when testing a core in isolation with fakes/test doubles
