---
name: workflow-guide
description: "Use when defining, executing, inspecting, composing, or validating lifecycle workflow capabilities, provider-native result contracts, ephemeral result handles, workflow profiles, development and assurance work, human Acceptance, privileged Integration, Transition, Release, Observation, Incident, Corrective Action, Retirement, or cross-provider handoffs. Triggers on workflow result, PhaseResultHandle, exact-revision assurance, evidence freshness, authorization drift, rollback, break-glass, provider-native evidence, or workflow conformance, even when the user doesn't say 'workflow architecture'."
---

# Composable Workflow Contracts and Operations

Keep lifecycle meaning stable while methods, executors, providers, topology, and governance vary independently.

## Core Principles

- **Semantic results, native persistence** — each capability owns required meaning; provider skills own representation, authentication, native identifiers, revisions, and side effects.
- **Opaque handoffs** — exchange provider context plus opaque native references; never invent a central workflow identity, database, or universal persisted envelope.
- **Ephemeral reconstruction** — use `PhaseResultHandle` only in runtime handoffs and reconstruct it from provider-native state after context loss.
- **Profiles compose** — profiles include, omit, sequence, parallelize, loop, and present capabilities without erasing their independent results or publication boundaries.
- **Independent composition** — resolve method, provider, workspace, policy, and learning axes separately; workflow contracts do not require governance hooks.
- **Least-adaptive execution** — use deterministic tools for authoritative facts and mechanical work, bounded models for narrow transforms, and bounded agents only for adaptive multi-step work.
- **Exact-revision assurance** — inventory, review, QA, assessment, and enforcement evidence is valid only for its recorded subject, policy, evaluator, and environment versions.
- **Authority stays explicit** — agents may assemble evidence or advise, but human Acceptance and protected target mutations require exact-scope authorization and provider/external enforcement.

## Lifecycle Operations

- **Development run** — execute independently assigned development work and publish one result per assignment — see [references/develop-run.md](references/develop-run.md)
- **Development consolidation** — combine completed Development results into a development workspace without claiming Integration — see [references/develop-consolidate.md](references/develop-consolidate.md)
- **Development abandonment** — stop one assignment while preserving partial state, evidence, and retry safety — see [references/develop-abandon.md](references/develop-abandon.md)
- **Deliverable publication** — publish a reviewable candidate through its native provider at an immutable revision — see [references/deliver-publish.md](references/deliver-publish.md)
- **Inventory generation** — deterministically inventory exact subjects as software, AI, ML, cryptographic, service, or agent-environment components — see [references/inventory-generate.md](references/inventory-generate.md)
- **Assessment run** — evaluate any exact workflow result against versioned criteria — see [references/assessment-run.md](references/assessment-run.md)
- **Review run** — produce deliverable-specific findings with explicit assessor origin and independence — see [references/review-run.md](references/review-run.md)
- **QA run** — produce deliverable-specific test and environment evidence — see [references/qa-run.md](references/qa-run.md)

## Operational Lifecycle

- **Acceptance validation** — assemble fresh exact-revision evidence without fabricating human sign-off — see [references/acceptance-validate.md](references/acceptance-validate.md)
- **Acceptance decision** — record accountable human Acceptance with exact target, evidence, policy, and expiry bindings — see [references/acceptance-decide.md](references/acceptance-decide.md)
- **Integration preflight** — validate authorization and protected target capabilities without mutation — see [references/integration-validate.md](references/integration-validate.md)
- **Integration execution** — mutate an accepted target only through externally enforced explicit capability — see [references/integration-run.md](references/integration-run.md)
- **Transition run** — plan, execute, verify, or roll back data, user, provider, flag, support, and resilience change — see [references/transition-run.md](references/transition-run.md)
- **Release run** — execute or recover a release through controlled automation and protected environments — see [references/release-run.md](references/release-run.md)
- **Observation run** — publish exact-window operational evidence from system, user, security, AI, cost, accessibility, and delivery signals — see [references/observe-run.md](references/observe-run.md)
- **Incident run** — declare, update, contain, recover, or close an incident with explicit emergency authority and applicability review — see [references/incident-run.md](references/incident-run.md)
- **Corrective action run** — plan, execute, verify, and close corrections with effectiveness and learning evidence — see [references/corrective-action-run.md](references/corrective-action-run.md)
- **Retirement run** — retire models, data, credentials, features, APIs, infrastructure, dependencies, or provider configuration with authoritative verification — see [references/retirement-run.md](references/retirement-run.md)

## Contract Operations

- **Inspect** — resolve native references and report effective capabilities, profile topology, result freshness, and gaps — see [references/inspect.md](references/inspect.md)
- **Conformance** — validate result semantics, ephemeral handles, provider handoffs, profile topology, publication boundaries, evidence, and cross-plane requirements — see [references/conformance.md](references/conformance.md)

## Gotchas

- A runtime trace or session ID correlates execution; it is not a workflow identity.
- A local path is one provider-native reference, not the fallback for an unavailable explicitly selected hosted provider.
- Composite labels such as Discover or Review are presentation; constituent capability results remain independently publishable.
- Consolidating parallel Development results changes a development workspace; it is not authorized Integration into an accepted target.
- A model may enrich a non-authoritative inventory description but may not invent a component, identity, version, digest, relationship, or provenance.
- A passing check, summary, label, comment, agent response, or ordinary tool call is neither accountable human Acceptance nor authorization for a privileged target change.
- Successful rollback or recovery does not turn the failed operation into a successful one; preserve both results and their native evidence.
- Installing this skill explains contracts but does not enforce policy or prove a control is active.

## Progressive Disclosure

- Read [references/architecture.md](references/architecture.md) - Load when assigning ownership, checking adoption modes, or deciding allowed dependency directions between planes
- Read [references/results.md](references/results.md) - Load when authoring or interpreting a lifecycle result or `PhaseResultHandle`
- Read [references/providers.md](references/providers.md) - Load when defining or validating result-provider resolve, read, publish, revise, relate, version, capability, or restart behavior
- Read [references/profiles.md](references/profiles.md) - Load when composing topology, variation axes, exit rules, cumulative completion, or cross-plane requirements
- Read [references/development-contracts.md](references/development-contracts.md) - Load when selecting a development executor, assigning parallel work, isolating workspaces, retrying partial failures, or distinguishing consolidation from Integration
- Read [references/develop-run.md](references/develop-run.md) - Load when executing one or more exact Planning assignments as independently publishable Development results
- Read [references/develop-consolidate.md](references/develop-consolidate.md) - Load when consolidating exact Development revisions into a development workspace
- Read [references/develop-abandon.md](references/develop-abandon.md) - Load when stopping a Development assignment and preserving its partial result and reason
- Read [references/deliver-publish.md](references/deliver-publish.md) - Load when publishing a local, hosted, or non-file reviewable candidate through a selected provider
- Read [references/assurance-contracts.md](references/assurance-contracts.md) - Load when composing inventory, review, QA, assessment, scanners, CI, human evidence, independence, or freshness rules
- Read [references/inventory-generate.md](references/inventory-generate.md) - Load when generating an SBOM, AIBOM/AI-SBOM, ML-BOM, CBOM, service inventory, or agent-environment inventory
- Read [references/assessment-run.md](references/assessment-run.md) - Load when assessing any exact workflow result against a pinned framework, policy, risk, security, accessibility, AI, or supply-chain criterion set
- Read [references/review-run.md](references/review-run.md) - Load when reviewing an exact Deliverable Publication revision and publishing independent findings
- Read [references/qa-run.md](references/qa-run.md) - Load when validating an exact Deliverable Publication revision in one or more recorded environments
- Read [references/operational-contracts.md](references/operational-contracts.md) - Load when binding authorization, privileged operations, exception or break-glass behavior, agent assistance, or operational failure policy
- Read [references/acceptance-validate.md](references/acceptance-validate.md) - Load when assembling fresh evidence for a separate accountable Acceptance decision
- Read [references/acceptance-decide.md](references/acceptance-decide.md) - Load when an accountable human accepts, conditionally accepts, or rejects an exact Deliverable revision
- Read [references/integration-validate.md](references/integration-validate.md) - Load when preflighting exact-revision Integration authorization and external-enforcement capabilities without mutation
- Read [references/integration-run.md](references/integration-run.md) - Load when executing or rolling back an authorized Integration through a protected target-side capability
- Read [references/transition-run.md](references/transition-run.md) - Load when planning, executing, verifying, or rolling back an operational Transition
- Read [references/release-run.md](references/release-run.md) - Load when executing, verifying, rolling back, or recovering a controlled Release
- Read [references/observe-run.md](references/observe-run.md) - Load when collecting and publishing operational Observation evidence for an exact subject and time window
- Read [references/incident-run.md](references/incident-run.md) - Load when declaring, updating, containing, recovering, escalating, reporting, or closing an Incident
- Read [references/corrective-action-run.md](references/corrective-action-run.md) - Load when planning, executing, verifying, or closing a Corrective Action
- Read [references/retirement-run.md](references/retirement-run.md) - Load when planning, executing, verifying, or rolling back Retirement of lifecycle resources
- Read [references/inspect.md](references/inspect.md) - Load when inspecting an effective workflow from native references and environment facts
- Read [references/conformance.md](references/conformance.md) - Load when validating a result, handle, provider handoff, or workflow profile
