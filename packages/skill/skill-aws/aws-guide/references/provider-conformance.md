# Provider Conformance

Run conformance per partition, organization/account class, region set, identity provider, and supported service configuration.

## Required cases

- Caller/account/org/OU/region and every policy-layer discovery; unknown or stale facts fail privileged authorization.
- OIDC issuer/audience/subject and provider-specific claim constraints, least-privilege role/boundary/session, temporary expiry, cross-repository/project/environment/account/action denial, and no generated access keys.
- Identity/resource policy, boundary, session policy, SCP and explicit-deny interactions; simulation plus safe live probes; privilege-escalation and confused-deputy cases.
- SCP progressive rollout, inherited attachment/effective-policy checks, management-account limitation, critical-service allow, target deny, emergency exceptions, rollback, and hierarchy drift.
- CloudTrail region/event coverage, retention/access/integrity, Config recorder/pack/rule states, Security Hub regions/standards/findings/workflow, and opaque native references.
- Authorization failure, throttling, bounded retry, provider outage, delayed evidence, duplicate event/finding, partial apply, rollback failure, deliberate drift, and fresh-context recovery.

Fixtures contain dated service/policy facts and opaque example ARNs/IDs without account secrets or copied sensitive event payloads. Documentation conformance remains distinct from live account evidence.

The repository-wide enterprise fixture validates shared onboarding, federation, telemetry relationship, failure, and mixed-stack semantics; live AWS probes add STS, policy, CloudTrail, Config, Security Hub, and target references.
