# Dependency Inversion and the Dependency Rule

Every source-code dependency points inward toward a port the core owns, never outward at a concrete adapter. Inner code (core, policy, domain) names only inner things and ports; outer code (adapters, frameworks, I/O) names inner code. No adapter type, driver struct, client SDK, or transport name may appear in the core — if it leaked in, the boundary is fictional.

The rule is mechanically checkable — grep the core tree for any adapter symbol; a hit is a violation:

```sh
grep -rn 'PostgresRepository\|sql.Open\|postgres://' core/   # any hit = leak
```

## Invert by ownership, not by file moves

Both sides depend on the abstraction: the core declares the port it needs (in its own vocabulary, beside the consuming code), and the adapter imports the core to implement it. Moving an interface into a shared `interfaces` package everyone imports inverts nothing — the **consumer** must own it.

```go
// core: declares the port it needs; never constructs a concrete.
package core
type Repository interface { Save(o Order) (OrderID, error) }
func PlaceOrder(r Repository, o Order) error { _, err := r.Save(o); return err }

// adapter: imports core, implements the port (arrow points adapter -> core).
package postgres
import "platform/core"
type Repository struct{ /* db handle */ }
func (p Repository) Save(o core.Order) (core.OrderID, error) { /* ... */ }
```

Wiring lives in a [composition root](composition-root.md), never a service locator (which hides the dependency and re-couples the core to a global lookup).

## Instability and the Zone of Pain

A module many others depend on should be abstract (a port) — concrete **and** heavily depended-on is the "zone of pain": stable yet rigid, expensive to change. Keep stability on interfaces; let adapters stay concrete but unstable (few inbound deps) so they are cheap to swap or delete. Binding late at the composition root keeps the abstract center from hardening around one implementation.

Back to [SKILL.md](../SKILL.md).
