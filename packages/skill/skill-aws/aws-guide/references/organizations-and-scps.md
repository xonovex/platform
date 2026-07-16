# Organizations and Service Control Policies

## Model

Record organization, root, OU/account targets, policy type/status, attachments at every ancestor, delegated administrators, management-account behavior, organization feature set, regions/services involved, and the actor authorized to change guardrails.

An SCP limits the maximum available permissions for affected member accounts; it does not grant permissions. Effective access still needs identity/resource policy authorization and remains subject to boundaries, session policy, explicit denies, and service-specific rules.

## Change procedure

1. Inventory current root/OU/account hierarchy, inherited SCPs, exemptions, critical services/roles, break-glass path, and recent authorization evidence.
2. Write the smallest guardrail and test policy syntax, service/action/resource/condition coverage, and interactions with all inherited policies.
3. Preview attachments/detachments, effective before/after maximum permissions, affected accounts/workloads, delegated administration, rollout order, monitoring, rollback, and irreversibility.
4. Authorize through segregation of duties; start with a test OU/account and progressive rollout.
5. Probe allowed critical paths and denied target behavior in each affected class; monitor CloudTrail and service health.
6. Roll back immediately on unintended denial, then re-read hierarchy and effective attachments.

Do not attach a broad deny at the root without a tested break-glass and progressive rollout. Do not use `FullAWSAccess` attachment as evidence that workloads have permissions; it only avoids an SCP-level restriction.

Drift includes hierarchy moves, new accounts/OUs, attachment changes, policy edits, delegated-admin changes, newly used regions/services, and break-glass use.
