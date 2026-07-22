# Sources

## API and application keys

- **URL:** https://docs.datadoghq.com/account_management/api-app-keys/
- **Additional URLs:** https://docs.datadoghq.com/api/latest/
- **Last reviewed:** 2026-07-22
- **Used for:** `SKILL.md`, `references/auth.md`, `references/privacy-and-onboarding.md`
- **Aspects extracted:** API, application, and client-token roles; application-key scopes and ownership; request headers; one-time-read behavior; validation endpoints; propagation delay; and revocation.

## Datadog CI/CD Visibility and DORA

- **URL:** https://docs.datadoghq.com/continuous_integration/
- **Additional URLs:** https://docs.datadoghq.com/continuous_delivery/ · https://docs.datadoghq.com/dora_metrics/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/ci-cd-and-dora.md`, `references/provider-conformance.md`
- **Aspects extracted:** CI pipelines/tests, deployments, service/version/environment correlation, and DORA metrics. Exact source binding and completeness limitations are Xonovex adapter constraints.

## OpenTelemetry and LLM Observability

- **URL:** https://docs.datadoghq.com/opentelemetry/
- **Additional URLs:** https://docs.datadoghq.com/llm_observability/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/opentelemetry-and-llm.md`, `references/provider-conformance.md`
- **Aspects extracted:** OTel ingestion/configuration and LLM/agent traces, evaluations, metrics, content, and costs.

## Audit Trail and Software Catalog

- **URL:** https://docs.datadoghq.com/account_management/audit_trail/
- **Additional URLs:** https://docs.datadoghq.com/internal_developer_portal/software_catalog/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/audit-catalog-aws-and-cloud-security.md`, `references/privacy-and-onboarding.md`
- **Aspects extracted:** Organization/configuration audit events, catalog service ownership/metadata, access, and native records.

## AWS integration and Cloud Security

- **URL:** https://docs.datadoghq.com/integrations/amazon-web-services/
- **Additional URLs:** https://docs.datadoghq.com/security/cloud_security_management/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/audit-catalog-aws-and-cloud-security.md`, `references/provider-conformance.md`
- **Aspects extracted:** AWS integration roles/resources/data and Cloud Security posture/runtime findings, supported resources, and operating modes.

## Privacy and data minimization

- **URL:** https://www.nist.gov/privacy-framework
- **Additional URLs:** https://eur-lex.europa.eu/eli/reg/2016/679/oj
- **Last reviewed:** 2026-07-16
- **Used for:** `references/opentelemetry-and-llm.md`, `references/privacy-and-onboarding.md`, `references/provider-conformance.md`
- **Aspects extracted:** Privacy risk management and principles for purpose limitation, data minimization, retention, access, and deletion. Applicability and legal interpretation require qualified review.

## Refresh Workflow

1. Re-check product/tier/site behavior for CI/CD, DORA, OTel, LLM, Audit Trail, Catalog, AWS, and Cloud Security.
2. Re-run exact-correlation, redaction/content, sampling/retention/residency/access/deletion/cost, role, outage, rollback, and drift probes.
3. Keep documentation conformance separate from live organization evidence and update **Last reviewed**.
