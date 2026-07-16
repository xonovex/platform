# Core, Plug-ins, and the Registry

Minimal core + open set of plug-ins + one registry. The core holds only orchestration, shared ports, and the registry lookup — no concrete plug-in, no feature branch. Each plug-in adds one feature behind a port and depends on the port, never on the core or a sibling.

The **registry** records per plug-in a **name** (selection key), a **contract** (the port it satisfies), and a **connection** (a factory). It is the sole importer of concrete plug-ins; fan-in stays one-way (plug-ins → port; registry → plug-ins; business logic → neither).

## Rules

- **Open/closed: map, not switch** — a central `switch kind` is closed to extension; every new plug-in edits it. A `key → factory` lookup adds a plug-in with one registration line and touches nothing already written.
- **Bind a factory, not an instance** — map the key to a constructor, call it only for the selected plug-in (lazy activation), never a pre-built instance of every variant.
- **One narrow port per axis** — the port lives in the axis's shared package. Pass neutral values (a message payload defined on the port) across the seam, never a concrete type; a wide/leaky port forces every plug-in to know the others.
- **Self-declared capabilities, not identity** — the core asks "does the selected plug-in declare `Encrypted`?", never "is this the email plug-in?" (see [capabilities-fail-closed.md](capabilities-fail-closed.md)).
- **Close only what varies** — apply the rule-of-three before opening a seam; a registry behind a single-implementation axis is speculative generality.

## Pluggability test

A new plug-in touches exactly three things: its own component, one registration line, its own options. If it forces an edit in the core or a sibling, a seam leaked (a `switch` that should be a map, a capability the core hardcoded, a port method only some plug-ins implement) — push it back behind the port.

```go
// registry.go — the one composition root, sole importer of concrete plug-ins
func DefaultRegistry() Registry {
  return Registry{
    Channel: map[string]func() Channel{ // key → factory (lazy)
      "email": email.New, "sms": sms.New, "push": push.New, // "push" plugs in by one line
    },
  }
}
ch, ok := reg.Channel[kind] // open/closed: no switch to edit
```

Resolve `key → port` once at the root and inject the port; passing the whole registry inward is service location (see [wiring.md](wiring.md)). For ports/adapters and dependency inversion see **hexagonal-pattern-guide**; for coupling vocabulary see **connascence-guide**.

Back to [SKILL.md](../SKILL.md).
