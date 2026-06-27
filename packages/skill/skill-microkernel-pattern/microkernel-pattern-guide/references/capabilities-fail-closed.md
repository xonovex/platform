# Capabilities and Fail-Closed

Each plug-in self-declares its own capabilities as data; the core unions them generically and gates on what is declared — never on which plug-in it is — so a plug-in lacking a required capability is rejected with zero edits.

## Self-declared capabilities, generically unioned

- **Every plug-in advertises its own capabilities** — a plug-in is the authority on what it provides. It exposes its capabilities through the port, and the core collects them without knowing which concrete plug-in answered.
- **The core unions generically, names nothing concrete** — the core folds the selected plug-in's declared set into a lookup and asks _"is capability X present?"_, never _"is this the push plug-in?"_. No concrete name, enum, or type appears in the gate.
- **A future plug-in costs zero core edits** — because the check reads declared data, a plug-in added later that declares the required capability passes, and one that does not is rejected — both without touching the core. This is the pluggability test applied to security: see [core-plugins-registry.md](core-plugins-registry.md).

```go
type Capability string // Encrypted, Bulk, Ordered

type Channel interface {
  Capabilities() []Capability // each plug-in self-declares its own, as data
  // ...
}
```

## Ground the gate in secure-by-default principles

The gate is the same cross-cutting policy woven once at the composition root, not scattered into each leaf. Three secure-by-default principles make it sound.

- **Fail-safe defaults** — decide on the basis of _permission, not exclusion_: grant only when a required capability is positively present, deny otherwise. Default deny, never default allow; a missing or unreadable declaration is a denial, not a pass.
- **Principle of least authority** — rely only on a plug-in's _declared_ capabilities, never on its identity or ambient authority. The core must never say "trust the push plug-in" or "it's internal, so it's fine"; it asks "does the selected plug-in declare `Encrypted`?" and grants exactly that.
- **Complete mediation** — the same generic check runs uniformly for _every_ plug-in, with no bypass and no special case. Nothing is grandfathered through; that uniformity is precisely why next month's plug-in is mediated identically to today's.

## Declarative beats imperative

- **A manifest the core reads, not code the core trusts** — the strongest form of capability declaration is _declarative_: the plug-in states its capabilities as inert data the core inspects, rather than running plug-in-supplied imperative code that asserts its own compliance.
- **Why declarative wins** — data the core reads is uniform, inspectable, and cannot lie by side effect; trusted imperative code can branch on context, claim a capability it does not honor, or do work the core never asked for. Capabilities-as-data keep the trust boundary on the core's side.

```text
BAD  — plug-in runs code the core trusts to self-certify
  ch.AssertEncrypted()   // opaque: did it really check? what else did it run?
GOOD — plug-in declares data the core reads and decides on
  have := index(ch.Capabilities())   // inert, inspectable, cannot do work
  if !have[Encrypted] { deny }
```

## A fail-closed gate

Read the selected plug-in's self-declared capabilities, index them, and reject on the first required one that is absent. Name the unmet _capability_, never a plug-in.

```go
func enforce(required []Capability, declared []Capability) error {
  have := make(map[Capability]bool, len(declared))
  for _, c := range declared { // union the plug-in's own declarations
    have[c] = true
  }
  for _, c := range required {
    if !have[c] { // fail closed on the first unmet capability
      return fmt.Errorf("required capability %s not provided", c) // names the capability, never a plug-in
    }
  }
  return nil
}
```

- **The gate names the requirement, not the provider** — the error reads "required capability `Encrypted` not provided", never "plug-in `sms` rejected". That is capability negotiation, not a hardcoded allowlist the core must remember to update.
- **No allowlist of blessed plug-ins** — an `if kind == "push" || kind == "email"` gate closes the open set: every new plug-in must edit it, and a forgotten edit silently denies a compliant plug-in or admits a non-compliant one. The capability check has neither failure mode.

```go
// BAD — hardcoded allowlist: closed to extension, must edit per plug-in
if kind != "push" && kind != "email" {
  return errors.New("channel not allowed")
}
// GOOD — generic capability gate: open set, names no plug-in
if err := enforce(policy.Required, ch.Capabilities()); err != nil {
  return err
}
```

## Where this sits

- Resolve the plug-in once at the composition root, then run the gate on its declared capabilities before handing the resolved port to business logic — see [core-plugins-registry.md](core-plugins-registry.md).
- The wider pattern — a minimal core, an open set of plug-ins behind a port, and a registry as the single composition root — is the subject of the **microkernel-pattern-guide**; the coupling vocabulary behind "names the capability, not the plug-in" belongs to **connascence-guide**; the port-and-adapter boundary the plug-ins implement is **hexagonal-pattern-guide**.

Back to [SKILL.md](../SKILL.md).
