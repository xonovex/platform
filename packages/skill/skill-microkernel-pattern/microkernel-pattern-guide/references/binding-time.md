# Binding Time: When a Plug-In Is Chosen

Choosing a plug-in is not timeless — **binding time** is the lifecycle moment a variation point resolves to one concrete variant, and that moment dictates which mechanism is even appropriate.

## Binding time decides the mechanism

A variation point is a decision likely to change; a plug-in is one value of it. _When_ the decision resolves is its own design choice, separate from what varies and how. Pick the moment first — it rules out most techniques and leaves you the right one.

- **Compile-time** — the variant is fixed when the binary is built. Conditional compilation, build flags, generics/monomorphization. The unused path is never compiled in. Example: a batched-vs-single delivery channel selected by a build tag.
- **Load/deploy-time** — the variant is fixed when the program starts or is wired. Config, dependency injection, dynamic linking, _which_ plug-ins get registered. An absent backend is simply never added. Example: which channels (`email`, `sms`, `push`) the composition root registers, decided by deploy config.
- **Run-time** — the variant is fixed per request or per call, from input. A registry keyed by a user/request value. Example: `channel` chosen on each invocation by a `--channel` flag, looked up in the registry.

The spectrum runs compile-time → load/deploy-time → run-time, trading frozen simplicity for live flexibility as you move right.

## Bind as late as the cost allows

- **Later binding buys flexibility; earlier binding buys simplicity** — run-time selection lets one binary serve every combination, but it costs a registry, a key, and dispatch machinery. Compile-time selection has zero runtime cost but ships one frozen combination.
- **The rule** — bind as late as the cost of variation management _requires_, and no later. Later binding buys flexibility at the price of runtime machinery, so do not pay for flexibility you will not use. If the variant genuinely changes per request, bind at run time. If it is chosen once per deployment, bind at load time. If it never varies for a given build target, bind at compile time.

## Three separate questions

Keep these apart; conflating them is how a clean variation point acquires the wrong mechanism.

- **Binding unit — what varies.** The axis and its variants: `channel {email, sms, push}`.
- **Binding technique — how it is realized.** `#ifdef` / generics / DI / config / registry lookup.
- **Binding time — when it resolves.** Compile / load / run.

Binding time _selects_ the technique:

```
compile-time → build tag / generics / monomorphization
load-time    → DI / config / registry population
run-time     → registry lookup by key
```

Decide the unit and the time first; the technique follows. Reaching for a technique first (`#ifdef` because it is familiar) silently picks compile-time even when run-time was free. See [core-plugins-registry.md](core-plugins-registry.md) for the run-time lookup and [wiring.md](wiring.md) for the load-time population.

## Compositional vs annotative variability

Two ways to _realize_ a variation point — they age very differently.

- **Compositional (preferred)** — a variant is a self-contained plug-in behind the port. Selection is composition: register a leaf, look it up by key. Adding a variant adds a file; it edits nothing shared.
- **Annotative** — toggles, `#ifdef`, and feature flags sprinkled through one shared base. The variants live inline, interleaved, guarded by conditions. By construction this _scatters_ one decision across the base and _tangles_ unrelated decisions together — it is how a clean variation point rots into a god-module.

```go
// BAD — annotative: variants tangled into one shared body
func Send(cfg Config) error {
  // #ifdef PUSH ... #else ... #endif equivalent, at run time
  if cfg.Kind == "push" {
    openSocket(); pingDevice() // push-only, inline
  } else if cfg.Kind == "sms" {
    dialGateway(); splitParts() // sms-only, inline
  }
  if cfg.Mode == "bulk" { /* tangled second axis */ }
  return nil // every new channel edits THIS function
}
```

```go
// GOOD — compositional: each variant is a plug-in behind the port
type Channel interface { Send(msg Payload) (Receipt, error) }

ch := reg.Channels[cfg.Kind] // "push" plugs in by registering its leaf
return ch.Send(msg)          // no shared body to edit
```

The annotative form binds inline and forces every variant to share one function; the compositional form binds by lookup and keeps each variant in its own leaf. Annotative variability spans the **connascence-guide** spectrum at its worst — scattered connascence of meaning and algorithm across a shared body.

## Binding-time pitfalls

- **Bound too early** — hard-coding a variant with `#ifdef` or a build tag when load- or run-time was affordable. The cost: pluggability is gone. You now ship N binaries, or rebuild to switch a backend that a config line could have selected. A fail-closed capability check that varies per request cannot live at compile time.
- **Bound too late** — adding a registry, keys, and dispatch for a decision that resolves once per build. The cost: runtime machinery you never exercise, and a configuration surface that can express invalid states the build target never has. If `capability` is a fixed two-value set (`Encrypted`, `Bulk`) for a given deployment, a per-request lookup is ceremony.
- **The check** — name the moment the variant actually changes (per build target, per deploy, per request), then pick the _cheapest_ technique that resolves at that moment.

Back to [SKILL.md](../SKILL.md).
