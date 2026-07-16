# Harness Onboarding Transactions

## Advisor input

Inspect authoritative state before recommending a change:

- product surface, executable version, edition, operating system, and workspace trust;
- selected governance profile and mandatory semantic intents;
- active organization, system, user, project, local, plugin, and session configuration;
- installed modules with source, version, digest, permissions, and lifecycle state;
- native diagnostics, configuration errors, duplicate handlers, conflicts, and drift;
- requested filesystem, command, network, secret, model, and data access.

Missing executables or unreadable managed state remain explicit unknowns. Never fabricate an empty configuration.

## Advisor report

Return a structured report before any mutation:

```text
subject and native scope
observed product/runtime version and matrix version
selected profile and capability matches
unsupported, experimental, partial, stale, or conflicting requirements
exact native files/settings/modules to change
before/after diff or create/delete preview
permissions and data-flow report
trust and provenance decision
expected evidence and diagnostic probes
idempotency key
rollback target and retained evidence
authorization required
```

The report distinguishes knowledge-only, advisory, enforcing, configuration-changing, and privileged modules. A skill or instruction-only installation never appears as enforcing.

## Transaction

1. **Discover** — inspect versions, scopes, trust, native files, installed modules, and diagnostics.
2. **Assess** — resolve every mandatory intent against the matching fresh capability matrix.
3. **Preview** — show an exact native diff, permission/data flow, verification, disable, rollback, and drift plan.
4. **Authorize** — capture explicit consent for the exact subject, version, scope, and authority request; managed authorization remains separate.
5. **Apply** — use the platform's native configuration or package mechanism with an idempotency key. Do not overwrite unrelated entries.
6. **Verify** — re-inspect authoritative state and run deterministic probes for mapping, blocking, context, concurrency, precedence, and evidence.
7. **Record** — retain the approved preview, native references, observed digest/version, probe results, and limitations.
8. **Operate** — dry-run, diagnose, update, disable, remove, roll back, and detect drift through the same adapter.

Dry-run stops after preview and validation. Apply without both preview and authorization fails. A partial apply triggers rollback or a visible degraded state; it never reports success. Retry requires the same idempotency key or a fresh preview.

## Disable, rollback, update, and drift

- **Disable** uses a native per-module switch when one exists; otherwise remove only the owned native entry while preserving the rollback artifact.
- **Rollback** restores the verified prior native state and then re-runs probes. Deleting a new file is not sufficient if another index or package registration still references it.
- **Update** resolves the new version and permissions as a fresh trust decision. Never update a moving source in place.
- **Drift** compares intended and observed source, version, digest, enabled state, configuration, handler semantics, permissions, data flow, evidence freshness, and runtime capability.

Organization-managed modules retain their provider-native source and provenance references. Project-provided executable hooks, extensions, plugins, or packages remain disabled until repository trust and executable review are established.

## Failure behavior

Fail visibly on unsupported mandatory intents, stale matrices, untrusted project executables, permission expansion, lower-authority weakening, unreviewed generated hooks, unresolved duplicate handlers, invalid output schemas, missing rollback targets, failed verification, or drift that changes a guarantee.
