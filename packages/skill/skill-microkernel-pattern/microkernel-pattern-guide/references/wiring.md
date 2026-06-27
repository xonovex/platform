# Wiring: Injection vs Service Locator

A key→factory registry belongs at the composition root; threading it into business logic turns dependency injection into a service locator.

## Resolve at the root, inject the plug-in

- **A registry is a composition-root tool, not a runtime dependency** — a `key → factory` map binds a selection key to a constructor and is the one place that names concrete plug-ins. Wiring belongs there; the connascence of name on string keys stays local to a single file, the strongest locality the registry can have.
- **Passing the registry into business logic is service location** — a function handed the whole registry and left to look up its own plug-in hides its real dependency, fails at _runtime_ with a missing key instead of at compile time, and cannot be tested without standing up the whole map.
- **Inject the resolved port, not the lookup** — resolve `key → plug-in` once at the root and hand the business logic the one port it needs. Then the dependency is visible in the signature, checkable by the compiler, and fakeable in a test. This is the composition root doing dependency inversion — see **hexagonal-pattern-guide** for where that root sits relative to the ports.

```go
// BAD — service locator: registry threaded into core logic
func Send(reg Registry, kind string) error {
  ch := reg.Channels[kind]() // what does Send actually need? unknowable from the signature
  // runtime KeyNotFound if kind is unregistered; untestable without the whole map
  return ch.Deliver()
}
```

```go
// GOOD — dependency injection: resolve at the root, inject the one port
func Send(ch Channel) error { // needs exactly this — checkable, fakeable
  return ch.Deliver()
}

// at the composition root, once:
ch, ok := reg.Channels[kind]
if !ok { return fmt.Errorf("unknown channel %q", kind) }
return Send(ch())
```

Business logic depends on the port (connascence of name on one interface), never on a registry whose shape it must know. For why that asymmetry holds the import graph one-way, see [core-plugins-registry.md](core-plugins-registry.md).

## Lazy activation: bind a factory, not an instance

- **Map key → factory, construct on select** — bind the selection key to a constructor and call it only for the variant actually chosen.
- **Eager construction does N backends' work to use one** — building every variant up front drags each leaf's heavy init (opening an SMTP connection, authenticating a push gateway) into startup whether or not it is selected. Defer activation to the moment of selection.

```go
// BAD — eager: every leaf is constructed (and initialized) at startup
reg.Channels := map[string]Channel{
  "email": email.New(), "sms": sms.New(), // both run their init now
}

// GOOD — lazy: store the constructor, call only the selected one
reg.Channels := map[string]func() Channel{
  "email": email.New, "sms": sms.New, // nothing runs until selected
}
```

## Contract versioning: keep the port narrow and stable

- **An open plug-in set behind a churning port breaks every leaf at once** — if the port or the value passed across it keeps changing, each change forces an edit across every existing plug-in. That is the open/closed violation wearing an interface: the set is "open" to new plug-ins but every change closes over all of them.
- **Publish a narrow port** — a new plug-in should need only the port and the neutral message payload handed between channels, nothing more. A wide port makes leaves implement methods they do not use; segregate it to what _every_ variant truly provides.
- **Version a contract that genuinely must evolve** — when the data contract must change, version it (`ChannelV2`, a versioned message payload) so existing leaves keep working while new ones opt in, instead of breaking the whole set on each revision.

## OCP scope: close only what varies

- **Open/closed has a scope smaller than the codebase** — apply it to _the variation point that actually varies_. Close that one behind a port and a registry; leave everything else concrete.
- **Do not open extension points "in case"** — an unused registry behind a single-implementation axis is speculative generality: the wrong abstraction, paid for in maintenance with no variant to show for it. Earn the seam with a real second implementation before you carve it.
- **The tell of a leaked seam** — if adding a plug-in forces an edit in the core or a sibling, the variation point was not actually closed: a `switch` masquerading as a map, a capability the core hardcoded, a port method only some variants implement. Push that seam back behind the port.

For the registry's place in the import graph and the connascence vocabulary used here, see **connascence-guide** and [core-plugins-registry.md](core-plugins-registry.md). Back to [SKILL.md](../SKILL.md).
