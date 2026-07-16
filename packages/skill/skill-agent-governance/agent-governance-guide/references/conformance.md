# Governance Conformance

Validate each selected contract independently, then validate the effective composition.

## Checks

- Executor: permitted class, deterministic-first rationale, inputs/outputs, bounds, authority, validation, evidence, failure, cancellation.
- Event adapter: version, support state, handler execution, blocking, output/context, ordering/concurrency, configuration, limitations, trust.
- Policy: applicability, version, contextual facts, outcome, reasons, evidence requirements, freshness.
- Policy provider: deterministic equivalence, explanations, evidence requests, exceptions, versioning, historical replay, and opaque native decisions.
- Configuration provider: inspect, diff, preview, authorized idempotent apply, authoritative verify, rollback, export/import, partial application, and drift.
- Evidence/telemetry provider: provider-native references, correlation, minimization, authorized content capture, redaction, retention, residency, and access.
- Enforcement: native event/action, adequate guarantee, explicit failure, separate decision/evidence references.
- Module: identity, classification, adoption modes, authority zones, provenance, compatibility, permissions, tools, filesystem, network, secrets, data flows, side effects, timeout, retry, ordering, concurrency, idempotency, reentrancy, failure, ownership, support, upgrade, disable, and rollback.
- Profile: additive strengthening, non-weakening, adequate enforcement, actor rules, data/telemetry/distribution requirements.
- Exception/break-glass: owner, authority, scope, rationale, compensating controls, expiry, evidence, review.
- Onboarding: exact preview, authorization, idempotent apply, verification, rollback, drift.

## Required failures

Reject unsupported or experimental mandatory hooks, unbounded model/agent execution, unresolved module conflicts, assumed serial hooks, stale or unreplayable mandatory policy, expired exceptions, unauthorized weakening, missing enforcement guarantees, runtime trace identity, universal file/config/provider requirements, tampered or moving modules, unreviewed trust, permission expansion, non-idempotent retries, unsafe duplicate execution, unauthorized sensitive-content capture, and failed required rollback.

Run `node scripts/validate-fixtures.mjs` from the guide directory or `npx moon run skill-agent-governance:test` from the repository root. Fixture JSON is deterministic test input for the abstract contract, not a required module or policy manifest. `scripts/conformance-helpers.mjs` exports the provider and module checks for harness and CI adapters.

## Scenario fixture corpus

`assets/fixtures/` holds adversarial given/when/expect scenarios — one JSON file per scenario with an `expect.must_not` prohibition list — spanning adapter semantics, module contracts, trust, exceptions, external enforcement, provider policy and configuration, data governance, and the enterprise platforms. `assets/fixtures/index.json` maps every scenario to the reference file owning its contract; `scripts/validate-conformance-scenario-fixtures.mjs` enforces schema, contract-to-owner agreement, owner existence, index bijection, and its own mutation guards. A conformance verification replays each scenario against the owning contract and reports any `must_not` behavior as a failure.
