---
name: microkernel-pattern-guide
description: "Use when building an extensible system: a minimal core plus interchangeable plug-ins selected through a registry — the microkernel / plug-in architecture. Triggers on plugin/extension-point/registry design, adding a backend or driver without editing the core, open/closed extension, capability negotiation and fail-closed / secure-by-default gating, binding time (compile vs load vs run-time), dependency-injection vs service-locator wiring, lazy activation, or versioning a plug-in contract — even when the user only says 'plugin architecture'."
---

# Microkernel (Plug-in) Architecture

A minimal core plus independent plug-ins, selected through a registry — extend the system by adding a plug-in, never by editing the core.

## Essentials

- **Core + plug-ins + registry** - the minimal core holds the least; features are registered plug-ins, see [references/core-plugins-registry.md](references/core-plugins-registry.md)
- **Open/closed** - add a plug-in and register it once; never edit a central switch, see [references/core-plugins-registry.md](references/core-plugins-registry.md)
- **Capabilities, fail-closed** - plug-ins self-declare guarantees; the core gates generically and securely-by-default, see [references/capabilities-fail-closed.md](references/capabilities-fail-closed.md)
- **Bind as late as cost allows** - compile vs load vs run-time selection picks the mechanism, see [references/binding-time.md](references/binding-time.md)
- **Wire at the root, don't locate** - inject the resolved plug-in; passing the registry into business logic is the service-locator anti-pattern, see [references/wiring.md](references/wiring.md)
- **Built on ports and adapters** - plug-ins are adapters behind contracts; for the port substrate see **hexagonal-pattern-guide**

## Gotchas

- A microkernel is not "just a hexagon" — it adds an _open, registered_ plug-in set plus extensibility machinery; ports/adapters is only the substrate (hexagonal-pattern-guide).
- Passing the registry into business logic is the service-locator anti-pattern — hidden dependencies, runtime not-found errors, untestable; resolve at the root and inject the plug-in.
- A churning plug-in contract forces edits across every plug-in — the open/closed violation wearing an interface; keep the contract narrow and versioned.
- The capability gate must name no concrete plug-in — fail-closed on a self-declared guarantee, or a new plug-in can violate an invariant silently.
- The OS microkernel and this architecture pattern share a name, not a definition — keep the two senses separate, see [references/metapatterns-lens.md](references/metapatterns-lens.md).

## Example

```go
// composition root: the only importer of concrete plug-ins
reg := Registry{Backends: map[string]Backend{
	"s3": s3.New(), "gcs": gcs.New(), // adding "azure" = one leaf + one line
}}
b, ok := reg.Backends[kind] // open/closed: no central switch to edit
if !ok { return ErrUnknownBackend }
if !b.Guarantees().Has(Durable) { return ErrPolicy } // fail-closed, names no plug-in
```

## Progressive Disclosure

- Read [references/core-plugins-registry.md](references/core-plugins-registry.md) - Load when designing the core, plug-in contract, and registry (open/closed)
- Read [references/capabilities-fail-closed.md](references/capabilities-fail-closed.md) - Load when gating plug-ins by self-declared capabilities, secure-by-default
- Read [references/binding-time.md](references/binding-time.md) - Load when deciding when a plug-in is selected (compile/load/run-time)
- Read [references/wiring.md](references/wiring.md) - Load when wiring plug-ins (dependency injection vs service locator, lazy activation, contract versioning)
- Read [references/metapatterns-lens.md](references/metapatterns-lens.md) - Load when relating microkernel to hexagonal/middleware/mesh (one non-canonical synthesis, with caveats)
