# Provider Conformance

Run Cloud and each supported Data Center version as separate suites.

## Cloud cases

- Workspace/plan/API/runner detection and tier-restricted or unknown capability outcomes.
- Repository/commit/pull-request reads/writes, native reference reconstruction after restart, conditional/duplicate writes, and exact-revision statuses.
- Branch permissions, merge/custom checks, bypass inventory, omitted/renamed/spoofed/stale/failed status and callback failures.
- Versioned shared pipeline configuration, artifact/deployment/status linkage, secret exposure protection, hosted/self-hosted runner boundary, and provider-native evidence.
- OIDC issuer/audience/subject constraints, least-privilege AWS role, temporary expiry, out-of-scope denial, and no generated access key.
- REST pagination/rate limits, webhook authentication/replay/duplicate/order, outage, partial apply, rollback, and drift.

## Data Center cases

- Deployment/version/API and installed-app detection; Cloud-only operations fail unsupported.
- Repository/pull-request version conflicts, permissions and bypass paths, build/deployment status, webhooks/hooks, and restart reconstruction.
- Installed-app provenance/permissions/compatibility, clustered behavior, upgrade transition, disable/rollback, outage, and failed mandatory control.

Fixtures contain dated product/version facts and opaque example identifiers without credentials or copied provider payloads. Mark documentation-only cases `documentation-conformant`; a missing live environment is `not-probed`, never passed.

The repository-wide enterprise fixture validates the shared capability, onboarding, failure, OIDC, and mixed-stack contract. Live suites add provider-native evidence references.
