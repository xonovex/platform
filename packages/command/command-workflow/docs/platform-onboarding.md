# External Platform Onboarding

External controls and enterprise providers are optional owners. They expose native automation, authorization, evidence, configuration, rollback, and drift operations without becoming a universal workflow representation.

| Platform owner                                                                                      | Native mechanisms to prefer                                                                                                                                    | Capability boundaries to preserve                                                                                                                   |
| --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| [`github-guide`](../../../skill/skill-github/github-guide/references/automation-and-enforcement.md) | Full-SHA-pinned reusable workflows, composite actions, required checks, layered rulesets, protected environments, least-privilege tokens/OIDC, attestations    | Plan/feature availability, bypass actors, check identity/source, environment reviewer rules, and GitHub-hosted evidence                             |
| [`gitlab-guide`](../../../skill/skill-gitlab/gitlab-guide/references/automation-and-enforcement.md) | Versioned CI/CD components, typed inputs, pipeline execution policies, compliance frameworks, protected environments/deployment approvals                      | GitLab version/tier, policy merge and injection behavior, job conflicts, framework assignment limits, pipeline/job/artifact/deployment evidence     |
| [`azure-devops-guide`](../../../skill/skill-azure-devops/skills/azure-devops-guide/SKILL.md)        | Boards work items, Repos branch policies, reusable Pipelines templates, approvals/checks, artifacts, workload-identity service connections, REST/service hooks | Services versus Server version, organization/project scope, hosted/self-hosted agents, supported checks, identities, and provider-native references |
| [`bitbucket-guide`](../../../skill/skill-bitbucket/skills/bitbucket-guide/SKILL.md)                 | Cloud shared pipeline configuration, deployments, merge/custom checks, branch restrictions, OIDC, REST/webhooks; Data Center native plugins/REST/statuses      | Detect Cloud versus Data Center before selection; no inferred API, pipeline, permission, or deployment parity                                       |
| [`bitrise-guide`](../../../skill/skill-bitrise/skills/bitrise-guide/SKILL.md)                       | Versioned Workflows/Pipelines/Steps, triggers, secret scopes, artifacts, build status, API/webhooks, Bitrise-to-AWS OIDC                                       | Workspace/tier, hosted versus self-hosted runner, verified-step/source trust, artifact/status linkage, claim availability, and cost                 |
| [`aws-guide`](../../../skill/skill-aws/skills/aws-guide/SKILL.md)                                   | Workforce/workload federation, IAM roles/boundaries, Organizations/SCPs, CloudTrail, Config, Security Hub, provider-native break glass                         | Organization/account/region, effective policy layers, temporary credentials, trust claims, evidence delays, and service-specific authorization      |
| [`datadog-guide`](../../../skill/skill-datadog/skills/datadog-guide/SKILL.md)                       | CI/CD Visibility, DORA correlation, OpenTelemetry, LLM Observability, Audit Trail, Software Catalog, AWS integration, Cloud Security                           | Site/region/tier, source-of-truth boundaries, collection manifest, redaction, sampling, retention, residency, access, export, and cost              |

## Transactional recipe

For each platform, discover product, deployment model, edition/version, account/tier, scopes, identities, native configuration, enabled modules, current permissions, credentials, network/data flows, evidence, and feature availability. Missing access is an unknown, not an empty configuration.

Before mutation, preview:

- exact native resource and configuration changes;
- reusable module identity, immutable revision, dependencies, and requested permissions;
- credential issuance, issuer/audience/subject claims, expiry, secret access, and escalation paths;
- data fields/content, network destinations, retention/residency/access/export, volume, and cost;
- decision, enforcement, authorization, build/artifact/deployment, telemetry, and audit evidence;
- idempotency, partial failure, outage behavior, disable/removal, ordered rollback, and drift ownership.

Apply only after authorization, then re-read authoritative provider state and run safe positive, negative, unsupported-capability, duplicate/retry, outage, rollback, and evidence-resolution probes. Preserve foreign resources and retained evidence.

## Independent and mixed adoption

GitHub or GitLab external enforcement can be installed without agent hooks or lifecycle commands. Azure Boards can own work items while Bitbucket owns source, Bitrise owns mobile CI, AWS owns runtime governance, and Datadog owns observation. Correlate exact native references; do not coerce them into one provider identity.

Prefer temporary workload federation for CI-to-AWS access. Do not generate long-lived static access keys by default. A provider framework, conformance pack, finding, dashboard, or crosswalk is evidence within its stated scope, not automatic certification or legal compliance.
