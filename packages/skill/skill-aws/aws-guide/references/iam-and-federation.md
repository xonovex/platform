# IAM and Federation

## Effective authority

Discover partition, account, caller ARN and session context, organizations membership, identity policies, group/role attachments, permissions boundary, session policy, resource policies, role trust, SCP/RCP constraints where applicable, explicit denies, service control conditions, and target resource tags/policies.

Evaluate an explicit subject, action, resource, context, source identity, requested session, and policy versions. Preserve the policy/decision evidence separately from the eventual API mutation and CloudTrail record.

## Federation default

Prefer an existing identity center, workload identity, OIDC/SAML provider, or cross-account role that issues temporary credentials. For OIDC preview exact issuer, audience, subject and provider-specific repository/project/workspace/workflow/environment claims; role session duration/tags/source identity; trust conditions; permissions boundary; role/session policies; and target resource policies.

Never create an IAM user or access key by default for CI or agent workloads. When federation is unsupported, return the limitation and require a separately authorized alternative with owner, storage, rotation, expiry, revocation, detection, and migration back to temporary credentials.

## Least privilege procedure

1. Start from observed API actions/resources and required conditions, not a broad managed policy.
2. Use Access Analyzer and service authorization references to refine actions/resources/conditions.
3. Apply an appropriate boundary and short session; separate read-only discovery from configuration/deployment roles.
4. Simulate where supported, then run a safe allowed probe and representative denied actions/resources/claim variants.
5. Verify temporary expiry and resolve STS, CloudTrail, target-state, and policy references without exposing tokens.

Re-preview on any issuer/claim, trust, role policy, boundary, session, resource policy, or organization guardrail drift.
