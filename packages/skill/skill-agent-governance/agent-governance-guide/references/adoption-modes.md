# Harness Adoption Modes

Use this taxonomy before selecting a product-native hook, plugin, handler, or agent mechanism. The harness-specific guide owns the concrete mapping and its caveats; this guide owns the semantic mode and the enforcement claim.

| Mode                 | Required result                                         | Maximum claim                                     |
| -------------------- | ------------------------------------------------------- | ------------------------------------------------- |
| Knowledge-only       | Project instructions or an Agent Skill                  | Model-visible guidance only                       |
| Advisory             | Post-event observer or other non-blocking handler       | Evidence or feedback only                         |
| Enforcing            | Deterministic pre-event control with verified coverage  | Only the documented intercepted operations        |
| Model evaluator      | Declared model executor with a closed, validated result | Bounded inference, never silent authority         |
| Specialist agent     | Explicit bounded launcher with attenuated authority     | Only its declared tools, depth, host, and budgets |
| Organization-managed | Managed configuration plus pinned executable delivery   | Only verified native and external guarantees      |

## Selection

1. Start at knowledge-only and move to a stronger mode only when the requested outcome needs it.
2. Separate the trigger from the executor. A hook that launches a model is a model evaluator, not a deterministic control.
3. For enforcement, identify the protected operation, interception point, matcher, failure signal, bypass paths, and runtime evidence.
4. For model evaluators and specialist agents, declare provider, data transfer, output schema, timeout, cancellation, depth, tool authority, and budget.
5. For organization-managed adoption, verify both native precedence and executable/configuration distribution. Neither proves the other.

## Translation record

For every harness-native mapping, record the native event, handler or runner, matcher, input schema, output or exit behavior, timeout, concurrency, permissions, data exposure, evidence, rollback, and unsupported paths. Installation or discovery proves availability only; it does not prove execution or enforcement.
