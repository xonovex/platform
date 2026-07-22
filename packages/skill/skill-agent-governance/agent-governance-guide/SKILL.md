---
name: agent-governance-guide
description: "Use when composing or inspecting agent and workflow execution across independent trigger, executor, host, control, evidence, and maturity dimensions. Triggers on workflow hooks, CI/CD triggers, script-plus-LLM execution, agent workflow skills, observe-vs-enforce policy, capability requirements, composition explanation, or questions about what agent governance does and does not enforce, even when the user doesn't say 'agent governance'."
---

# Composable Agent Controls

Describe a governance overlay for an already selected operation by choosing independent capabilities. Start with no controls, then add only the controls and evidence behavior the accountable owner requires.

## Essentials

- **Keep the operation external** - Accept the caller's exact operation and subject; governance never selects, sequences, or changes operation semantics
- **Keep dimensions independent** - Trigger, executor, and host describe how work starts and runs; controls, evidence sinks, and maturity assessments remain separate selections, see [references/architecture.md](references/architecture.md)
- **Treat every trigger as an adapter** - Normalize manual use, hooks, CI/CD, schedules, webhooks, and domain events to one trigger contract, see [references/adapters.md](references/adapters.md)
- **Select one executor capability** - Deterministic commands, command-plus-model execution, and agents with workflow skills are adapters rather than governance branches, see [references/composition.md](references/composition.md)
- **Select zero or more controls** - Declare each control phase and choose `observe` or `enforce` explicitly, see [references/controls.md](references/controls.md)
- **Name the adoption mode** - Distinguish knowledge, advice, enforcement, model evaluation, specialist agents, and organization-managed delivery before mapping them to a harness, see [references/adoption-modes.md](references/adoption-modes.md)
- **Select evidence deliberately** - State whether each evidence-sink failure is ignored or fails the composition; no sink is mandatory by default
- **Require capabilities by name** - Treat a declared capability as available only when the effective composition or an independent assessment verifies it
- **Derive maturity** - Assess A1, A2, A3, or another scale after composition; maturity never selects execution behavior, see [references/maturity.md](references/maturity.md)
- **Explain before execution** - Record selected capabilities, missing requirements, enforcement points, failure behavior, and unresolved assumptions before work starts

## Enforcement Boundary

This skill explains a governance composition; it does not execute or enforce one. An implementation may claim enforcement only for controls attached to an adequate protected enforcement point with verified failure behavior. Schema validity, registered capabilities, policy presence, maturity labels, hooks, evidence sinks, or approval records alone do not prove enforcement.

The selected operation remains owned by **workflow-guide**. Load **reliability-guide** for failure, recovery, and operating objectives; **security-assurance-guide** for selected security criteria and evidence; and **ai-governance-guide** when the governed subject is an AI system. This skill only composes those capabilities and their enforcement modes.

## Gotchas

- An observing control can report `deny` without blocking; this is intentional and visible in the result.
- An after-phase enforcing control can deny the composition result but cannot undo an executor side effect.
- A hook is one possible trigger and one possible enforcement point; it is not governance by itself.
- A composition record is trusted wiring, not a policy bundle. Review which capabilities and modes it selects.
- Kubernetes `AgentPolicy` constrains its execution host only; it does not imply workflow maturity or agent governance.

## Progressive Disclosure

- Read [references/architecture.md](references/architecture.md) - Load when deciding boundaries or auditing what is intrinsic versus selected
- Read [references/composition.md](references/composition.md) - Load when creating or reviewing a governance composition and its decision record
- Read [references/controls.md](references/controls.md) - Load when authoring controls or choosing observe/enforce and before/after behavior
- Read [references/adoption-modes.md](references/adoption-modes.md) - Load when choosing a harness adoption mode or reviewing an enforcement claim
- Read [references/adapters.md](references/adapters.md) - Load when wiring manual, hook, CI/CD, schedule, webhook, or host adapters
- Read [references/maturity.md](references/maturity.md) - Load when defining or assessing A1/A2/A3 or another capability scale
