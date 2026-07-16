# The Composition Root

One place — next to `main` (or the request/job entry) — constructs every concrete adapter and binds it to a port. It is the **only** code allowed to name concretes; the core and every other adapter speak only to interfaces. It runs once at startup, reads config, and assembles the object graph before any business logic executes.

## Inject through constructors, not ambient lookup

A service locator (passing a registry into business logic that pulls what it needs) defeats constructor injection: the signature says `OrderService(reg Registry)` but the real dependency is discovered at runtime, untyped, so the compiler can no longer catch a missing wire. Keep the registry **inside** the root, where it resolves names to concretes and injects the result as a typed port; it must not cross into the core. For an open plug-in registry, self-declared guarantees, and lazy activation, see **microkernel-pattern-guide**.

## Wiring a graph — one switch per axis

```go
func main() {
    cfg := LoadConfig()

    var repo Repository // choose one concrete per axis
    switch cfg.Store {
    case "postgres": repo = postgres.New(cfg.Postgres)
    case "s3":       repo = s3repo.New(cfg.S3)
    }

    repo = audit.Wrap(repo, cfg.Audit) // cross-cutting wrapper composed around the port

    service := NewOrderService(repo)
    if err := http.Serve(cfg.Addr, service); err != nil { log.Fatal(err) }
}
```

Choosing each concern independently keeps every axis open for extension: adding an adapter touches one branch in the root and nothing in the core. Cross-cutting wrappers (audit, logging, retries) decorate a port here, where the graph is assembled, so the core stays unaware of them. A unit test builds its own graph with fakes the same way — see [testability.md](testability.md).

Back to [SKILL.md](../SKILL.md).
