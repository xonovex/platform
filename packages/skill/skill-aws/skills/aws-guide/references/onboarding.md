# Safe Onboarding, Break-Glass, and Operations

## Lifecycle

1. **Discover** partition/org/account/OU/region, effective actor/policy layers, identity providers/roles, guardrails, audit/config/security services, costs, data destinations, and evidence sinks.
2. **Assess** requested actions, escalation paths, existing controls, missing temporary identity, unsupported regions/services, evidence gaps, blast radius, and rollback/break-glass readiness.
3. **Propose** the smallest temporary-identity, least-privilege, guardrail, audit, configuration, and security composition with native evidence.
4. **Preview** exact policy/trust/attachment/service changes, permissions, credential/network/data/cost effects, rollout, failure/partial-state behavior, verification, break-glass, rollback, and drift.
5. **Authorize** the preview digest and exact subjects/resources through appropriate segregation of duties.
6. **Apply** idempotently against observed policy/configuration versions and stop on drift or authority expansion.
7. **Verify** authoritative state, policy simulation plus safe allow/deny probes, temporary expiry, region/account coverage, evidence resolution, outage behavior, and rollback availability.
8. **Record** separate preview, authorization, apply, policy decision, target mutation, audit/config/finding, verification, exception, and rollback references.
9. **Operate** diagnose, dry-run, rotate, update, detect drift, review exceptions/findings, disable, remove, and roll back.

## Break-glass

Define owner, eligible actors, trigger, scope, maximum duration, authentication, approval, session restrictions, compensating controls, monitoring, notification, evidence, automatic expiry/revocation, and post-use review. Test it before an incident without leaving standing broad credentials.

Break-glass cannot silently bypass organization policy. Its use is an explicit, expiring exception with independently retained CloudTrail and authorization evidence.

Removal deletes only owned trust, roles/policies, attachments, rules/configuration, and integrations. Preserve foreign resources and retained evidence. Preview ordered compensation where rollback is not atomic.
