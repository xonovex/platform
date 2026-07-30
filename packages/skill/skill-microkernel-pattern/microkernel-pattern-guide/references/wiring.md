# Wiring: Injection vs Service Locator

A `key → factory` registry belongs at the composition root. Resolve `key → port` once there and inject the one port the business logic needs. Passing the whole registry inward is service location: it hides the real dependency, fails at **runtime** on a missing key instead of at compile time, and cannot be tested without the whole map.

```go
// BAD, service locator: registry threaded into core logic
func Send(reg Registry, kind string) error {
  ch := reg.Channels[kind]() // real dependency unknowable; runtime KeyNotFound; untestable
  return ch.Deliver()
}

// GOOD, dependency injection: resolve at the root, inject the port
func Send(ch Channel) error { return ch.Deliver() } // checkable, fakeable
ch, ok := reg.Channels[kind]
if !ok { return fmt.Errorf("unknown channel %q", kind) }
return Send(ch())
```

- **Lazy activation**: bind key → constructor, construct only the variant selected; eager construction drags every leaf's heavy init (SMTP connection, push-gateway auth) into startup.
- **Version a churning contract**: an open plug-in set behind a changing port breaks every leaf at once (open/closed violation wearing an interface). Keep the port narrow (only what every variant provides); when the contract genuinely must evolve, version it (`ChannelV2`) so existing leaves keep working while new ones opt in.
- **Prefer central registration over `init()` self-registration**: central wiring is visible, ordered, and testable; self-registration hides wiring behind import order.

See [core-plugins-registry.md](core-plugins-registry.md) and **connascence-guide**. Back to [SKILL.md](../SKILL.md).
