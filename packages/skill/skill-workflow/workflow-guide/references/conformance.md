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

## Fixtures

Run `node scripts/validate-fixtures.mjs` from the guide directory or `npx moon run skill-workflow:test` from the repository root. The JSON fixture is test data for the abstract semantics; it is not a persisted workflow schema.

The suite covers all canonical result kinds, extended inventory components, ephemeral handles, composite preservation, explicit-provider failure, mandatory enforcement gaps, a self-controlled non-file task system, optional repository/GitHub/GitLab providers, restart reconstruction, and runtime trace identity rejection. Reusable provider assertions live in `scripts/conformance-helpers.mjs` for later harness and CI adapters.
