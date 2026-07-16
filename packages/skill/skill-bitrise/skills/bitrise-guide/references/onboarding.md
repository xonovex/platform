# Safe Onboarding and Operations

## Lifecycle

1. **Discover** workspace/app/plan, configuration source, Pipelines/Workflows/Steps, triggers, stack/runners, Git integration, secrets, artifacts, statuses, API/webhooks, OIDC, and evidence sinks.
2. **Assess** required capabilities, source-host/cloud ownership, tier gaps, untrusted-code boundary, duplicate triggers, Step provenance, permissions, data, cost, and unsupported behavior.
3. **Propose** the smallest version-pinned composition with least privilege, protected secrets/deployments, native evidence, verification, rollback, and drift owner.
4. **Preview** exact configuration/API changes, requested identities/scopes/secrets/claims, network/data/cost effects, failure/retry/partial-apply behavior, verification, and rollback.
5. **Authorize** the preview digest, app/workspace, actor, scope, version, and expiry.
6. **Apply** idempotently against observed configuration and stop on drift or authority expansion.
7. **Verify** authoritative configuration/build state plus trigger allow/deny, fork-secret denial, exact-status/artifact linkage, outage/duplicate behavior, OIDC denial, and rollback.
8. **Record** separate preview, authorization, apply, verification, build/artifact/status/cloud evidence, and rollback references.
9. **Operate** diagnose, dry-run, rotate, update Steps/stacks, detect drift, disable, uninstall, and roll back.

Removal restores captured trigger/configuration versions and deletes only owned integrations, variables, webhooks, and AWS trust. Preserve foreign Workflows/Steps and retained native evidence. Preview compensating changes when atomic rollback is unavailable.
