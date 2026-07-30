# Testability Is the Payoff

If the core obeys the dependency rule, you can unit-test it with no real I/O: substitute test doubles for every driven adapter and drive it through a primary adapter. If you cannot, the boundary is not where you think it is.

- **Driving the core in a test is itself a primary adapter**: the test plays the role the HTTP handler or CLI plays in production: construct the core, call its inbound port, assert on the result. No new abstraction is needed.
- **Driven ports get fakes**: replace the real repository, clock, or network with an in-memory implementation of the same outbound port; the core cannot tell the difference.
- **A unit test that needs real infrastructure is a design signal**: if it forces you to stand up a database or open a socket, a boundary leaked. Fix the boundary (invert the dependency), not the test.

```go
type Repository interface { Save(context.Context, Order) error } // port owned by the core

func TestPlaceOrder(t *testing.T) {
    repo := &InMemoryRepository{}                  // in-memory driven adapter
    svc := order.NewService(repo)                  // inject the fake through the port
    if err := svc.Place(ctx, order); err != nil { t.Fatal(err) } // test is the primary adapter
    if len(repo.saved) != 1 { t.Fatalf("want 1 saved, got %d", len(repo.saved)) }
}

type InMemoryRepository struct{ saved []Order }
func (r *InMemoryRepository) Save(_ context.Context, o Order) error {
    r.saved = append(r.saved, o); return nil
}
```

- **A hard-to-fake port is a too-wide port**: if the fake must reimplement query languages or transaction semantics, narrow the port to the few operations the core needs (`Save`, `ByID`) and the fake collapses to a map.
- **Push impure work outward**: inject a `Clock` or `IDSource` port rather than calling `time.Now()` inside the core; the boundary that makes the core pure is the one that makes it testable.
- **Keep the composition root out of the unit test**: production wiring lives in one place with real adapters; the unit test wires fakes directly. See [composition-root.md](composition-root.md).
- **Real infrastructure means it's an integration test**: allowed, but keep those separate, fewer, and on the slow path. A test touching only the core and its ports is a unit test; one touching a concrete adapter's external dependency is not. For the coupling vocabulary behind "the boundary leaked", see **connascence-guide**.

Back to [SKILL.md](../SKILL.md).
