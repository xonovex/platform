# Safe Onboarding and Operations

## Lifecycle

1. **Discover** offering (GitLab.com, Self-Managed, or Dedicated), edition/tier/version and feature flags, group/project scope, includes and resolved component refs, policy projects and linked scopes, compliance frameworks, protected branches/tags/environments, push rules, approval rules and required approvals, CODEOWNERS, runners, roles, project/group access tokens and `CI_JOB_TOKEN` allowlists, masked/protected CI/CD variables, `id_token` cloud trust, webhooks, deploy keys/tokens, and audit-event sinks — from an authenticated read ([first-time-setup.md](first-time-setup.md) covers a fresh `glab` install and login).
2. **Assess** requested capabilities against the pinned baseline; report tier/flag gaps, approval-rule and job-name conflicts, bypass actors, missing full-SHA pins, untrusted-runner and fork exposure, unsupported features, and token/variable/data risks.
3. **Propose** the smallest native composition with SHA-pinned typed components, mandatory pipeline-execution-policy injection, stable job names, protected targets, least-privilege identities and constrained `id_token` claims, provider-native evidence, verification probes, rollback, and drift owner ([automation-and-enforcement.md](automation-and-enforcement.md) owns which mechanism realizes each intent).
4. **Preview** exact before/after component, policy, framework, protected-branch/tag/environment, role, token/variable, `id_token`, webhook, and deploy-key changes plus the resolved component/policy SHAs, the merged job graph, network/data/cost effects, partial failure behavior, verification, and rollback.
5. **Authorize** the preview digest, offering, group/project subject, actor, scope, component/policy pin, and expiry.
6. **Apply** idempotently against observed revisions and the exact head SHA; stop on drift or authority expansion.
7. **Verify** authoritative policy, framework, protected-setting, and approval-rule state plus compliant, omission, duplicate-job-name, variable-override, bypass-role, fork/detached-pipeline secret denial, protected-environment, policy-outage, evidence-resolution, and credential-expiry probes.
8. **Record** separate preview, authorization, apply, verification, evidence, and rollback references.
9. **Operate** diagnose, dry-run, rotate tokens/keys, update pins, detect drift, disable, uninstall, and roll back.

## Drift and rollback

Drift includes offering/tier or Self-Managed version and feature-flag changes, moved or unpinned component and include refs, unlinked policy projects, altered policy merge strategy or suffix behavior, renamed or duplicated policy job names, framework reassignment, weakened protected branches/tags/environments, push-rule and required-approval changes, new bypass actors, CODEOWNERS and approver-eligibility changes, runner trust, masked/protected variable and `id_token` claim changes, new project/group access tokens or deploy keys/tokens, widened `CI_JOB_TOKEN` allowlists, webhook destinations, artifact/log retention, and stale audit-event evidence.

Rollback restores captured component pins, policy links and configuration, framework assignment, protected settings, approval rules, variables, and credentials in dependency order, revokes temporary trust and grants, and re-runs both positive and negative probes. Preview compensating changes and partial-state risk when atomic rollback is unavailable.

Adopt an existing compatible component, policy, framework, or protected setting only after explicit ownership transfer. Uninstall deletes only owned includes, policy links, framework assignments, protected settings, variables, webhooks, deploy keys, and cloud `id_token` trust; it preserves foreign policies, pipelines, approval rules, identities, and retained audit events.
