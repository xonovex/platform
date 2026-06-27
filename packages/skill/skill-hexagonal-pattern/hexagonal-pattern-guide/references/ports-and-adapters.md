# Ports and Adapters

Isolate an application core behind interfaces it owns, so all I/O and delivery live outside it.

## The shape

- **Core owns the ports** — A port is an interface declared by the application core, in core terms. Adapters implement or call those ports; the core never imports an adapter.
- **The hexagon is not a layer cake** — Six sides carry no meaning. The shape only signals "many ports around a center", not a top-to-bottom stack with the database at the bottom. Resist re-drawing it as layers.
- **Everything outside is an adapter** — Databases, message buses, HTTP frameworks, CLIs, files, clocks, devices, and test doubles are all adapters. None of them is privileged.
- **One direction of dependency** — Adapters depend on the core; the core depends on nothing external. This is the dependency rule applied at the boundary. See [dependency-inversion.md](dependency-inversion.md).

## Driving vs driven

- **Driving (primary) adapters call INTO the core** — A CLI, an HTTP handler, a scheduler, or a test invokes a port the core exposes for being driven. They translate an external request into a core call.
- **Driven (secondary) adapters are called OUT BY the core** — A repository, a publisher, a clock, or a device is reached through a port the core defines for its own needs. The core states what it requires; the adapter satisfies it.
- **The core defines both port shapes** — Driving and driven port interfaces both live with the core, phrased in domain language. Adapters do the translating to wire protocols, SQL, or syscalls.

```go
// Core owns both port shapes.
type PlaceOrder interface {              // driving port: the world calls in
    Place(ctx context.Context, o Order) (OrderID, error)
}
type OrderRepository interface {         // driven port: the core calls out
    Save(ctx context.Context, o Order) error
}
```

## Why it pays off

- **Swap adapters without touching the core** — Replace a Postgres repository with an in-memory one, or an HTTP front with a CLI front, and the core is untouched. Tests run against fakes through the same ports.
- **Defer and vary binding** — Which adapter satisfies a port is decided once, late, at startup — not scattered through the core. Wire it in a [composition-root.md](composition-root.md), not by reaching for a global.
- **Grade each boundary** — A port should expose only name and type (weak connascence), never positional or algorithmic assumptions about its implementation. Grade the boundary with the **connascence-guide**.

## BAD -> GOOD

```go
// BAD: the core imports the driver and speaks SQL directly.
package order

import "database/sql"

func Place(db *sql.DB, o Order) error {
    _, err := db.Exec("INSERT INTO orders(id, total) VALUES($1,$2)", o.ID, o.Total)
    return err // core now coupled to Postgres, hard to test, hard to swap
}
```

```go
// GOOD: the core declares a port; an adapter implements it elsewhere.
package order

type Repository interface {              // port, owned by the core
    Save(ctx context.Context, o Order) error
}

func Place(ctx context.Context, repo Repository, o Order) error {
    return repo.Save(ctx, o)             // core depends on the interface only
}

// package postgres (an adapter, outside the core)
type Repo struct{ db *sql.DB }
func (r Repo) Save(ctx context.Context, o Order) error {
    _, err := r.db.ExecContext(ctx, "INSERT INTO orders(id,total) VALUES($1,$2)", o.ID, o.Total)
    return err
}
```

## Illustration: pluggable axes

An `OrderService` can split orthogonal concerns into ports the core owns — a `Repository`, a `Notifier`, a `Clock`. Each is a driven port; concrete `PostgresRepository`, `S3Repository`, or `InMemoryRepository` implementations are adapters. A neutral `OrderRecord` value crosses the boundary so no adapter leaks its own row or blob type into the core, and writes stay transactional by default. The core orchestrates; the adapters self-declare their capabilities.

## Relationship to other patterns

- **This skill owns the substrate** — Ports and adapters define the boundary itself: who declares the interface, who implements it, which way dependencies point.
- **An open, registered set of adapters is a microkernel** — Once adapters are discovered, registered, and loaded by extensibility machinery rather than hand-wired, you are building on the plug-in / registry layer. See the **microkernel-pattern-guide**, which sits on top of this substrate.
- **Back to the overview** — [SKILL.md](../SKILL.md).
