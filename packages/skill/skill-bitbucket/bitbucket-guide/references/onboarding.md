# Safe Onboarding and Operations

## Lifecycle

1. **Discover** product/version, workspace/project/repository, plan, runners/apps, existing controls, bypass actors, identities/scopes, secrets, integrations, hooks, and evidence sinks.
2. **Assess** the request against the matching pinned baseline and identify unsupported parity assumptions, tier/app dependencies, conflicts, missing exact-revision binding, and credential/data risks.
3. **Propose** the smallest native composition with stable identities, immutable module refs, least privilege, provider-native evidence, probes, rollback, and drift owner.
4. **Preview** exact before/after mutations, scopes/claims/secrets, network/data/cost effects, failure/retry/partial-apply behavior, verification, and rollback.
5. **Authorize** the preview digest, product, subject, actor, scope, version, and expiry.
6. **Apply** idempotently against observed versions and stop on drift or authority expansion.
7. **Verify** authoritative state plus allow, deny/unsupported, bypass, outage, duplicate, evidence-resolution, and rollback probes.
8. **Record** separate preview, authorization, apply, verification, evidence, and rollback references.
9. **Operate** diagnose, dry-run, rotate, upgrade pins/apps, detect drift, disable, uninstall, and roll back.

## Drift and removal

Drift includes product/version/plan changes, installed-app changes, moving shared configurations/pipes, altered branch/merge/deployment controls, new bypass actors, renamed/spoofable statuses, runner/secret/OIDC claim changes, webhook destinations, API/auth changes, artifact retention, and stale evidence.

Adopt an existing compatible resource only after explicit ownership transfer. Removal deletes only owned pipeline imports, restrictions/checks, deployments, variables, webhooks, apps/configuration, and OIDC trust; it preserves foreign controls and retained native evidence.

Rollback restores captured native versions and re-runs both positive and negative probes. Preview compensating changes and partial-state risk when atomic rollback is unavailable.
