# Workflow Conformance

## Validation order

1. Identify the claimed result, handle, result-provider capability, or profile contract.
2. Validate semantic requirements from [results.md](results.md) or [profiles.md](profiles.md).
3. Resolve native references and revisions through providers where available.
4. Check exact-revision binding, evidence origin, freshness, actor authority, and supported follow-up capabilities.
5. Check that representation remains native and the runtime handle remains ephemeral.
6. Check cross-plane requirements against declared governance guarantees without requiring governance installation for workflow-only profiles.
7. Return pass/fail per rule with evidence and remediation.

For a result provider, run every operation in [providers.md](providers.md), restart the adapter, resolve the published native reference again, and compare semantic state and revision. Exercise stale revisions, unavailable capabilities, duplicate publish/revise requests, relationship reconstruction, authorization failure, and partial provider failure without assuming a file representation.

## Required failures

Fail conformance when any of these is true:

- semantic result content is missing or a composite erased a constituent result;
- a persisted universal result envelope, central workflow identity, or runtime trace identity is required;
- an explicitly selected provider is unavailable and a side effect silently falls back elsewhere;
- provider context, native reference, source relationships, or reconstruction capability is insufficient for a handoff;
- exact-revision evidence is stale or the relevant subject changed;
- profile topology has invalid edges or missing prerequisites;
- a mandatory cross-plane requirement has no adequate enforcement guarantee;
- YAML, JSON, files, Git, tickets, or one provider/harness are treated as universal requirements.

## Whole composition

Validate a full assembled selection — a workflow profile, the governance profile it pairs with, and the capabilities, providers, modules, and methods selected under them — as one artifact, so a catalog cannot be assembled inconsistent or incomplete. `validateAssembledComposition` (`scripts/composition-helpers.ts`) composes the per-plane `validateProfile` and `validateComposition` and the capability registry rather than re-deriving them, and returns the first failure. A whole composition fails when:

- the adoption mode's expected-absence report is missing (`absence-report-missing`);
- either plane's profile fails its own contract;
- an integrated workflow profile names a governance profile that is not the one assembled (`dangling-governance-reference`);
- selected modules leave an unresolved conflict (`module-conflict`);
- a selected capability is neither shipped nor adopter-supplied in the capability registry (`dangling-capability`);
- a required capability is not selected (`missing-capability`);
- a selected provider is unavailable or incompatible (`incompatible-provider`);
- a mandatory cross-plane control has no supported, guaranteed enforcement point (`unenforced-mandatory-control`).

Run `node scripts/validate-assembled-composition.ts`; it assembles a shipped reference profile into a complete composition and rejects each adversarial fixture with its own code.

## Fixtures

Run `node scripts/validate-fixtures.mjs` from the guide directory or `npx moon run skill-workflow:test` from the repository root. The JSON fixture is test data for the abstract semantics; it is not a persisted workflow schema.

The suite covers all canonical result kinds, extended inventory components, ephemeral handles, composite preservation, explicit-provider failure, mandatory enforcement gaps, a self-controlled non-file task system, optional repository/GitHub/GitLab providers, restart reconstruction, and runtime trace identity rejection. Development and assurance fixtures additionally cover executor selection, isolated concurrency, consolidation versus Integration, abandonment, partial failure, retries, local/non-file and hosted publication, inventory specializations, generic Assessment, focused Review/QA, stale evidence, prompt injection, poisoned evidence, assessor independence, and external CI evidence identity. Operational fixtures cover evidence-only versus human Acceptance, exact authorization, target/evidence/policy drift, protected Integration, Transition rollback, controlled Release failure and recovery, Observation inputs, Incident escalation and applicability, Corrective Action effectiveness, exceptions including emergency ones, agent authority boundaries, data deletion, and Retirement verification. Reusable assertions live in `scripts/` for later harness and CI adapters.

## Scenario fixture corpus

`assets/fixtures/` holds adversarial given/when/expect scenarios — one JSON file per scenario with an `expect.must_not` prohibition list — spanning lifecycle authority, executor selection, privileged operations, provider round-trips, evidence freshness, and profile weakening. `assets/fixtures/index.json` maps every scenario to the reference file owning its contract (early-lifecycle scenarios resolve to plan-guide operations); `scripts/validate-conformance-scenario-fixtures.mjs` enforces schema, contract-to-owner agreement, owner existence, index bijection, and its own mutation guards. A conformance verification replays each scenario against the owning contract and reports any `must_not` behavior as a failure.
