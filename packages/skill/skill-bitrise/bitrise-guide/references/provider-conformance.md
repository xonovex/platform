# Provider Conformance

Run conformance for every supported workspace plan, runner type, and stack/configuration baseline.

## Required cases

- Workspace/app/plan/configuration/stack/runner detection; unsupported/tier-restricted and stale baselines fail clearly.
- Workflow/Pipeline dependency execution, exact Step versions/provenance, timeouts, retries, concurrency, and effective build configuration.
- Push/tag/pull-request/target/manual/API trigger precedence, duplicate event reconciliation, recursive trigger prevention, and privileged near-miss denial.
- Protected secret behavior for trusted and fork/untrusted pull requests; self-hosted runner isolation, cleanup, cache, network, and credential boundaries.
- Build/artifact/status/deployment linkage to exact source commit, retention/access/redaction, missing/spoofed/stale/failed status, and fresh-context resolution.
- API pagination/rate limits/auth failure, webhook replay/duplicates/order, provider outage, partial apply, rollback, uninstall ownership, and drift.
- AWS OIDC constrained claims, temporary expiry, least privilege, cross-app/environment/action denial, no long-lived access keys, CloudTrail/build evidence, and trust removal.

Fixtures contain dated capability facts and opaque example identifiers without secret values or provider payload copies. Documentation conformance and live results remain separate.

The repository-wide enterprise fixture validates shared onboarding, federation, failure, and mixed-stack semantics; live Bitrise probes supply native references.
