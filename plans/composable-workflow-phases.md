---
type: plan
has_subplans: true
status: complete
completed_date: '2026-07-16'
updated: '2026-07-16'
dependencies:
  plans: []
  subplans:
  - workflow-governance-contracts-profiles-and-execution
  - provider-policy-and-module-conformance
  - agent-harness-adapters-and-onboarding
  - external-automation-ci-and-platform-enforcement
  - governance-onboarding-walking-skeleton
  - workflow-discovery-research-design-decision-and-planning
  - workflow-development-delivery-inventory-assessment-review-and-qa
  - workflow-acceptance-integration-transition-release-observation-incidents-and-retirement
  - governance-policy-learning-observability-trust-and-operations
  - documentation-migration-crosswalk-and-validation
  - enterprise-platform-skills-and-onboarding
proposed_subplans:
- workflow-governance-contracts-profiles-and-execution
- provider-policy-and-module-conformance
- agent-harness-adapters-and-onboarding
- external-automation-ci-and-platform-enforcement
- governance-onboarding-walking-skeleton
- workflow-discovery-research-design-decision-and-planning
- workflow-development-delivery-inventory-assessment-review-and-qa
- workflow-acceptance-integration-transition-release-observation-incidents-and-retirement
- governance-policy-learning-observability-trust-and-operations
- documentation-migration-crosswalk-and-validation
- enterprise-platform-skills-and-onboarding
parallel_groups:
- group: 1
  plans:
  - workflow-governance-contracts-profiles-and-execution
- group: 2
  plans:
  - provider-policy-and-module-conformance
- group: 3
  plans:
  - agent-harness-adapters-and-onboarding
  - external-automation-ci-and-platform-enforcement
  - enterprise-platform-skills-and-onboarding
- group: 4
  plans:
  - governance-onboarding-walking-skeleton
- group: 5
  plans:
  - workflow-discovery-research-design-decision-and-planning
  - workflow-development-delivery-inventory-assessment-review-and-qa
  - workflow-acceptance-integration-transition-release-observation-incidents-and-retirement
- group: 6
  plans:
  - governance-policy-learning-observability-trust-and-operations
- group: 7
  plans:
  - documentation-migration-crosswalk-and-validation
skills_to_consult:
- command-guide
- skill-guide
- orthogonal-pattern-guide
- microkernel-pattern-guide
- hexagonal-pattern-guide
- connascence-guide
- testing-guide
- git-guide
- reflect-guide
- instruction-guide
- versioning-guide
research_sources:
  retrieved: '2026-07-14'
  source_hierarchy:
  - binding-law-and-regulation
  - international-standards
  - government-frameworks
  - vendor-primary-documentation
  - industry-specifications-and-authoritative-guidance
  - research-and-practitioner-material
  documentation:
  - https://code.claude.com/docs/en/hooks-guide
  - https://learn.chatgpt.com/docs/hooks
  - https://learn.chatgpt.com/docs/plugins
  - https://learn.chatgpt.com/docs/skills
  - https://learn.chatgpt.com/docs/managed-configuration
  - https://docs.github.com/en/copilot/concepts/agents/hooks
  - https://kiro.dev/docs/hooks/
  - https://opencode.ai/docs/plugins/
  - https://pi.dev/docs/packages
  - https://agentskills.io/specification
  - https://modelcontextprotocol.io/specification/2025-06-18
  - https://opentelemetry.io/docs/specs/semconv/gen-ai/
  - https://www.openpolicyagent.org/docs
  - https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
  - https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets
  - https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments
  - https://docs.github.com/en/actions/reference/security/secure-use
  - https://docs.gitlab.com/ci/components/
  - https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/
  - https://docs.gitlab.com/user/compliance/compliance_frameworks/
  - https://www.iso.org/standard/90219.html
  - https://www.iso.org/standard/81702.html
  - https://www.iso.org/standard/81118.html
  - https://www.iso.org/standard/81230.html
  - https://www.iso.org/standard/77304.html
  - https://www.iso.org/standard/77520.html
  - https://www.w3.org/TR/WCAG22/
  - https://eur-lex.europa.eu/eli/reg/2024/1689/oj
  - https://eur-lex.europa.eu/eli/reg/2022/2554/oj
  - https://www.nist.gov/itl/ai-risk-management-framework
  - https://csrc.nist.gov/pubs/sp/800/218/final
  - https://dora.dev/capabilities/
  - https://owasp.org/www-project-application-security-verification-standard/
  - https://genai.owasp.org/llm-top-10/
  - https://slsa.dev/spec/v1.2/
  - https://spdx.dev/learn/areas-of-interest/ai/
  - https://cyclonedx.org/capabilities/mlbom/
  - https://eur-lex.europa.eu/eli/reg/2024/2847/oj/eng
  - https://eur-lex.europa.eu/eli/dir/2022/2555/oj/eng
  - https://eur-lex.europa.eu/eli/reg/2016/679/oj
  - https://www.w3.org/TR/prov-o/
  - https://csrc.nist.gov/pubs/sp/800/53/r5/upd1/final
  - https://csrc.nist.gov/pubs/sp/800/61/r3/final
  - https://csrc.nist.gov/pubs/sp/800/218/a/final
  - https://www.nist.gov/privacy-framework
  - https://www.nist.gov/cyberframework
  - https://csrc.nist.gov/pubs/sp/800/207/final
  - https://owaspsamm.org/model/
  - https://in-toto.io/
  - https://theupdateframework.io/
  - https://docs.sigstore.dev/
  - https://securityscorecards.dev/
  - https://alistair.cockburn.us/hexagonal-architecture
  - https://learn.microsoft.com/en-us/azure/devops/pipelines/process/templates?view=azure-devops
  - https://learn.microsoft.com/en-us/azure/devops/pipelines/process/approvals?view=azure-devops
  - https://learn.microsoft.com/en-us/azure/devops/repos/git/branch-policies?view=azure-devops
  - https://learn.microsoft.com/en-us/azure/devops/boards/work-items/about-work-items?view=azure-devops
  - https://learn.microsoft.com/en-us/azure/devops/pipelines/library/connect-to-azure?view=azure-devops
  - https://learn.microsoft.com/en-us/rest/api/azure/devops/?view=azure-devops-rest-7.2
  - https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops
  - https://support.atlassian.com/bitbucket-cloud/docs/bitbucket-pipelines-configuration-reference/
  - https://support.atlassian.com/bitbucket-cloud/docs/integrate-pipelines-with-resource-servers-using-oidc/
  - https://support.atlassian.com/bitbucket-cloud/docs/set-up-and-monitor-deployments/
  - https://support.atlassian.com/bitbucket-cloud/docs/suggest-or-require-checks-before-a-merge/
  - https://support.atlassian.com/bitbucket-cloud/docs/use-branch-permissions/
  - https://support.atlassian.com/bitbucket-cloud/docs/share-pipelines-configurations/
  - https://support.atlassian.com/bitbucket-cloud/docs/set-up-and-use-custom-merge-checks/
  - https://developer.atlassian.com/cloud/bitbucket/rest/intro/
  - https://support.atlassian.com/bitbucket-cloud/docs/manage-webhooks/
  - https://confluence.atlassian.com/bitbucketserver
  - https://developer.atlassian.com/server/bitbucket/rest/v1003/
  - https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines.html
  - https://docs.bitrise.io/en/bitrise-ci/configure-builds/secrets
  - https://docs.bitrise.io/en/bitrise-ci/run-and-analyze-builds/managing-build-files/build-artifacts-online
  - https://docs.bitrise.io/en/bitrise-ci/run-and-analyze-builds/build-triggers/configuring-build-triggers
  - https://docs.bitrise.io/en/bitrise-platform/integrations/oidc-authentication/oidc-for-aws
  - https://docs.bitrise.io/en/bitrise-ci/configure-builds/configuring-build-settings/reporting-the-build-status-to-your-git-hosting-provider
  - https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/verified-steps
  - https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html
  - https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html
  - https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps.html
  - https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-user-guide.html
  - https://docs.aws.amazon.com/config/latest/developerguide/conformance-packs.html
  - https://docs.aws.amazon.com/securityhub/latest/userguide/what-is-securityhub.html
  - https://docs.datadoghq.com/continuous_integration/
  - https://docs.datadoghq.com/continuous_delivery/
  - https://docs.datadoghq.com/opentelemetry/
  - https://docs.datadoghq.com/account_management/audit_trail/
  - https://docs.datadoghq.com/internal_developer_portal/software_catalog/
  - https://docs.datadoghq.com/integrations/amazon-web-services/
  - https://docs.datadoghq.com/security/cloud_security_management/
  - https://docs.datadoghq.com/llm_observability/
  - https://docs.datadoghq.com/dora_metrics/
  versions:
    xonovex-workflow: 5.0.0
    agent-skills-spec: retrieved 2026-07-14
    mcp: '2025-06-18'
    opentelemetry-semconv: 1.43.0
    iso-iec-ieee-12207: '2026'
    iso-iec-ieee-15288: '2023'
    iso-iec-5338: '2023'
    iso-iec-42001: '2023'
    iso-iec-23894: '2023'
    iso-9241-210: 2019, confirmed 2025
    wcag: 2.2 Recommendation
    eu-ai-act: Regulation (EU) 2024/1689
    eu-dora: Regulation (EU) 2022/2554
    nist-ai-rmf: '1.0'
    nist-ssdf: '1.1'
    owasp-asvs: 5.0.0
    owasp-llm-top-10: '2025'
    slsa: '1.2'
    nist-sp-800-53: Rev. 5, Release 5.2.0
    nist-sp-800-61: Rev. 3 (2025)
    nist-sp-800-218a: 2024 final
    nist-csf: '2.0'
    nist-privacy-framework: 1.0 stable baseline
    nist-sp-800-207: '2020'
    eu-cra: Regulation (EU) 2024/2847
    eu-nis2: Directive (EU) 2022/2555 plus national transposition
    eu-gdpr: Regulation (EU) 2016/679
    traceability-baseline: '2026-07-14'
    azure-devops: official documentation retrieved 2026-07-14
    bitbucket-cloud: official documentation retrieved 2026-07-14
    bitbucket-data-center: 10.3 documentation baseline; deployment version must be detected
    bitrise: official documentation retrieved 2026-07-14
    aws: official documentation retrieved 2026-07-14
    datadog: official documentation retrieved 2026-07-14
traceability:
  source_registry: composable-workflow-phases/traceability/source-registry.md
  decision_source_matrix: composable-workflow-phases/traceability/decision-source-matrix.md
  control_crosswalk: composable-workflow-phases/traceability/control-crosswalk.md
  platform_capability_matrix: composable-workflow-phases/traceability/platform-capability-matrix.md
  subplan_traceability: composable-workflow-phases/traceability/subplan-traceability.md
  validation_policy: composable-workflow-phases/traceability/validation-policy.md
---

# Composable Workflow, Governance, Enablement, and Traceability

## Overview

Expand the approved Composable Workflow Phases architecture into a modular operating model with two independent but interoperable planes:

1. **Workflow plane** — lifecycle capabilities, provider-native results, methods, profiles, evidence, acceptance, and completion.
2. **Governance and enablement plane** — policies, enforcement points, harness adapters, scripts, bounded LLM evaluators, bounded agent launchers, external automation, CI/CD controls, onboarding advisors, observability, trust, and managed configuration.

The workflow plane remains fully supported. Named lifecycle operations may still compose workflow skills, method skills, provider skills, workspace skills, assessment skills, and result contracts. The new governance plane is not another lifecycle and is not restricted to workflow commands. It can control or advise any agent activity through harness hooks, external scripts, CI/CD, repository rules, admission controls, policy engines, provider APIs, and human approval points.

The two independently adoptable planes are modular and mix-and-match. A user may install only a Claude Code hook pack, only a GitHub Actions assurance pack, a complete organization governance bundle, or none of them. Harness-specific skills explain and optionally help configure native mechanisms for Claude Code, Codex, Kiro, GitHub Copilot CLI, Pi, and OpenCode. Provider skills similarly own GitHub and GitLab onboarding references for reusable CI components, rules, environments, compliance controls, and evidence publication.

The architecture remains storage-neutral. Workflow results and governance evidence are persisted through provider-native mechanisms and exchanged by opaque native references. Harness and CI configuration formats are native implementation details rather than workflow contracts.

An optional enterprise-platform extension pack adds source-backed skills and onboarding references for Azure DevOps, Bitbucket Cloud and Data Center, Bitrise, AWS, and Datadog. These modules remain independently installable and preserve the mixed-provider model: for example, Azure Boards may own work items, Bitbucket may own source and pull requests, Bitrise may own mobile CI evidence, AWS may own runtime controls, and Datadog may own observability evidence without introducing a central workflow record.

## Intent

The target is not a universal collection of commands. It is a **modular governance, execution, and onboarding architecture** that can answer five separate questions:

- What work or evidence is required? — profiles and policies.
- How should a capability be executed? — deterministic script, script plus bounded LLM, bounded agent, human task, or external system.
- Where can behavior be enforced or observed? — harness hooks and external enforcement points.
- How is an environment enabled safely? — advisory onboarding with discovery, preview, consent, application, verification, rollback, and drift detection.
- Where are results and evidence kept? — provider-native persistence and opaque references.

## Adoption Modes

The architecture supports independent adoption without forcing a complete platform migration:

- **Workflow-only** — lifecycle commands, method skills, provider skills, result contracts, and workflow profiles without harness governance modules.
- **Governance-only** — hooks, scripts, policy modules, CI/CD controls, onboarding advisors, trust, and observability around ordinary agent activity without requiring lifecycle commands.
- **Enablement-only** — platform-specific skills, references, diagnostics, setup helpers, and recommendations with no enforced controls.
- **External-enforcement-only** — CI/CD, repository rules, deployment approvals, admission controls, or policy services without agent-harness hooks.
- **Integrated composition** — any selected combination of workflow, governance, enablement, and external enforcement modules.

Every module must declare whether it is advisory, evidence-producing, enforcing, configuration-changing, or privileged. Presets may recommend useful combinations, but users remain free to select individual modules where dependencies and mandatory policy requirements are satisfied. Lifecycle commands are neither required by governance modules nor privileged over non-lifecycle agent activity.

## Execution Patterns

The execution plane supports four bounded technical patterns plus human and external execution:

1. **Deterministic script or API operation** — preferred for inspection, validation, policy decisions, provider actions, inventories, state transitions, and privileged execution.
2. **Script plus bounded LLM** — deterministic coordinator resolves inputs, constructs bounded context, calls a model, validates structured output, and publishes evidence.
3. **Script launching a bounded specialist agent** — explicit adaptive investigation with restricted tools, attenuated authority, recursion limits, budget, cancellation, and a required result contract.
4. **Advisory onboarding agent** — discovers the actual environment, recommends native modules, explains permissions and trade-offs, previews exact changes, and never mutates configuration without authorization.
5. **Human or external authoritative execution** — accountable decisions, CI, scanners, deployment systems, identity systems, monitoring, GRC, and other systems of record.

Hooks are deterministic orchestration and enforcement points by default. A hook may call a bounded model or launch a bounded agent only through an explicit module contract; it must not hide open-ended work, silently grant authority, or treat model output as authoritative evidence when deterministic or external evidence is available.

## Architectural Planes

### 1. Semantic workflow plane

Owns stable result contracts for Discovery, Research, Formulation, Experience Design, Solution Design, Decision, Planning, Development, Deliverable Publication, Review, QA, Assessment, Inventory, Data Governance, Acceptance, Integration, Transition, Release, Observation, Incident, Corrective Action, Retirement, and Learning. Profiles compose these capabilities and evidence requirements without prescribing actor, runtime, invocation syntax, or storage format.

### 2. Governance and policy plane

Owns applicability, control crosswalks, policy bundles, actor requirements, segregation of duties, exception/waiver handling, break-glass behavior, failure policy, authorization, data handling, model routing constraints, and evidence freshness. Policy decision logic is separate from enforcement. A policy engine such as OPA may be used, but no policy language or engine is mandatory.

### 3. Execution plane

Defines executor classes and capability execution contracts:

- `deterministic` — script, API call, scanner, CI job, or provider operation.
- `model` — bounded LLM call with fixed inputs, output contract, validation, and retry limits.
- `agent` — adaptive multi-step execution with explicit tools, budget, depth, authority, and result boundary.
- `human` — accountable review, approval, judgment, or manual action.
- `external` — authoritative system such as CI, deployment, monitoring, GRC, or identity platform.

A capability declares permitted and preferred executor classes. Deterministic-first is the default. An agent is used only when discovery-driven branching or adaptive tool use is genuinely needed.

### 4. Enforcement plane

Defines semantic enforcement intents and maps them to native control points. Examples include session start/end, prompt submission, before/after tool use, model call boundaries, subagent start/stop, capability before/after/failure, result publication, configuration change, compaction, workspace changes, and privileged operations. Harness adapters report which events and handler types they actually support; profiles must not assume feature parity.

External enforcement points include Git hooks, pre-commit frameworks, CI/CD, repository rules, protected environments, deployment approvals, Kubernetes admission webhooks, API gateways, secrets systems, policy decision services, and provider-native controls. Defense in depth is encouraged: a harness hook is not treated as the sole security boundary.

### 5. Enablement and onboarding plane

Uses skills, scripts, templates, plugins, packages, and advisory agents to help people configure their own environment. Onboarding follows an explicit lifecycle:

1. Discover the environment and current controls.
2. Assess supported harness/platform capabilities and gaps.
3. Recommend compatible modules and explain trade-offs.
4. Preview exact changes, permissions, data flows, and failure behavior.
5. Obtain consent or required approval.
6. Apply idempotently through native configuration mechanisms.
7. Dry-run and verify expected behavior.
8. Record an opaque native reference to the applied configuration or evidence.
9. Support rollback, disablement, upgrade, and drift detection.

Advisory skills may customize recommendations using repository, organization, harness, CI, runtime, and policy context. They may invoke a bounded LLM or launch a bounded specialist agent, but must not hide open-ended agent work inside an opaque hook or silently mutate configuration.

### 6. Provider and evidence plane

Provider skills remain the only adapters for persistence and platform-native operations. The same model applies to policy, telemetry, configuration, and evidence providers. No central workflow database, universal result envelope, or required YAML sidecar is introduced.

### 7. Observability and assurance plane

Produces audit and operational evidence without collecting more sensitive content than necessary. OpenTelemetry semantic conventions are the preferred interoperability direction for traces and metrics covering generative AI, agents, MCP, CI/CD, and CLI activity, but the workflow does not mandate one telemetry backend. Runtime trace IDs are execution correlation identifiers, not workflow identities.

### 8. Distribution and trust plane

Treats hooks, scripts, plugins, extensions, MCP servers, skills, CI components, actions, and policy bundles as executable supply-chain artifacts. Modules must declare source, version, compatibility, requested permissions, side effects, data access, network use, secrets, failure mode, evidence output, upgrade policy, and rollback. Organization-managed modules may be enforced; user/project modules must be reviewable and trusted before execution.

## Trust and Authority Zones

Module composition crosses several authority zones that must remain visible:

- **Organization-managed** — centrally approved and enforced modules, policies, CI controls, and managed harness configuration.
- **Repository/project** — project-owned instructions, plugins, hooks, CI components, and provider configuration that may execute code and therefore require repository trust.
- **User** — personal skills, hooks, credentials, models, extensions, and preferences that cannot silently weaken managed controls.
- **Session/runtime** — ephemeral context, tool grants, workspaces, temporary credentials, budgets, and agent-specific restrictions.
- **External authority** — CI, identity, secrets, deployment, monitoring, policy, GRC, and provider systems whose evidence or decisions may be authoritative.

Adapters document native precedence, but the semantic rule is stable: lower-authority configuration cannot silently weaken a mandatory higher-authority control. Conflicts, unsupported enforcement, and authority expansion fail visibly.

## Research Findings

### Harnesses expose different native extension models

- Claude Code supports deterministic command hooks and also prompt, agent, HTTP, and MCP-tool hook types. Its event surface includes session, prompt, tool, subagent, task, configuration, worktree, compaction, and session-end events. Agent hooks are explicitly experimental, so the adapter must version capability support rather than assume stability.
- Codex supports lifecycle command hooks, plugin-packaged hooks, and managed hooks. Current official documentation states that prompt and agent handler types are parsed but not executed, so model- or agent-based governance must be implemented as a command hook invoking an external bounded runner or through another explicit capability until native support exists.
- Kiro supports shell-command and agent-prompt actions on file, prompt, tool, task, session, and stop events.
- GitHub Copilot supports repository and personal hooks for the cloud agent and Copilot CLI. Its native hook handlers are shell commands, so model or agent evaluation must be launched explicitly by those commands.
- OpenCode uses JavaScript/TypeScript plugins with session, permission, shell, tool-before/after, file, command, and UI events; plugins can add tools and alter arguments.
- Pi is deliberately minimal. TypeScript extensions, skills, prompt templates, and packages supply hooks, permissions, context injection, subagent patterns, and integrations. Its lack of a mandatory built-in sandbox means onboarding must make isolation and trust requirements explicit.

These differences require a common **semantic intent taxonomy plus per-harness adapters and a versioned capability matrix**, not one universal hook configuration file.

### Skills, plugins, and protocol integrations have different responsibilities

The Agent Skills specification supports instructions plus optional scripts, references, and assets with progressive disclosure. Skills are therefore suitable for platform-specific onboarding knowledge and setup helpers, but are not enforcement boundaries by themselves. Plugins or packages can bundle skills with hooks, extensions, and configuration where a harness supports that model.

MCP provides a portable way to expose resources, prompts, and tools and includes explicit consent and tool-safety principles. MCP is useful for integrations and context, but it is not a replacement for harness permissions, policy enforcement, sandboxing, or provider authorization.

### Policy decision and enforcement should be decoupled

A policy decision point evaluates contextual facts and returns allow, deny, ask, advise, or required evidence. A policy enforcement point applies that decision at a harness hook, CI gate, repository rule, deployment environment, admission webhook, or provider operation. OPA/Rego is one possible implementation for portable policy-as-code, but simple deterministic policies and provider-native engines remain valid.

### CI/CD onboarding should use native reusable primitives

GitHub onboarding should prefer version-pinned reusable workflows, composite actions, rulesets, required checks, protected environments, least-privilege tokens, and artifact attestations rather than copying large workflow files into every repository. GitLab onboarding should prefer versioned CI/CD components with typed inputs, pipeline execution policies, compliance frameworks, and provider-native evidence. Platform adapters must preserve native strengths and document merge/precedence behavior.

### Enterprise platforms require product- and edition-specific adapters

Azure DevOps exposes distinct native mechanisms for Boards work items, Repos branch policies, Pipeline templates and approvals, service connections, REST APIs, and service hooks. Bitbucket Cloud and Bitbucket Data Center must be treated as separate capability surfaces rather than one interchangeable provider; Cloud adds Pipelines, OIDC, deployment environments, shared configurations, merge checks, and cloud REST/webhook behavior, while Data Center has separately versioned server APIs and deployment constraints. Bitrise contributes mobile-focused Workflows, Pipelines, verified Steps, triggers, artifacts, secrets, build status, and OIDC-based AWS access. AWS contributes workload federation, temporary credentials, least privilege, organization guardrails, audit, configuration, and security-posture evidence. Datadog contributes CI/CD visibility, DORA metrics, OpenTelemetry ingestion, LLM observability, audit, catalog, AWS, and cloud-security evidence.

The architecture therefore adds focused optional skills rather than embedding these platforms into workflow contracts. Every skill must detect product edition, account/tier and tested version, preserve native identifiers and evidence, disclose credential/data flows, and fail clearly when a capability is unavailable. Mixed-stack onboarding is a first-class composition, not an exception.

## Authoritative Grounding and Traceability

The architecture is **authoritatively grounded but not wholly mandated by external sources**. Every requirement and decision must be labeled as one of:

- binding legal obligation, after applicability and qualified review;
- international-standard or government-framework alignment;
- vendor-platform fact, verified through a version-pinned conformance test;
- industry-specification or authoritative-practice guidance;
- Xonovex architectural synthesis or explicit scope decision.

The plan therefore maintains six auditable artifacts:

- [Authoritative source registry](composable-workflow-phases/traceability/source-registry.md)
- [Decision–source matrix](composable-workflow-phases/traceability/decision-source-matrix.md)
- [Control and obligation crosswalk](composable-workflow-phases/traceability/control-crosswalk.md)
- [Platform capability and limitation matrix](composable-workflow-phases/traceability/platform-capability-matrix.md)
- [Subplan task traceability](composable-workflow-phases/traceability/subplan-traceability.md)
- [Traceability validation policy](composable-workflow-phases/traceability/validation-policy.md)

### Mapping rules

1. A source supports only claims within its recorded scope.
2. A crosswalk is contextual and never proves equivalence, certification or legal compliance.
3. Regulated controls activate only after applicability is evaluated for jurisdiction, actor role, entity/product/system class and date.
4. Public ISO abstracts support broad architectural direction; detailed normative claims require the licensed text.
5. Vendor documentation supports a candidate adapter capability; release requires a pinned-version conformance probe.
6. Architectural synthesis remains explicitly labeled even when multiple authoritative sources support the concern.
7. Evidence identifies its source, actor/executor, native subject revision, policy/control version and freshness.
8. Source drift, supersession or platform behavior changes trigger review and can invalidate mappings.

### Audited decision baseline

The detailed matrix contains **39 decisions**, including the two-plane architecture, deterministic-first execution, storage-neutral opaque references, authority attenuation, defense in depth, provider-native evidence, independent adoption modes, exact-revision assurance and the separation of Acceptance from Integration. Decisions classified as synthesis are not presented as external mandates.

### Audited control baseline

The crosswalk contains **45 control/obligation families** spanning access and authority, executable-module trust, privacy and telemetry, incident/recovery, secure development, AI governance, EU AI Act, EU DORA, CRA, NIS2, GDPR, accessibility, inventories/provenance, onboarding, external enforcement, evidence freshness, exceptions and lifecycle retirement.

### Legal and standards caveat

This plan is an engineering traceability baseline, not legal advice, certification or an ISO conformity assessment. Legal profiles require current consolidated legislation and qualified review. ISO clause-level verification requires access to the licensed standards. The implementation must preserve those review statuses rather than converting source mappings into automatic compliance claims.

## Goals

- Maintain complete source, decision, control, platform and task traceability with explicit mapping strength and caveats.
- Prevent architectural synthesis, vendor behavior, standards alignment and legal obligation from being conflated.
- Validate source freshness, exact legal/standards review status and platform conformance before release.

- Preserve all approved composable workflow and provider-native result contracts.
- Support workflow-only, governance-only, enablement-only, external-enforcement-only, and integrated adoption modes.
- Add an executor-neutral capability model with deterministic-first selection.
- Define semantic governance events and enforcement intents without requiring one hook schema.
- Create versioned harness capability matrices and adapters for Claude Code, Codex, Kiro, GitHub Copilot CLI, Pi, and OpenCode.
- Create platform onboarding references and reusable modules for GitHub Actions and GitLab CI/CD.
- Add independently composable, source-backed skills for Azure DevOps, Bitbucket Cloud/Data Center, Bitrise, AWS, and Datadog, including a mixed-enterprise-stack composition guide.
- Allow hooks/scripts to call deterministic tools, bounded LLM evaluators, or bounded agents with explicit budgets and authority.
- Support advisory environment onboarding separate from mandatory enforcement.
- Support organization-managed, project, user, and session configuration layers without pretending their precedence is identical across platforms.
- Decouple policy decision points from enforcement points.
- Add exception, waiver, break-glass, rollback, and drift-management contracts.
- Add module inventory and provenance to AIBOM/InventoryResult where agent environments are in scope.
- Provide portable observability and audit guidance using OpenTelemetry-compatible semantics where possible.
- Maintain storage neutrality and opaque provider-native references.
- Make advisory, evidence-producing, enforcing, configuration-changing, and privileged module behavior visible before installation or execution.

## Non-goals

- Define one universal hook JSON/YAML format or normalize away native harness capabilities.
- Require lifecycle commands in order to use governance, onboarding, CI, or harness modules.
- Require governance modules in order to use lifecycle commands and provider-native workflow results.
- Require every user to install every harness, CI, policy, or lifecycle module.
- Treat skills as enforcement, or hooks as the sole security boundary.
- Run an LLM or agent where deterministic validation is sufficient.
- Hide open-ended agent launches in routine hooks.
- Automatically modify user, organization, repository, CI, or harness configuration without preview and consent.
- Assume all harnesses support prompt hooks, agent hooks, tool blocking, managed policy, or equivalent event ordering.
- Make MCP, OPA, OpenTelemetry, GitHub Actions, GitLab CI/CD, Kubernetes, or one model provider mandatory.
- Require Azure DevOps, Bitbucket, Bitrise, AWS, Datadog, or one enterprise stack; the new skills are optional platform modules.
- Persist a universal governance object or require YAML sidecars.
- Replace legal, compliance, security, privacy, safety, accessibility, or operational professionals.

## Settled Decisions

1. The architecture has separate workflow and governance/enablement planes that compose but remain independently adoptable.
2. Capabilities are executor-neutral. Deterministic scripts and authoritative systems are preferred; bounded LLMs and agents are selected only where semantic or adaptive work requires them.
3. Semantic enforcement intents are portable; concrete hook events and configuration are harness-native and versioned.
4. A harness adapter publishes a capability matrix including events, handler types, blocking semantics, ordering/concurrency, managed configuration, context injection, and limitations.
5. Policy decisions are separate from enforcement. Enforcement points fail closed only when the selected profile says the control is mandatory.
6. Agent-launch modules require explicit depth, tool, model, token/cost, time, network, data, and authority limits. Child agents inherit no more authority than their launcher and normally receive less.
7. Onboarding is advisory-first and transactional: discover, recommend, preview, approve, apply, verify, rollback, and monitor drift.
8. Skills may bundle or reference scripts and templates, and platform plugins/packages may install native hooks. Skills alone do not prove enforcement.
9. Repository- or project-provided executable extensions must be trusted before execution; organization-managed modules require provenance and change control.
10. CI onboarding uses native reusable components, rules, and protected environments rather than copied pipeline boilerplate.
11. Telemetry uses data minimization, explicit retention, redaction, and access policy. Raw prompts and tool outputs are not logged by default.
12. All workflow results, policy evidence, onboarding evidence, and module records remain provider-native and are exchanged by opaque references.
13. Workflow, governance, enablement, and external-enforcement modules are independently adoptable; presets are compositions rather than mandatory products.
14. Hooks are deterministic by default. Model and agent execution from hooks is explicit, bounded, authority-attenuated, observable, cancellable, and represented as a declared module.
15. Organization, project, user, session, and external authority zones remain explicit; lower-authority configuration cannot silently weaken mandatory higher-authority controls.

## Module Model

Every reusable governance or enablement module declares:

```text
module identity and version
source and provenance
supported platforms and versions
semantic intent(s)
native events or external enforcement points
executor class: deterministic | model | agent | human | external
inputs and expected output/evidence
permissions, tools, filesystem, network, secrets, and data access
side effects and idempotency
ordering, concurrency, timeout, retry, and reentrancy behavior
failure mode: deny | ask | warn | observe | ignore
sensitivity, retention, and redaction behavior
upgrade, disable, rollback, and drift-detection behavior
owner and support status
```

This is a semantic conformance contract, not a mandatory serialized manifest. Each platform may represent it natively.

## Harness Adapter Model

Each harness-specific skill owns focused references and optional setup helpers for:

- Native hook/event vocabulary.
- Native plugin/package/extension format.
- Configuration scopes and precedence.
- Managed versus user/project configuration.
- Blocking and permission semantics.
- Context injection and compaction behavior.
- Subagent/agent-launch options.
- Security, trust, sandbox, and execution caveats.
- Testing, dry-run, diagnostics, and rollback.
- Mapping from semantic governance intents to native mechanisms.

Initial platform owners:

- Claude Code governance and hooks reference.
- Codex governance, hooks, plugins, and managed-configuration reference.
- Kiro hooks and steering reference.
- GitHub Copilot CLI/cloud-agent hooks and plugin reference.
- Pi extensions, packages, skills, trust, and isolation reference.
- OpenCode plugins, permissions, custom tools, and event reference.

## External Automation Model

Platform-specific onboarding references and modules initially cover:

- GitHub Actions reusable workflows and composite actions.
- GitHub repository/organization rulesets, required checks, protected environments, approvals, and secure workflow practices.
- GitLab versioned CI/CD components, typed inputs, pipeline execution policies, compliance frameworks, and security-policy projects.
- Generic Git hooks and pre-commit integrations where appropriate.
- Kubernetes admission and runtime policy integration with Xonovex AgentPolicy.
- Optional policy decision services and provider-native GRC integrations.

## Governance and Onboarding Flow

```text
Environment discovery
    ↓
Capability and risk assessment
    ↓
Recommended composition
    ↓
Human-readable preview and permission report
    ↓
Approval / managed authorization
    ↓
Native installation or configuration
    ↓
Dry-run and conformance verification
    ↓
Operational evidence and telemetry
    ↓
Drift, update, exception, rollback, or removal
```

The advisor may use deterministic probes, a bounded LLM for interpretation, or a bounded specialist agent for complex discovery. Every recommendation distinguishes advisory guidance from enforcing controls.

## Missing Aspects Addressed by This Revision

- **Configuration trust and supply chain:** hooks and plugins are executable code and require provenance, pinning, review, least privilege, and update policy.
- **Capability negotiation:** profiles must validate what the selected harness and platform can actually enforce.
- **Ordering and concurrency:** multiple hooks may run concurrently; deny decisions do not necessarily suppress sibling side effects. Modules must not assume serial execution unless the adapter guarantees it.
- **Idempotency and reentrancy:** file-save, tool, retry, and nested-agent events can recur. Setup and enforcement modules must tolerate repetition.
- **Exception management:** waivers include scope, owner, justification, compensating controls, expiry, and revalidation.
- **Break-glass:** emergency bypass is explicit, time-bound, logged, reviewed, and cannot silently become a default.
- **Authority attenuation:** spawned agents and external jobs receive the minimum authority needed and cannot elevate beyond the launcher.
- **Data governance:** prompt, context, tool, telemetry, and external model data flows require classification, redaction, retention, residency, and consent rules.
- **Drift and version compatibility:** harness APIs and hook behavior change; capability matrices and installed modules are versioned and continuously checked.
- **Testing and simulation:** hooks, policies, onboarding, and modules need fixtures, dry-run, fault injection, adversarial input, timeout, concurrency, and rollback tests.
- **Explainability:** denials and recommendations identify the policy, evidence, native enforcement point, and next remediation action.
- **Ownership and support:** every module and policy has an accountable owner, lifecycle status, support channel, and retirement path.
- **Catalog and composition UX:** users need searchable modules, compatibility information, presets, conflicts, and a clear effective composition.
- **Observability and cost:** model/agent calls, tool use, CI, and policy decisions need metrics and tracing without defaulting to sensitive-content capture.

## Proposed Approach

0. Establish the authoritative source registry, decision register, control/article crosswalk, platform capability matrix, subplan task traceability and automated validation policy before implementation proceeds.

1. Freeze the two-plane contracts, independent adoption modes, execution classes and patterns, semantic event intents, policy decision/enforcement separation, module conformance, profile composition, authority zones, actor requirements, and onboarding lifecycle.
2. Extend provider conformance to policy, configuration, evidence, telemetry, and module-distribution providers while preserving opaque references.
3. Add harness-specific skills and adapters for Claude Code, Codex, Kiro, GitHub Copilot CLI, Pi, and OpenCode, each with a versioned capability matrix and native setup/rollback guidance.
4. Add GitHub and GitLab automation/onboarding modules using reusable native primitives and protected policy mechanisms.
5. Add optional enterprise-platform skills for Azure DevOps, Bitbucket Cloud/Data Center, Bitrise, AWS, and Datadog, plus a mixed-stack onboarding reference and conformance fixtures.
6. Prove the architecture with a walking skeleton that discovers an environment, recommends a composition, previews and applies a native hook/module, verifies a deterministic policy, launches one bounded advisory evaluator, publishes provider-native evidence, and rolls back.
7. Retain and consolidate the complete workflow lifecycle implementation under the new execution and enforcement contracts.
8. Add privileged-operation enforcement around Acceptance, Integration, Transition, Release, data deletion, incident handling, and Retirement.
9. Implement cross-cutting governance, policy, observability, trust, exceptions, module inventory, drift, learning, and operations.
10. Reconcile documentation, diagrams, migration, source crosswalks, platform compatibility matrices, evaluation suites, package metadata, and release/versioning.

## Proposed Child Plans

| Group | Child plan | Purpose | Depends on |
| --- | --- | --- | --- |
| 1 | `workflow-governance-contracts-profiles-and-execution` | Freeze semantic workflow, policy, execution, hook-intent, module, profile, actor, and onboarding contracts | None |
| 2 | `provider-policy-and-module-conformance` | Extend storage-neutral provider conformance to policy/evidence/configuration/module providers and executable-module trust | Group 1 |
| 3 | `agent-harness-adapters-and-onboarding` | Implement harness-specific skills, adapters, capability matrices, setup helpers, and rollback guidance | Groups 1–2 |
| 3 | `external-automation-ci-and-platform-enforcement` | Implement GitHub/GitLab CI and external enforcement/onboarding modules | Groups 1–2 |
| 3 | `enterprise-platform-skills-and-onboarding` | Add optional Azure DevOps, Bitbucket Cloud/Data Center, Bitrise, AWS, and Datadog skills, onboarding, native evidence mappings, and mixed-stack guidance | Groups 1–2 |
| 4 | `governance-onboarding-walking-skeleton` | Prove advisory onboarding, deterministic enforcement, bounded evaluator/agent launch, evidence, and rollback | Group 3 |
| 5 | `workflow-discovery-research-design-decision-and-planning` | Implement early lifecycle capabilities under executor-neutral and provider-native contracts | Group 4 |
| 5 | `workflow-development-delivery-inventory-assessment-review-and-qa` | Implement development and assurance with hooks, CI, inventories, and exact-revision evidence | Group 4 |
| 5 | `workflow-acceptance-integration-transition-release-observation-incidents-and-retirement` | Implement human/privileged operations and operational lifecycle with external enforcement | Group 4 |
| 6 | `governance-policy-learning-observability-trust-and-operations` | Apply policy, exceptions, telemetry, catalogs, trust, drift, learning, and operational governance across both planes | Group 5 and harness/CI plans |
| 7 | `documentation-migration-crosswalk-and-validation` | Reconcile docs, diagrams, migration, compatibility, crosswalks, packages, and full conformance | Group 6 |

## Risk Assessment

| Traceability or crosswalk overstates authority | Teams may treat guidance as law, assume compliance, or enforce the wrong requirement | Classify every mapping, require applicability and qualified review, label synthesis, and fail validation on equivalence/certification claims |
| Sources or platform behavior drift | Controls may silently become stale or ineffective | Pin versions/dates, run link and conformance checks, record supersession, and block release when mandatory mappings are unresolved |

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Governance plane becomes a second monolith | Hooks, policies, onboarding, and lifecycle logic couple | Separate semantic intents, adapters, policy providers, modules, and lifecycle owners; enforce progressive disclosure |
| One lowest-common-denominator hook model | Native strengths are lost or unsupported controls appear portable | Keep semantic intents portable but publish per-harness capability matrices and native adapters |
| Hooks are mistaken for a security boundary | Equivalent tool paths or external actions bypass controls | Defense in depth through sandbox, CI, repository rules, provider permissions, and external enforcement |
| Open-ended agents run invisibly from hooks | Cost, permissions, recursion, and audit become unpredictable | Agent launches are explicit modules with authority attenuation, budgets, depth limits, evidence, and kill switches |
| Onboarding mutates environments unexpectedly | Trust and adoption collapse | Preview, consent, idempotency, dry-run, rollback, and change records are mandatory |
| Executable module supply chain is compromised | Hooks/plugins gain code execution and secrets | Pin versions, verify provenance, least privilege, trust project code, scan dependencies, and provide emergency disablement |
| Platform updates break enforcement | Controls silently stop firing | Versioned compatibility matrices, conformance probes, drift monitoring, and fail-closed behavior for mandatory controls |
| Telemetry leaks prompts, source, or secrets | Privacy, legal, and security harm | Data minimization, redaction, sampling, retention, residency, access control, and content logging off by default |
| Multiple hooks interfere | Ordering, duplicated side effects, or inconsistent decisions | Declare concurrency/idempotency and test combined modules; adapters document merge semantics |
| Policy-as-code becomes mandatory complexity | Small teams cannot adopt the system | Support simple deterministic rules and provider-native policies; OPA remains optional |
| CI onboarding copies brittle boilerplate | Divergence and insecure defaults proliferate | Use reusable workflows/actions/components and centrally managed policies with version pinning |
| Cloud and self-managed editions are conflated | An adapter may promise unavailable APIs, CI, merge checks, or governance behavior | Detect Azure DevOps Services/Server and Bitbucket Cloud/Data Center explicitly; publish separate capability matrices and conformance results |
| CI-to-AWS setup relies on static keys | Long-lived credentials create avoidable exposure and difficult rotation | Prefer OIDC/workload federation and short-lived role credentials where supported; preview trust policies and scope claims before applying |
| Observability onboarding leaks sensitive content | Source, prompts, personal data, secrets, or customer telemetry may be over-collected | Datadog modules declare collection, redaction, retention, residency, access and sampling; content capture is off unless explicitly selected |
| Skills are confused with enforcement | Documentation is treated as a control | Surface enforcement evidence separately from skill installation or recommendation |

## Success Criteria

- Every settled decision, mandatory control, platform capability and numbered subplan task resolves to valid source/decision/control IDs.
- Legal mappings identify applicability conditions, candidate article references, current-text review and qualified-review status.
- Detailed ISO claims identify licensed-text verification status; public abstracts are not used as clause-level evidence.
- Architectural synthesis is explicitly labeled and never represented as a source mandate.
- Crosswalks state gaps and mapping strength and never claim automatic equivalence, certification or compliance.
- Platform adapter claims are backed by official documentation and pinned-version conformance evidence.

- The existing composable lifecycle remains operational and storage-neutral.
- Workflow-only, governance-only, enablement-only, external-enforcement-only, and integrated compositions are independently valid and documented.
- Workflow, governance, execution, enforcement, enablement, provider, observability, and distribution concerns have explicit ownership boundaries.
- Capabilities declare permitted/preferred executor classes and validation requirements.
- Semantic hook intents map to native Claude Code, Codex, Kiro, Copilot CLI, Pi, and OpenCode mechanisms through versioned adapters.
- The capability matrix accurately records unsupported or experimental features; no documentation claims parity that does not exist.
- Harness-specific skills include focused native hook/plugin/extension references and safe onboarding helpers.
- GitHub and GitLab skills include reusable CI/CD, policy, protected-environment, and onboarding references.
- Azure DevOps, Bitbucket Cloud/Data Center, Bitrise, AWS, and Datadog have separate optional skills with official-source registries, edition/version detection, native capability matrices, safe onboarding, rollback, and conformance tests.
- A mixed-stack fixture demonstrates Azure Boards work items, Bitbucket source/pull requests, Bitrise CI evidence, AWS workload federation/runtime controls, and Datadog observability without centralizing provider-native results.
- CI-to-AWS onboarding prefers temporary federated credentials and least privilege; long-lived static keys are never generated by default.
- Onboarding can discover, recommend, preview, apply, verify, rollback, and detect drift without requiring one configuration format.
- Deterministic scripts are preferred; script-plus-model and script-launched-agent modules have schemas, budgets, authority attenuation, cancellation, and evidence.
- Mandatory controls can fail closed at at least two independent enforcement layers in a reference composition.
- Advisory modules remain nonblocking and visibly distinct from enforcement.
- Policies support allow, deny, ask, advise, required evidence, exception, and break-glass outcomes.
- Executable modules declare provenance, permissions, data flows, side effects, failure behavior, compatibility, and lifecycle status.
- Project-provided executable modules are trusted before execution; managed modules are auditable and version-pinned.
- OpenTelemetry-compatible events can correlate sessions, model/tool calls, policy decisions, CI activity, and evidence without creating a workflow identity or logging sensitive content by default.
- CI/CD onboarding uses versioned reusable workflows/actions/components and provider-native rules/policies rather than duplicated boilerplate.
- The walking skeleton proves a non-file provider, one harness, one CI platform, one deterministic hook, one bounded LLM or agent evaluator, result publication, fresh-context recovery, and rollback.
- Full tests cover concurrency, retries, reentrancy, timeouts, failures, bypass attempts, drift, upgrade, rollback, and conflicting modules.

## Estimated Effort

**Very large.** This is now a platform architecture spanning workflow semantics, policy/governance, harness adapters, onboarding, CI/CD, provider conformance, executable-module trust, observability, lifecycle implementation, and documentation. The eleven child plans include one deliberate exception to the usual ten-plan split guideline: the company-specific enterprise-platform pack is independently releasable and optional; the remaining work should be released incrementally behind capability/version declarations.
