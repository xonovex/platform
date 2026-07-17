# Sources

## IAM security best practices

- **URL:** https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/iam-and-federation.md`, `references/onboarding.md`, `references/provider-conformance.md`
- **Aspects extracted:** Federation, temporary credentials, least privilege, Access Analyzer, permissions review, and credential protection.

## IAM OIDC identity providers

- **URL:** https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/iam-and-federation.md`, `references/provider-conformance.md`
- **Aspects extracted:** OIDC identity providers, role trust, web-identity role assumption, and temporary credentials. Provider-claim minimization and no-static-key defaults are Xonovex security constraints.

## Organizations service control policies

- **URL:** https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/organizations-and-scps.md`, `references/provider-conformance.md`
- **Aspects extracted:** SCP maximum-permission semantics, inheritance, attachment, management-account limitations, and testing guidance.

## AWS CloudTrail

- **URL:** https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-user-guide.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/audit-config-and-security.md`, `references/onboarding.md`
- **Aspects extracted:** Trails, Event history, Lake event data stores, events, regions, selectors, retention, validation, and native audit identity.

## AWS Config conformance packs

- **URL:** https://docs.aws.amazon.com/config/latest/developerguide/conformance-packs.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/audit-config-and-security.md`, `references/provider-conformance.md`
- **Aspects extracted:** Pack templates, parameters, rules, organization deployment, evaluations, and remediation guidance.

## AWS Security Hub CSPM

- **URL:** https://docs.aws.amazon.com/securityhub/latest/userguide/what-is-securityhub.html
- **Last reviewed:** 2026-07-16
- **Used for:** `references/audit-config-and-security.md`, `references/provider-conformance.md`
- **Aspects extracted:** Findings, standards/controls, regions, aggregation, integrations, workflow, and provider-native evidence.

## Refresh Workflow

1. Re-check IAM/OIDC, Organizations/SCP, CloudTrail, Config, and Security Hub service documentation and regional/partition behavior.
2. Re-run policy-layer, OIDC, temporary-expiry, organization rollout, emergency-exception, evidence, outage, rollback, and drift probes.
3. Keep documentation conformance separate from live account evidence and update **Last reviewed**.
