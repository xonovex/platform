# Ports and Adapters

A port is an interface declared by the core in its own domain terms; everything outside — DBs, buses, HTTP, CLIs, files, clocks, devices, test doubles — is an adapter, and none is privileged. Adapters depend on the core; the core imports no adapter. The hexagon is not a layer cake: six sides carry no meaning, they only signal "many ports around a center", not a stack with the database at the bottom.

## Driving vs driven

- **Driving (primary)** adapters call INTO the core through an inbound port — CLI, HTTP handler, scheduler, test. They translate an external request into a core call.
- **Driven (secondary)** adapters are called OUT BY the core through an outbound port — repository, publisher, clock, device. The core states what it needs; the adapter satisfies it.
- The core owns **both** port shapes, phrased in domain language. Adapters translate to wire protocols, SQL, or syscalls.

```go
// Core owns both port shapes; it never imports an adapter.
type PlaceOrder interface { // driving port: the world calls in
    Place(ctx context.Context, o Order) (OrderID, error)
}
type OrderRepository interface { // driven port: the core calls out
    Save(ctx context.Context, o Order) error
}
```

- Grade each boundary with **connascence-guide**: a port should carry only name/type, never positional or algorithmic assumptions about its implementation.

## Relationship to other patterns

- This skill owns the substrate: who declares the interface, who implements it, which way dependencies point (see [dependency-inversion.md](dependency-inversion.md)).
- Once adapters are discovered, registered, and loaded by extensibility machinery rather than hand-wired, you are on the plug-in/registry layer — see **microkernel-pattern-guide**.

Back to [SKILL.md](../SKILL.md).
