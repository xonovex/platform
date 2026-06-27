# Testability Is the Payoff

A clean port boundary is not an aesthetic goal; it is the thing that lets you unit-test the core with no real I/O.

- **Treat testability as proof, not decoration** — if the core obeys the dependency rule, you can exercise it in a test by substituting test doubles for every driven adapter and driving it through a primary adapter. If you cannot, the boundary is not where you think it is.
- **Driving the core in a test is itself a primary adapter** — the test plays the role the HTTP handler or CLI plays in production: it constructs the core, calls its inbound port, and asserts on the result. No new abstraction is needed to make the core testable; the same ports do it.
- **Driven ports get fakes** — replace the real repository, clock, or network with an in-memory implementation of the same outbound port. The core cannot tell the difference, because it only knows the port.
- **A test that needs real infrastructure is a design signal** — if a unit test forces you to stand up a database, open a socket, or wire the whole object graph, a boundary leaked. Fix the boundary, not the test.

## The leak shows up as slow, fragile tests

```text
BAD  — the "unit" test needs a real database
func TestPlaceOrder(t *testing.T) {
    db := postgres.Connect(os.Getenv("DATABASE_URL")) // real I/O
    svc := order.NewService(db)                        // core reaches a concrete driver
    if err := svc.Place(ctx, order); err != nil {      // can't run without a live DB
        t.Fatal(err)
    }
}
```

The core depends on a concrete `*postgres.DB`, so the test inherits that dependency. The fix is not a test container; it is dependency inversion — the core should depend on an outbound port it owns.

```text
GOOD — the core is driven against an in-memory adapter
type Repository interface { Save(context.Context, Order) error } // port owned by the core

func TestPlaceOrder(t *testing.T) {
    repo := &InMemoryRepository{}       // in-memory driven adapter
    svc := order.NewService(repo)       // inject the fake through the port
    if err := svc.Place(ctx, order); err != nil { // the test is the primary adapter
        t.Fatal(err)
    }
    if len(repo.saved) != 1 {
        t.Fatalf("want 1 saved order, got %d", len(repo.saved))
    }
}

type InMemoryRepository struct{ saved []Order }
func (r *InMemoryRepository) Save(_ context.Context, o Order) error {
    r.saved = append(r.saved, o); return nil
}
```

Same core, same port — only the adapter changed. The test is fast, deterministic, and has no I/O.

## Narrow the port until the fake is trivial

- **A hard-to-fake port is a too-wide port** — if your fake must reimplement query languages or transaction semantics, the core is asking the outside world for too much. Narrow the port to the few operations the core actually needs (`Save`, `ByID`), and the fake collapses to a map.
- **Push impure work outward** — when a test wants to control time or randomness, inject a `Clock` or `IDSource` port rather than calling `time.Now()` inside the core. The boundary that makes the core pure is the same boundary that makes it testable. See [ports-and-adapters.md](ports-and-adapters.md).
- **Keep the composition root out of the unit test** — production wiring lives in one place and uses real adapters; the unit test wires fakes directly. The test should never invoke the application's full assembly. See [composition-root.md](composition-root.md).

## One illustration: an order service

An order service keeps persistence behind its own port — a `Repository`. The core builds a placement from a neutral `Order` value the adapter returns and enforces a fail-closed validation gate before saving.

```text
type Repository interface { ByID(context.Context, OrderID) (Order, error) }

func TestPlaceRejectsInvalidOrder(t *testing.T) {
    repo := stubRepository{order: Order{Valid: false}}
    svc := order.NewService(repo)                 // test drives the core
    verdict, err := svc.Place(ctx, id)
    if err != nil { t.Fatal(err) }
    if verdict != Rejected {                      // assert the gate fired
        t.Fatalf("fail-closed gate let an invalid order through: %v", verdict)
    }
}
```

The fail-closed gate is testable precisely because the order is a plain value crossing a port, not a side effect buried in a real repository. To exercise the gate you supply a stub that self-declares an invalid order — no database, no live socket.

## When a test still needs the real thing

- **That is an integration test, and it is allowed** — adapters must be verified against real infrastructure somewhere. Keep those tests separate, fewer, and on the slow path; do not let their cost contaminate the core's unit tests.
- **The boundary tells you which is which** — a test that touches only the core and its ports is a unit test; a test that touches a concrete adapter's external dependency is an integration test. If your "unit" tests keep drifting into the second kind, the core is leaking.

For the precise coupling and connascence vocabulary behind "the boundary leaked" — why a real-DB dependency in a unit test is connascence of execution and position across a boundary that should carry none — see **connascence-guide**.

Back to [SKILL.md](../SKILL.md).
