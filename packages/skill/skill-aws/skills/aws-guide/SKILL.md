---
name: aws-guide
description: "Use when designing, operating, or onboarding AWS identity, federation/OIDC, least privilege, permissions boundaries, Organizations/SCPs, CloudTrail, Config conformance packs, Security Hub, evidence, or break-glass access. Triggers on temporary CI credentials, role trust claims, cross-account access, organization guardrails, audit trails, configuration/security findings, provider-native references, rollback, drift, or privileged cloud mutation — even when the user doesn't say 'AWS governance'."
---

# AWS Governance and Evidence

Operate AWS as an optional runtime, authorization, governance, audit, configuration, and security-evidence provider. Keep CI-provider identity claims and observability-provider configuration in their owning platforms.

## Essentials

- **Discover authority first** — resolve partition, organization/account/OU, region, caller ARN/session, identity policies, boundaries, SCPs, resource policies, trust policies, and service configuration before evaluation or mutation.
- **Prefer temporary federation** — use workforce/workload federation, OIDC, IAM roles, and short sessions; never create long-lived access keys by default.
- **Evaluate all policy layers** — an allow is effective only within identity/resource policy, permissions boundary, session policy, SCP/RCP where applicable, explicit denies, and service-specific authorization.
- **Separate guardrails from grants** — SCPs set maximum permissions and do not grant access; role policies and resource policies provide authorized actions.
- **Preserve native evidence** — return opaque STS session, IAM policy, CloudTrail event, Config evaluation, Security Hub finding, and target-resource references with account/region/revision/freshness.
- **Transact privileged setup** — discover, simulate and preview exact changes, authorize, apply idempotently, re-read and probe, retain break-glass/rollback, and detect drift.

## Workflow

1. Establish partition/account/organization/region, effective actor, source workload identity, and every applicable policy layer.
2. Resolve the exact resource and immutable artifact/revision for the intended privileged action.
3. Preview policies/trust/guardrails/service configuration, permissions and escalation paths, credential/network/data/cost effects, evidence, failure behavior, verification, rollback, and drift.
4. Require authorization with segregation of duties and break-glass handling appropriate to impact.
5. Apply against observed versions, verify effective authorization and native service state, then return separate decision/mutation/audit/evidence references.

## Gotchas

- An SCP never grants permission. Testing only the SCP document misses identity/resource policies, boundaries, session policy, explicit denies, and management-account limitations.
- OIDC trust with a broad subject can grant every repository/project/workflow the role even when the permissions policy looks narrow.
- IAM policy simulation is evidence, not final proof for every service condition or resource policy; run bounded allowed and denied live probes where safe.
- CloudTrail Event history, trails, Lake event data stores, region coverage, data events, retention, and integrity are distinct configuration choices.
- Config conformance packs and Security Hub standards/findings are assessment evidence, not automatic legal or security compliance.
- Break-glass is a scoped, monitored, expiring exception with independent evidence; it is not a permanent admin role shared by automation.

## Example

```text
Detected: aws partition · organization o-123 · account 111122223333 · eu-west-1
Preview: add OIDC provider and deployment role; constrain issuer/audience/repo/environment;
         attach least-privilege policy + boundary; SCP unchanged; 30-minute sessions
Verify: intended deployment allowed; other repo/account/action denied; credentials expire;
        CloudTrail and target-state references resolve
Rollback: remove owned trust/grants after proving break-glass and evidence retention
```

## Progressive Disclosure

- Read [references/iam-and-federation.md](references/iam-and-federation.md) - Load when designing roles, OIDC/workload identity, least privilege, boundaries, session policies, cross-account access, or credential rotation
- Read [references/organizations-and-scps.md](references/organizations-and-scps.md) - Load when changing organization/OU/account guardrails, authoring SCPs, testing impact, or handling management/delegated administration
- Read [references/audit-config-and-security.md](references/audit-config-and-security.md) - Load when configuring CloudTrail, Config conformance packs, Security Hub, evidence publication, retention, regions, or finding workflows
- Read [references/onboarding.md](references/onboarding.md) - Load when setting up, diagnosing, dry-running, authorizing, verifying, rolling back, removing, checking drift, or designing break-glass
- Read [references/provider-conformance.md](references/provider-conformance.md) - Load when testing policy layers, OIDC claims, temporary credentials, organization guardrails, outages, evidence, rollback, or fresh-context recovery
