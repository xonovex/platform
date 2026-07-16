# Privacy-Preserving Onboarding and Operations

## Collection manifest

Before mutation, enumerate each source, field/content category, purpose, data owner, subject/resource scope, classification, redaction/tokenization, sampling, route/destination/site, residency, retention, access roles, deletion/export, downstream sharing, volume/cost budget, evidence need, and prohibited data. Unknown fields default to not collected.

Use read-only API/application credentials for discovery. Configuration credentials are separate and narrowly scoped. Store keys in an approved secret manager or CI secret, never in source, commands, previews, logs, telemetry, fixtures, or client applications.

## Lifecycle

1. **Discover** site/org/tier/products, integrations, keys/roles, pipelines/services, sources, current collection, redaction, sampling, retention, residency, access, exports, and costs.
2. **Assess** required evidence/correlation against data minimization, source-of-truth boundaries, sensitive content, product/tier gaps, volume/cost, and outage behavior.
3. **Propose** the smallest metadata-first integration with explicit collection manifest and native references.
4. **Preview** exact integration/configuration changes, keys/roles, endpoints, fields/content, redaction/sampling, retention/residency/access/deletion/export, cost, failure, verification, rollback, and drift.
5. **Authorize** the preview digest and every sensitive-content purpose/scope separately.
6. **Apply** idempotently against observed configuration and stop on drift, new fields, or expanded access.
7. **Verify** authoritative configuration plus expected-data, prohibited-data canary, correlation, role denial, sampling/retention, outage/rate/backpressure, deletion, and rollback probes.
8. **Record** separate preview, authorization, apply, verification, source-provider, Datadog evidence, privacy, cost, and rollback references.
9. **Operate** diagnose, dry-run, rotate, review access/volume/cost, update mappings, detect drift, disable, remove, and roll back.

Drift includes new products/sources/fields/tags, enabled content capture, weakened redaction, changed sampling/retention/residency/access/export, broadened keys/roles, changed site/endpoints, ownership mapping changes, new AWS accounts/regions, and cost anomalies.

Removal disables only owned integrations/export/collection and revokes owned keys/roles/trust. Preserve foreign monitors/catalog entries and legally/operationally retained evidence according to the authorized policy.
