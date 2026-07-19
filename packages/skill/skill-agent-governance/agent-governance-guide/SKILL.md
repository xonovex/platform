---
name: agent-governance-guide
description: "Use when defining, inspecting, onboarding, or validating governance for agent, model, tool, hook, policy, module, evidence, exception, emergency-exception, or privileged-operation behavior. Triggers on policy decision versus enforcement, semantic hook intents, harness capability matrices, bounded model or child-agent execution, authority attenuation, governance profiles, module conformance, drift, managed configuration, or advisory onboarding, even when the user doesn't say 'agent governance'."
compatibility: "Node.js 22+ is required for bundled conformance validators and the local decision-service walking skeleton; scripts read repository fixtures and write only explicit evidence or temporary test paths."
allowed-tools: "Read Bash(node:*) Bash(npx:*)"
---

# Agent Governance Contracts

Compose policy, execution, enforcement, enablement, evidence, and trust without requiring a lifecycle workflow or one harness, policy engine, provider, or configuration format.

## Core Principles

- **Independent plane** — governance can protect ordinary agent activity without workflow commands; workflow can operate without harness governance.
- **Decision before enforcement** — a policy decision point returns a versioned outcome; a separate enforcement point applies it and records evidence.
- **Deterministic first** — inspect authoritative state directly; bounded model or agent execution is explicit, validated, attenuated, observable, cancellable, and never silently authoritative.
- **Capabilities, not parity** — semantic event intents are portable, while every harness publishes versioned native support, blocking, ordering, context, limitations, and trust boundaries.
- **Non-weakening authority** — lower-authority configuration cannot silently weaken a mandatory higher-authority control; conflict and authority expansion fail visibly.
- **Transactional enablement** — discover, assess, recommend, preview, approve, apply, verify, roll back, monitor drift, upgrade, and remove through native mechanisms.

## Operations

- **Inspect** — report effective policies, modules, enforcement guarantees, authority, evidence, exceptions, and gaps — see [references/inspect.md](references/inspect.md)
- **Conformance** — validate executors, events, capability matrices, policies, modules, profiles, actors, exceptions, and evidence — see [references/conformance.md](references/conformance.md)
- **Advise onboarding** — discover and preview a compatible composition without mutating before authorization — see [references/onboarding.md](references/onboarding.md)
- **Evaluate drift** — compare intended and observed versions, capabilities, policy, evidence freshness, and native configuration — see [references/drift.md](references/drift.md)
- **Manage modules** — list, inspect, preview, enable, disable, upgrade, or remove modules through lifecycle-safe native adapters — see [references/module-management.md](references/module-management.md)
- **Validate providers** — check policy, configuration, evidence, replay, privacy, rollback, and drift ports without requiring one engine or format — see [references/provider-contracts.md](references/provider-contracts.md)
- **Adapt external enforcement** — map semantic policy intents onto CI, repository, deployment, admission, provider-permission, or policy-service controls — see [references/external-enforcement.md](references/external-enforcement.md)
- **Onboard external controls** — discover, preview, apply, verify, roll back, and adopt provider-native controls without requiring harness or workflow modules — see [references/external-enforcement-onboarding.md](references/external-enforcement-onboarding.md)
- **Adapt harnesses** — map semantic event intents to versioned native capabilities without inventing a universal hook file — see [references/harness-adapters.md](references/harness-adapters.md)
- **Onboard harness modules** — inspect, preview, authorize, apply, verify, roll back, disable, update, and detect drift through native configuration — see [references/harness-onboarding.md](references/harness-onboarding.md)
- **Replay the walking skeleton** — prove independent adoption, native onboarding, deterministic and bounded execution, external enforcement, provider evidence, drift, and rollback — see [references/walking-skeleton.md](references/walking-skeleton.md)
- **Compose policy bundles** — select domain-owned policies, interchangeable decision providers, adequate enforcement, explicit failure, and separate evidence — see [references/policy-bundles.md](references/policy-bundles.md)
- **Govern data and telemetry** — classify, minimize, route, redact, correlate, retain, authorize, and test model/agent/tool/policy/provider data — see [references/data-and-telemetry.md](references/data-and-telemetry.md)
- **Resolve catalog and inventory** — search compatible modules/profiles and publish the observed effective environment when selected — see [references/catalog-and-inventory.md](references/catalog-and-inventory.md)
- **Operate and learn** — canary, roll back, emergency-disable, detect drift, respond, retire, measure, and promote learning through review — see [references/operations-and-learning.md](references/operations-and-learning.md)

## Gotchas

- Installing a skill or module is not evidence that a control executes or blocks.
- A hook is one enforcement point, not a complete security boundary; mandatory controls may require independent external enforcement.
- Concurrent hooks may still produce sibling side effects after a denial; never assume serial order unless the adapter guarantees it.
- Legal and standards mappings require applicability and qualified or licensed-text review where recorded; a crosswalk is not certification or compliance.

## Progressive Disclosure

- Read [references/architecture.md](references/architecture.md) - Load when assigning plane ownership, adoption modes, trust zones, or dependency directions
- Read [references/execution.md](references/execution.md) - Load when selecting or constraining deterministic, model, agent, human, or external execution
- Read [references/autonomy.md](references/autonomy.md) - Load when grading or raising an autonomy posture, coupling a level to required oversight, or bounding escalation windows and safe defaults
- Read [references/actors.md](references/actors.md) - Load when recording an actor, deciding what a role string may mean, requiring independence or segregation of duties, or checking which actor rules code enforces
- Read [references/events-and-capabilities.md](references/events-and-capabilities.md) - Load when mapping semantic event intents to a harness capability matrix
- Read [references/policy-and-authority.md](references/policy-and-authority.md) - Load when defining decisions, enforcement, actors, evidence origin, exceptions, or emergency exceptions
- Read [references/provider-contracts.md](references/provider-contracts.md) - Load when defining or validating policy, configuration, telemetry, or evidence provider operations
- Read [references/external-enforcement.md](references/external-enforcement.md) - Load when enforcing policy through CI, repository rules, protected environments, admission, provider permissions, local hooks, or a policy decision service
- Read [references/external-enforcement-onboarding.md](references/external-enforcement-onboarding.md) - Load when discovering, previewing, applying, verifying, rolling back, or adopting external controls without harness or lifecycle modules
- Read [references/modules.md](references/modules.md) - Load when authoring a script, evaluator, agent launcher, external job, plugin, skill, MCP integration, or human-task module
- Read [references/module-trust.md](references/module-trust.md) - Load when reviewing executable-module trust, provenance, permissions, signatures/checksums/attestations, retries, rollback, or upgrades
- Read [references/profiles.md](references/profiles.md) - Load when composing governance, executor, enforcement, data, telemetry, or distribution requirements
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, recommending, previewing, applying, verifying, rolling back, upgrading, or removing configuration
- Read [references/inspect.md](references/inspect.md) - Load when inspecting effective governance and evidence
- Read [references/conformance.md](references/conformance.md) - Load when validating any governance contract or fixture
- Read [references/drift.md](references/drift.md) - Load when comparing intended and observed governance state
- Read [references/module-management.md](references/module-management.md) - Load when changing module lifecycle state
- Read [references/harness-adapters.md](references/harness-adapters.md) - Load when defining or reviewing an agent-harness adapter or capability matrix
- Read [references/harness-onboarding.md](references/harness-onboarding.md) - Load when advising or executing harness-specific setup, rollback, disable, update, or drift checks
- Read [references/harness-patterns.md](references/harness-patterns.md) - Load when selecting deterministic hooks, bounded model evaluators, bounded agent launchers, or organization-managed harness modules
- Read [references/enterprise-platforms.md](references/enterprise-platforms.md) - Load when composing independently owned enterprise work-item, source, CI, cloud-governance, and observability providers
- Read [references/walking-skeleton.md](references/walking-skeleton.md) - Load when proving the complete onboarding composition, replaying fresh-context evidence, or exercising disablement, drift, rollback, and adversarial cases
- Read [references/policy-bundles.md](references/policy-bundles.md) - Load when selecting security, privacy, accessibility, reliability, AI, supply-chain, data, cost, or regulated policy bundles and mapping them to deterministic, provider-native, or optional OPA decisions and enforcement
- Read [references/data-and-telemetry.md](references/data-and-telemetry.md) - Load when classifying, routing, redacting, correlating, retaining, sampling, or authorizing prompt/context/model/agent/MCP/tool/policy/CI/provider/privileged-operation data and telemetry
- Read [references/catalog-and-inventory.md](references/catalog-and-inventory.md) - Load when searching or composing the module/profile catalog, comparing adoption modes and authority zones, resolving conflicts/compatibility/trust, or inventorying the effective agent environment
- Read [references/operations-and-learning.md](references/operations-and-learning.md) - Load when updating, canarying, rolling back, disabling, supporting, measuring, learning from, responding to incidents in, or retiring governance modules and policies
