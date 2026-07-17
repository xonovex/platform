# Safe Onboarding and Operations

## Lifecycle

1. **Discover** host/product (github.com or Enterprise Server), version/plan features, organization/repository scope, Actions and pinning policy, workflows and resolved `uses:` refs, rulesets/branch protections, required checks and their source apps, bypass actors, environments, runners, CODEOWNERS/required reviewers, App/PAT/`GITHUB_TOKEN` permissions, Actions/Dependabot secrets, OIDC trust, webhooks, deploy keys, and evidence sinks — from an authenticated read ([first-time-setup.md](first-time-setup.md) covers a fresh `gh` install and login).
2. **Assess** requested capabilities against the pinned baseline; report plan/feature gaps, ruleset and protection overlaps, bypass actors, renamed or spoofable check identities, missing full-SHA pins, untrusted-runner and fork exposure, and token/secret/data risks.
3. **Propose** the smallest native composition with SHA-pinned reusable workflows/actions, stable check identity, active layered rulesets, a protected environment for privileged jobs, least-privilege permissions and constrained OIDC claims, provider-native evidence, verification probes, rollback, and drift owner.
4. **Preview** exact before/after workflow, ruleset/protection, environment, permission, secret/OIDC, webhook, deploy-key, and App-installation changes plus network/data/cost effects, partial failure behavior, verification, and rollback.
5. **Authorize** the preview digest, host, organization/repository subject, actor, scope, module pin, and expiry.
6. **Apply** idempotently against observed resource revisions and the exact head SHA; stop on drift or permission expansion.
7. **Verify** authoritative ruleset/protection, environment, and permission state plus allow, deny (omitted, renamed, skipped, or spoofed check), direct/force-push, fork-origin, merge-queue, bypass-actor, API-merge, attestation-resolution, and credential-expiry probes.
8. **Record** separate preview, authorization, apply, verification, evidence, and rollback references.
9. **Operate** diagnose, dry-run, rotate tokens/keys, update pins, detect drift, disable, uninstall, and roll back.

## Drift and removal

Drift includes host/product/plan or Enterprise Server version changes, Actions allow/pinning policy changes, moved or unpinned action and reusable-workflow refs, altered workflow files, ruleset/protection status and bypass-actor changes, renamed checks or a changed source application, CODEOWNERS and reviewer-eligibility changes, App/PAT/`GITHUB_TOKEN` permission changes, secret and OIDC claim changes, new deploy keys, webhook destinations, artifact/log retention, and stale audit-log or attestation evidence.

Adopt an existing compatible resource only after explicit ownership transfer. Removal deletes only owned workflow callers, rulesets/protections, environments, secrets/variables, webhooks, deploy keys, App installations, and cloud OIDC trust; it preserves foreign controls, workflows, identities, and retained native evidence.

Rollback restores captured workflow, ruleset, environment, and pin revisions in dependency order, revokes temporary trust and grants, and re-runs both positive and negative probes. Preview compensating changes and partial-state risk when atomic rollback is unavailable.
