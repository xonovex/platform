---
name: agent-governance-guide
description: "Use when composing or inspecting agent and workflow execution across independent trigger, executor, host, control, evidence, and maturity dimensions. Triggers on workflow hooks, CI/CD triggers, script-plus-LLM execution, agent workflow skills, observe-vs-enforce policy, capability requirements, composition explanation, or questions about what Xonovex governance actually enforces."
compatibility: "Node.js 22+ is required for the executable workflow kernel and tests. Command plugins may add their own runtime requirements."
allowed-tools: "Read Bash(node:*)"
---

# Composable Workflow Controls

Build a workflow by selecting independent plugins. Start with no controls, then add only the controls and evidence behavior the owner actually wants.

## Essentials

- **Keep dimensions independent** - Trigger, executor, and host describe how work starts and runs; controls, evidence sinks, and maturity assessments are separate selections, see [references/architecture.md](references/architecture.md)
- **Treat every trigger as an adapter** - Manual use, an agent-harness hook, CI/CD, a schedule, a webhook, and a domain event all normalize to the same open trigger contract, see [references/adapters.md](references/adapters.md)
- **Select one executor plugin** - Script, script plus LLM, and agent plus workflow skill are provided command adapters, not kernel branches, see [references/composition.md](references/composition.md)
- **Select zero or more controls** - Each control declares before/after phases and each selection explicitly chooses `observe` or `enforce`, see [references/controls.md](references/controls.md)
- **Select evidence deliberately** - Each sink explicitly chooses whether a sink failure is ignored or fails the composition; no sink is mandatory by default
- **Require capabilities by name** - The kernel checks only declared `requiredCapabilities`; plugin capability declarations have no effect unless a composition or external assessor uses them
- **Derive maturity** - A1, A2, A3, or another scale comes from a caller-owned capability model after composition; maturity never selects execution behavior, see [references/maturity.md](references/maturity.md)
- **Explain before running** - Inspect selected plugins, available and missing capabilities, and exact enforcement points with the `explain` command

## Kernel Enforcement

The kernel enforces only schema validity, plugin registration, explicitly required capabilities, denials from controls selected as `enforce`, and evidence failure behavior selected as `fail`. It does not require a policy, maturity level, evidence sink, protected-target model, approval flow, provenance journal, escalation route, runtime host, or profile.

## Gotchas

- An observing control can report `deny` without blocking; this is intentional and visible in the result.
- An after-phase enforcing control can deny the composition result but cannot undo an executor side effect.
- A hook is one possible trigger and one possible enforcement point; it is not governance by itself.
- A template is trusted wiring, not a policy bundle. Review which plugins and modes it selects.
- Kubernetes `AgentPolicy` constrains its execution host only; it does not imply workflow maturity or agent governance.

## Progressive Disclosure

- Read [references/architecture.md](references/architecture.md) - Load when deciding boundaries or auditing what is intrinsic versus selected
- Read [references/composition.md](references/composition.md) - Load when creating a registry, invocation, or using the runtime CLI
- Read [references/controls.md](references/controls.md) - Load when authoring controls or choosing observe/enforce and before/after behavior
- Read [references/adapters.md](references/adapters.md) - Load when wiring manual, hook, CI/CD, schedule, webhook, or host adapters
- Read [references/maturity.md](references/maturity.md) - Load when defining or assessing A1/A2/A3 or another capability scale
