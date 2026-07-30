# Binding Time: When a Plug-In Is Chosen

**Binding time** is the lifecycle moment a variation point resolves to one variant. Pick the moment first: it rules out most techniques and leaves the right one.

- **Compile-time**: fixed when the binary is built (build tags, generics/monomorphization); the unused path is never compiled in.
- **Load/deploy-time**: fixed at startup/wiring (config, DI, dynamic linking, _which_ plug-ins the composition root registers); an absent backend is simply never added.
- **Run-time**: fixed per request/call from input (registry keyed by a request value, e.g. `--channel`).

The spectrum runs compile → load → run, trading frozen simplicity for live flexibility.

**The rule: bind as late as the cost of variation management requires, and no later.** Later binding buys flexibility at the price of runtime machinery (registry, key, dispatch). Don't pay for flexibility you won't use.

Keep three questions apart:

- **Binding unit, what varies:** `channel {email, sms, push}`.
- **Binding technique, how:** `#ifdef` / generics / DI / config / registry lookup.
- **Binding time, when:** compile / load / run.

Time _selects_ technique: `compile → build tag/generics` · `load → DI/config/registry population` · `run → registry lookup`. Reaching for a technique first (`#ifdef` because familiar) silently picks compile-time even when run-time was free.

**Compositional vs annotative**: prefer compositional (each variant a self-contained plug-in behind the port; adding one adds a file, edits nothing shared) over annotative (toggles/`#ifdef`/flags scattered through one shared base, which scatters one decision and tangles unrelated ones into a god-module).

```go
// BAD, annotative: variants tangled into one shared body
func Send(cfg Config) error {
  if cfg.Kind == "push" { openSocket(); pingDevice() }
  else if cfg.Kind == "sms" { dialGateway(); splitParts() }
  if cfg.Mode == "bulk" { /* tangled second axis */ }
  return nil // every new channel edits THIS function
}

// GOOD, compositional: each variant is a plug-in behind the port
ch := reg.Channels[cfg.Kind]
return ch.Send(msg) // no shared body to edit
```

**Pitfalls:** bound too late = registry/keys/dispatch for a decision that resolves once per build, plus a config surface that can express invalid states. Name the moment the variant actually changes, then pick the cheapest technique that resolves then.

See [core-plugins-registry.md](core-plugins-registry.md) (run-time lookup), [wiring.md](wiring.md) (load-time population), **connascence-guide**. Back to [SKILL.md](../SKILL.md).
