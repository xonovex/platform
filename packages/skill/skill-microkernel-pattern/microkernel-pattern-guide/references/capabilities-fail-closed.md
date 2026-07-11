# Capabilities and Fail-Closed

Each plug-in self-declares its capabilities as **data**; the core unions them generically and gates on what is declared — never on which plug-in it is. A plug-in lacking a required capability is rejected with zero core edits; a future plug-in that declares it passes with zero edits.

Ground the gate in three secure-by-default principles:

- **Fail-safe defaults** — grant only when a required capability is positively present; default deny. A missing or unreadable declaration is a denial, not a pass.
- **Least authority** — decide on declared capabilities, never on identity or ambient authority ("it's internal, so it's fine").
- **Complete mediation** — the same generic check runs for every plug-in, no bypass, nothing grandfathered.

**Declarative beats imperative** — the plug-in states capabilities as inert data the core reads (`ch.Capabilities()`), not code the core trusts to self-certify (`ch.AssertEncrypted()`). Data cannot lie by side effect; trusted imperative code can claim a capability it does not honor.

```go
type Capability string // Encrypted, Bulk, Ordered

func enforce(required, declared []Capability) error {
  have := make(map[Capability]bool, len(declared))
  for _, c := range declared { have[c] = true } // union generically
  for _, c := range required {
    if !have[c] { // fail closed on first unmet
      return fmt.Errorf("required capability %s not provided", c) // names the capability, never a plug-in
    }
  }
  return nil
}
```

The error names the requirement (`Encrypted`), never the provider (`sms`). A hardcoded allowlist (`if kind == "push" || kind == "email"`) closes the open set: every new plug-in must edit it, and a forgotten edit silently denies a compliant plug-in or admits a non-compliant one. Run the gate at the composition root on the resolved plug-in's declared capabilities before handing the port to business logic.

See [core-plugins-registry.md](core-plugins-registry.md), **connascence-guide**, **hexagonal-pattern-guide**. Back to [SKILL.md](../SKILL.md).
