# Safe Onboarding and Operations

## Lifecycle

1. **Discover** product/edition/version, API resources, scopes, existing controls, identities, extensions, tiers, protected resources, hooks, secrets, and evidence sinks.
2. **Assess** requested capabilities against the pinned baseline; report overlaps, conflicts, bypass actors, missing exact-revision binding, unsupported features, and data/credential risks.
3. **Propose** the smallest native composition with versioned templates, scoped identities, stable policy/status names, evidence, verification probes, rollback, and drift owner.
4. **Preview** exact before/after native changes, permission/credential/network/data/cost effects, partial failure behavior, verification, and rollback.
5. **Authorize** the preview digest, subject, actor, scope, version, and expiry.
6. **Apply** idempotently against observed revisions; stop on drift or permission expansion.
7. **Verify** authoritative state plus allow, deny/unsupported, outage, evidence-resolution, and credential-expiry probes.
8. **Record** separate preview, authorization, apply, verification, evidence, and rollback references.
9. **Operate** diagnose, dry-run, rotate, update pins, detect drift, disable, uninstall, and roll back.

## Drift and rollback

Drift includes Services/Server or API-version changes, tier/extension changes, altered templates, policies, bypass permissions, approvers/checks, pipeline grants, identity claims, role assignments, hooks, secret/data flows, artifact retention, and evidence freshness.

Rollback restores captured native revisions in dependency order, removes only owned resources, revokes temporary trust/grants, and re-runs both positive and negative probes. If an operation cannot roll back atomically, preview compensating steps and partial-state risk before authorization.

Coexist with existing controls by detecting and adopting compatible native resources only after explicit ownership transfer. Uninstall never deletes foreign policies, pipelines, identities, hooks, or retained evidence.
