---
name: datadog-guide
description: "Use when operating or onboarding Datadog CI/CD Visibility, DORA metrics, OpenTelemetry, LLM Observability, Audit Trail, Software Catalog, AWS integration, Cloud Security, or provider-native telemetry evidence. Triggers on pipeline/deployment correlation, trace and LLM content controls, service ownership, cloud findings, API/app keys, redaction, sampling, retention, residency, access, cost, rollback, drift, or privacy-preserving observability — even when the user doesn't say 'Datadog'."
---

# Datadog Observability and Evidence

Operate Datadog as an optional evidence and telemetry provider. Datadog correlates CI, deployment, runtime, cloud, security, and LLM signals but does not become the authoritative workflow, source, build, or cloud identity.

## Essentials

- **Discover product and data scope** — resolve site/region, organization, tier, enabled products, integrations, API/app-key roles, pipelines/services, retention, residency, access, sampling, and cost before collection.
- **Keep sources authoritative** — preserve native source-provider and Datadog references; a Datadog trace/event/finding correlation never replaces the original commit, build, deployment, AWS, or audit record.
- **Minimize before ingesting** — metadata is the default; prompts, model/tool content, source, logs, secrets, personal data, and payload bodies require an explicit purpose and authorization.
- **Govern the full data lifecycle** — preview fields, tags, redaction, sampling, routing, retention, residency, access, deletion, downstream export, volume, and cost.
- **Separate observation from enforcement** — monitors and product controls may support policy, but telemetry presence alone is not an unbypassable release or cloud authorization gate.
- **Transact onboarding** — discover, preview exact integrations/configuration and data flow, authorize, apply idempotently, re-read and probe, roll back, and detect drift.

## Workflow

1. Identify source systems, exact revisions/deployments/resources, Datadog site/org/products/tier, identities, integration paths, and data classifications.
2. Define the minimum correlation fields and independent opaque references; do not promote trace/session IDs into workflow identity.
3. Preview configuration, keys/roles, network endpoints, collected fields/content, redaction/sampling, retention/residency/access/deletion, volume/cost, evidence, failure behavior, verification, rollback, and drift.
4. Require authorization before mutation or sensitive-content collection.
5. Apply against observed versions, re-read effective configuration, run data-present/data-absent/privacy/outage probes, and return native evidence references with limitations.

## Gotchas

- CI Visibility, CD Visibility, DORA, Audit Trail, Software Catalog, Cloud Security, and LLM Observability have different sources, tiers, identities, retention, and semantics.
- A service, environment, version, commit SHA, pipeline ID, and deployment ID must map consistently; tag similarity alone can mis-correlate unrelated evidence.
- Prompt and completion capture can contain secrets, personal data, source, customer content, and regulated data. Default it off unless explicitly justified and governed.
- API keys identify the organization; application keys and roles authorize operations. Never place either in source, previews, logs, traces, fixtures, or client-side applications.
- Sampling and retention reduce volume but can remove evidence. Record the policy and state when an absent event is inconclusive.
- DORA metrics depend on event mapping and completeness; they are measurements with limitations, not proof of delivery quality.

## Example

```text
Detected: Datadog EU site · CI/CD + OTel enabled · LLM content capture disabled
Preview: collect pipeline/deployment metadata and opaque Bitrise/AWS refs;
         redact secrets/user IDs; 10% traces; 30-day retention; EU residency; bounded cost
Verify: expected correlation resolves; prompt/source/secret canaries are absent;
        unauthorized role and outage paths fail visibly
Rollback: disable owned integrations/export and preserve authorized retained audit evidence
```

## Progressive Disclosure

- Read [references/ci-cd-and-dora.md](references/ci-cd-and-dora.md) - Load when instrumenting pipelines, tests, deployments, services, DORA events, exact revisions, or delivery correlations
- Read [references/opentelemetry-and-llm.md](references/opentelemetry-and-llm.md) - Load when configuring OTel ingestion, GenAI/LLM traces, evaluations, prompt/content capture, attributes, sampling, or redaction
- Read [references/audit-catalog-aws-and-cloud-security.md](references/audit-catalog-aws-and-cloud-security.md) - Load when operating Audit Trail, Software Catalog, AWS integration, Cloud Security, ownership, or native cloud/security evidence
- Read [references/privacy-and-onboarding.md](references/privacy-and-onboarding.md) - Load when setting up, diagnosing, dry-running, authorizing data collection, managing keys/roles, retention/residency/access/cost, rolling back, or checking drift
- Read [references/provider-conformance.md](references/provider-conformance.md) - Load when testing event mapping, telemetry redaction, source references, product tiers, outages, rate limits, rollback, or fresh-context recovery
