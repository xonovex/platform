# Workflow Conformance

## Validation order

1. Identify the claimed result, handle, result-provider capability, or workflow profile.
2. Validate semantic requirements from [results.md](results.md) or [profiles.md](profiles.md).
3. Resolve native references and revisions through the selected provider.
4. Check exact-revision binding, evidence origin, freshness, and supported follow-up capabilities.
5. Check that provider representation remains native and the runtime handle remains ephemeral.
6. Return pass/fail per rule with evidence and remediation.

For a result provider, exercise every operation in [providers.md](providers.md), restart the
adapter, resolve the published native reference again, and compare semantic state and
revision. Include stale revisions, unavailable capabilities, duplicate requests,
authorization failure, and partial provider failure.

## Required failures

Fail conformance when:

- semantic result content is missing or a composite erased a constituent result;
- a universal persisted envelope, central workflow identity, or runtime trace identity is required;
- an explicitly selected provider is unavailable and silently replaced;
- a native reference cannot be reconstructed after restart;
- exact-revision evidence is stale;
- profile topology has invalid edges or missing prerequisites;
- one file format, provider, harness, executor, or runtime composition is treated as universal.

Executable composition conformance is separate and owned by `agent-governance-guide`.
Do not assemble a workflow profile with a control profile.

## Fixtures

Run `node scripts/validate-fixtures.mjs` from the guide directory or
`npx moon run skill-workflow:test` from the repository root. JSON fixtures are test data
for abstract semantics, not persisted workflow schemas.

The adversarial corpus in `assets/fixtures/` uses one given/when/expect document per
scenario. `scripts/validate-conformance-scenario-fixtures.mjs` validates schema, owning
contract references, index coverage, and mutation guards.
