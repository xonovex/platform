# Trigger and Host Adapters

Normalize a native event to `{kind, reference, actor?, idempotencyKey, data?, subject}` and bind it to a trusted governance composition. Trigger `kind` is open text so product-specific adapters do not require a governance-kernel release.

Common mappings include:

| Native origin            | Example trigger kind       |
| ------------------------ | -------------------------- |
| Human command            | `manual`                   |
| Agent harness hook       | `hook/claude/pre-tool-use` |
| CI/CD hook               | `ci/github/pull-request`   |
| Provider webhook         | `provider/github/webhook`  |
| Schedule                 | `schedule/nightly`         |
| Sensor                   | `sensor/dependency-drift`  |
| Agent-created child work | `agent/subtask`            |

Authenticate and minimize native events before normalization. Keep executable selection, controls, and evidence in a trusted template so untrusted event payloads cannot choose their own enforcement.

The host is independent too. A local process, CI runner, agent harness, or Kubernetes operator can host the same composition. Host hardening such as image pinning, runtime classes, resource limits, network policy, and secrets belongs to that host. Do not translate host hardening into a workflow maturity claim.
