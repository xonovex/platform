# Core, Plug-ins, and the Registry

The microkernel pattern is a minimal core plus an open set of independent plug-ins joined by one registry — keep the core empty of features, push every feature into a plug-in, and let the registry be the only thing that knows both.

## The three parts

- **Core** — holds the minimum to run: the orchestration loop, the shared ports, the registry lookup. It knows _no_ concrete plug-in and contains _no_ feature logic. If it grows a branch per feature, the feature belongs in a plug-in.
- **Plug-in** — an independent component that adds one feature behind a standard contract. It depends on the port, never on the core's internals or on a sibling plug-in.
- **Registry** — the connector. It records, per plug-in, a **name** (the selection key), a **data contract** (the port it satisfies), and a **connection** (a factory the core calls). The core reaches every plug-in through the registry and only through it.

```go
type Registry struct {
  Channel map[string]func() Channel    // key → factory (the connection)
  Render  map[string]func() Renderer   // each entry: name + contract + factory
}
```

A microkernel is _not_ "just a hexagon." Ports-and-adapters is the substrate it stands on; the microkernel adds the OPEN, REGISTERED set and the machinery to extend it. They are parallel, distinct patterns — for ports, adapters, and dependency inversion see **hexagonal-pattern-guide**; this skill builds on that and supplies the registry and the open extension model.

## Open/closed: add a plug-in, don't edit a switch

A central `switch` is closed to extension: every new plug-in edits the same function, so the core learns all plug-ins and grows a branch wherever they differ. Adding a feature should add a component and one registration — and touch nothing already written.

```go
// BAD — central switch every new channel must edit, inside the core
func (s Send) Channel(kind string) Channel {
  switch kind {
  case "email": return email.New()
  case "sms":   return sms.New()
  // adding "push" forces an edit HERE
  }
}
```

```go
// GOOD — registry lookup by key; existing code is untouched
ch, ok := reg.Channel[kind] // "push" plugs in by registering its factory once
```

**Open/closed has a scope smaller than the codebase.** Open exactly the variation point that actually varies; leave the rest concrete. A registry behind a one-implementation axis is speculative generality you pay to maintain — apply the rule-of-three before you open a seam at all.

## Plug-ins are reached through standard contracts

- **One narrow port per axis** — the port lives in the axis's shared package; plug-ins implement it; nothing outside the registry names a concrete type. Pluggability is dependency inversion applied to one variation point: callers depend on the port, the plug-ins depend on the port, the two never see each other.
- **Pass neutral values across the seam** — when one plug-in hands data to another (a renderer's output feeding a channel), pass a neutral message payload defined on the port, not a concrete type. A wide or leaky port forces every plug-in to know the others.
- **Self-declared capabilities, not identity** — let each plug-in advertise its own capabilities as data the core unions generically; the core asks "does the selected plug-in declare `Encrypted`?", never "is this the email plug-in?". The fail-closed gate that consumes those declarations is in [capabilities-fail-closed.md](capabilities-fail-closed.md).

## The registry is the single composition root

- **One map per axis, key → factory** — the registry binds the selection key to a constructor and is the _only_ importer of every concrete plug-in. Fan-in stays one-way: plug-ins import the port, the registry imports the plug-ins, business logic imports neither.
- **Bind a factory, not an instance** — map the key to a constructor and call it only for the one plug-in actually selected (lazy activation). Building every plug-in up front does the work of N channels to use one and drags every plug-in's heavy init into startup.

```go
// registry.go — the one composition root, sole importer of concrete plug-ins
func DefaultRegistry() Registry {
  return Registry{
    Channel: map[string]func() Channel{
      "email": email.New, "sms": sms.New, "push": push.New,
    },
    Render: map[string]func() Renderer{"html": html.New, "plain": plain.New},
  }
}
```

Resolve `key → port` once at the root and inject the resolved port into business logic. Passing the whole registry inward is the service-locator anti-pattern: it hides the real dependency, fails at runtime on a missing key instead of at compile time, and cannot be tested without the whole map. Prefer explicit central registration over `init()` self-registration — central wiring is visible, ordered, and testable; self-registration hides wiring behind import order. For the wiring rules and registration styles in full, see [wiring.md](wiring.md). For coupling and connascence vocabulary see **connascence-guide**.

## The pluggability test

- **A new plug-in touches exactly three things** — its own component, one registration line, and its own per-plug-in options. Nothing in the core, nothing in a sibling.
- **If it forces edits elsewhere, it is not pluggable** — an edit in the core or a sibling means a seam leaked: a `switch` that should be a map, a capability the core hardcoded, a port method only some plug-ins implement. Find the leaked seam and push it back behind the port.

Back to [SKILL.md](../SKILL.md).
