# Governance Module Contract

## Module kinds

| Kind                    | Semantics                                                                                                        |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Script                  | Deterministic command/API coordinator with explicit inputs, outputs, side effects, and exit behavior             |
| Bounded model evaluator | Deterministic coordinator plus fixed model context, structured output validation, budgets, and evidence labeling |
| Bounded agent launcher  | Explicit adaptive launch satisfying every constraint in [execution.md](execution.md)                             |
| External job            | Native CI, scanner, deployment, identity, monitoring, or policy-system execution and evidence                    |
| Plugin or extension     | Harness-native executable capability with declared events, permissions, configuration, and trust                 |
| Skill                   | Instructions, references, scripts, or assets; advisory/setup capability, never enforcement proof by installation |
| MCP integration         | Negotiated resources/prompts/tools with consent and host policy; not a complete sandbox or permission boundary   |
| Human task              | Accountable action with role, authority, expected evidence, due/failure/escalation behavior                      |

## Declaration

Every reusable module declares:

```text
identity and version
source and provenance
owner, support status, and lifecycle state
supported platforms and tested versions
classification: knowledge-only | advisory | evidence-producing | enforcing | configuration-changing | privileged
supported adoption modes and authority zones
semantic intents and native enforcement points
executor class and result/evidence contract
permissions, tools, filesystem, network, secrets, and data access
side effects, idempotency, reentrancy, and privileged status
ordering, concurrency, timeout, retry, cancellation, and recursion
failure behavior
sensitivity, redaction, retention, residency, and access
compatibility, dependencies, conflicts, and capability guarantees
upgrade, disable, rollback, drift, and retirement behavior
```

The declaration is semantic and may be represented in a native manifest, documentation, provider record, package metadata, or API resource. No YAML, JSON, file, plugin, or policy engine is universally required.

Trust review and provenance verification are separate from declaration completeness; apply [module-trust.md](module-trust.md) before loading executable project/user content or organization-managed distribution.

## Composition

- Resolve dependencies and conflicts before activation.
- Merge strengthening requirements additively.
- Reject unresolved conflicts, hidden authority expansion, and unsupported mandatory intents.
- Treat ordering as unknown unless the adapter guarantees it; make side effects idempotent and safe under retry/concurrency.
- Deny recursive or nested execution beyond declared limits.
- Reject moving versions, missing or mismatched provenance, permission expansion, unsafe retries/duplicates, unspecified timeouts, and unverifiable required rollback.
- Bind factories/adapters at a composition root; policy and business semantics never locate concrete modules through an ambient registry.
