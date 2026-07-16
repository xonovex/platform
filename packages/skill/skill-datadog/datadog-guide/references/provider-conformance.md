# Provider Conformance

Run conformance per Datadog site/region, organization, tier, product set, integration path, and source-provider version.

## Required cases

- Site/org/tier/product/key-role/integration discovery; unsupported/tier-restricted and stale claims fail clearly.
- Exact repository/commit/pipeline/job/artifact/deployment/service/environment/version mappings; duplicate/retry/cancel/skip/failure/rollback and incomplete DORA event cases.
- OTel semantic/config version, allowed attributes, high-cardinality limits, sampling, exporter queue/backpressure, outage/rate limit, duplicates, clock skew, and source-reference reconstruction.
- LLM metadata-only default; prompt/completion/tool/source/secret/personal-data canaries; pre-egress redaction; authorized content, sampling, retention, residency, access, deletion, and cost.
- Audit event scope/access/retention, catalog source-of-truth conflicts, AWS role/account/region/resource scope, Cloud Security coverage/finding workflow, and native references.
- Idempotent setup, permission expansion denial, partial apply, rollback/removal ownership, deliberate drift, unauthorized access, provider outage, and fresh-context recovery.

Fixtures contain dated product/config facts and opaque example IDs without live secrets or copied sensitive payloads. Mark documentation-only cases separately from live product probes. Sampling or retention limitations remain visible in every evidence result.

The repository-wide enterprise fixture validates shared onboarding, minimized telemetry, failure, and mixed-stack semantics; live Datadog probes add pipeline, deployment, trace, audit, catalog, integration, and finding references.
