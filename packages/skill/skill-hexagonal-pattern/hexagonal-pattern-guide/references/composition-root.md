# The Composition Root

The single place that builds concrete adapters and binds them to the core's ports — the only code allowed to name concretes.

## Principle

- **One root, one job** — the composition root constructs every concrete adapter and wires it to a port. Nothing else in the system imports a concrete adapter or knows which one is live.
- **Everything else depends on ports** — the core, and every other adapter, speaks only to interfaces. Concretes are an implementation detail confined to the root. This is the dependency rule and dependency inversion made physical: the wiring sits at the outermost layer, pointing inward. See [ports-and-adapters.md](ports-and-adapters.md).
- **Near the entry point** — put the root next to `main` (or the request/job entry). It runs once at startup, reads config, and assembles the object graph before any business logic executes. Binding time is explicit and early.
- **Visible, ordered, testable** — wiring is plain top-to-bottom code: construct dependencies, then the things that need them, then hand the graph to the core. A reviewer sees the whole graph in one file; a test can build a different graph the same way.

## Inject through constructors

Pass adapters in. Never let the core reach out and fetch them.

- **Constructor injection** — a component receives its collaborators as arguments and stores ports, not concretes. Its dependencies are stated in its signature; you cannot construct it without satisfying them.
- **No ambient lookup** — the core does not call a global, a singleton, or a package-level `Get…()` to obtain a collaborator. If it can reach out, the dependency is hidden and the root is no longer the single wiring point.

```go
// BAD — core reaches out; dependency is invisible and global
func PlaceOrder(o Order) error {
    repo := registry.Lookup("repository") // hidden, untyped, global
    return repo.Save(o)
}
```

```go
// GOOD — dependency named in the signature, supplied by the root
type OrderService struct {
    repo  Repository // port
}

func NewOrderService(r Repository) *OrderService {
    return &OrderService{repo: r}
}

func (s *OrderService) PlaceOrder(o Order) error {
    if err := o.Validate(); err != nil {
        return err
    }
    return s.repo.Save(o)
}
```

## Not a service locator

- **Service locator hides dependencies** — passing a registry into business logic and having it pull what it needs inverts the win: the signature says `OrderService(reg Registry)` but the real dependency (`Repository`) is discovered at runtime, untyped and unenforced. The compiler can no longer tell you a wire is missing.
- **Keep lookup at the edge** — a registry is fine _inside_ the composition root, where it resolves names to concrete adapters and the result is injected as a typed port. The locator must not cross into the core. For an open plug-in registry, self-declared guarantees, and the wiring trade-offs of late binding and lazy activation, see **microkernel-pattern-guide**.

## Wiring a graph

Select concretes once, on one axis at a time, then assemble:

```go
func main() {
    cfg := LoadConfig()

    // root names concretes; chooses one per axis
    var repo Repository
    switch cfg.Store {
    case "postgres":
        repo = postgres.New(cfg.Postgres)
    case "s3":
        repo = s3repo.New(cfg.S3)
    }

    // cross-cutting wrapper composed around the chosen port
    repo = audit.Wrap(repo, cfg.Audit)

    service := NewOrderService(repo)
    if err := http.Serve(cfg.Addr, service); err != nil {
        log.Fatal(err)
    }
}
```

- **One switch per axis** — choosing the store independently from other concerns keeps each axis open for extension and closed for modification: adding an adapter touches one branch in the root and nothing in the core.
- **Decorate at the root** — cross-cutting wrappers (auditing, logging, retries) compose around a port here, where the graph is assembled, so the core stays unaware of them. Complete mediation stays in one provable place.
- **Test the same way** — a test builds its own graph with fakes and exercises the core through ports, no global teardown required. See [testability.md](testability.md).

Back to [SKILL.md](../SKILL.md).
