# Dependency Inversion and the Dependency Rule

Keep the core abstract and stable: every source-code dependency points inward toward a port the core owns, never outward at a concrete adapter.

## The dependency rule

- **Point dependencies inward** — inner code (core, policy, domain) names only inner things and ports; outer code (adapters, frameworks, I/O) names inner code. The arrow crosses the boundary one way only.
- **An outer symbol must never appear in inner code** — no adapter type, driver struct, client SDK, or transport name in the core. If it leaked in, the boundary is fictional.
- **Grep to detect a leak** — the rule is mechanically checkable. Search the core tree for any adapter symbol; a hit is a violation, not a style nit.

```sh
# BAD: core references a concrete adapter -> dependency points outward
$ grep -rn 'PostgresRepository\|sql.Open\|postgres://' core/
core/order_service.go:42: r := repo.NewPostgresRepository()   # leak

# GOOD: core mentions only ports -> zero hits
$ grep -rn 'PostgresRepository\|sql.Open' core/   # (no output)
```

See [ports-and-adapters.md](ports-and-adapters.md) for how those boundary interfaces are shaped and sized.

## Dependency inversion (DIP)

Both sides depend on the abstraction. The high-level core does not depend on the low-level adapter; instead both depend on a port, and the adapter depends on the core — never the reverse.

- **The core owns the port** — declare the interface beside the code that consumes it, in the inner package. The port expresses what the core needs, in the core's vocabulary.
- **The adapter depends on the port** — it imports the core to implement the interface. The import direction is adapter -> core, satisfying the dependency rule.
- **Invert by ownership, not by file moves** — moving an interface into a shared "interfaces" package everyone imports does not invert anything; the consumer must own it.

```go
// BAD: core imports a concrete adapter; arrow points OUT.
package core
import "platform/repo/postgres"             // <- outward dependency

func PlaceOrder(o Order) error {
    r := postgres.New()                       // core welded to one repository
    return r.Save(o)
}
```

```go
// GOOD: core declares the port it needs; adapter depends on core.
package core

type Repository interface {                    // port, owned by the core
    Save(o Order) (OrderID, error)
}

func PlaceOrder(r Repository, o Order) error { // injected, not constructed
    _, err := r.Save(o)
    return err
}
```

```go
// adapter package -> imports core, implements the port
package postgres
import "platform/core"

type Repository struct{ /* dsn, db handle */ }
func (p Repository) Save(o core.Order) (core.OrderID, error) { /* ... */ }
```

The wiring lives in a **composition root** (a builder that picks one repository adapter — Postgres, S3, or in-memory — and injects it). The core asks for a `Repository`; it never learns which concrete one it got. Prefer this over a service locator, which hides the dependency and re-couples the core to a global lookup. See **microkernel-pattern-guide** for the plug-in registry view of the same composition root.

## Instability and the Zone of Pain

- **Most-depended-on means most-abstract** — a node many modules point at should be a port (abstract, stable). A concrete module that everything depends on is stable _and_ rigid: hard to change, expensive to break — the zone of pain.
- **Keep the center abstract** — let stability accrue to interfaces, not implementations. Adapters stay concrete but unstable (few or no inbound dependencies), so they are cheap to swap or delete.
- **Bind late** — choosing the concrete adapter at the composition root pushes binding time outward, so the abstract center never hardens around one implementation.

```text
BAD:  many modules ──> PostgresRepository (concrete + heavily depended-on = zone of pain)
GOOD: many modules ──> Repository (port) <── PostgresRepository (concrete + swappable)
```

Concrete and depended-upon is the trap; abstract and depended-upon is the goal. The connascence between core and adapter should be the weakest possible — name only, mediated by the port — never connascence of a concrete type or construction order (see **connascence-guide**).

Back to [SKILL.md](../SKILL.md).
