# Capabilities and Fail-Closed

The core gates on a plug-in's declared capabilities, never on which plug-in it is. A plug-in lacking a required capability is rejected with zero core edits.

- **Complete mediation**: the same generic check runs for every plug-in, no bypass, nothing grandfathered.

**Declarative beats imperative**: the plug-in states capabilities as inert data the core reads (`ch.Capabilities()`), not code the core trusts to self-certify (`ch.AssertEncrypted()`). Data cannot lie by side effect; trusted imperative code can claim a capability it does not honor.

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

The error names the requirement (`Encrypted`), never the provider (`sms`). Run the gate at the composition root on the resolved plug-in's declared capabilities before handing the port to business logic.

See [core-plugins-registry.md](core-plugins-registry.md), **connascence-guide**, **hexagonal-pattern-guide**. Back to [SKILL.md](../SKILL.md).
