# Sources

The contract is an Xonovex architectural synthesis. Sources support the recorded concern or constraint; they do not mandate this exact architecture or establish equivalence, certification, or legal compliance.

## S-ISO-12207 — Software lifecycle processes

- **URL:** https://www.iso.org/standard/90219.html
- **Last reviewed:** 2026-07-14
- **Used for:** `references/results.md`, `references/profiles.md`, `references/transition-run.md`, `references/retirement-run.md`
- **Aspects extracted:** Lifecycle breadth and flexible process application. The public abstract does not support detailed normative claims; those require licensed text.

## S-ISO-15288 — System lifecycle processes

- **URL:** https://www.iso.org/standard/81702.html
- **Last reviewed:** 2026-07-14
- **Used for:** `references/results.md`, `references/profiles.md`, `references/transition-run.md`, `references/retirement-run.md`
- **Aspects extracted:** Lifecycle breadth through retirement and flexible information representation. The public abstract does not support detailed normative claims; those require licensed text.

## S-W3C-PROV — PROV-O

- **URL:** https://www.w3.org/TR/prov-o/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/architecture.md`, `references/results.md`
- **Aspects extracted:** Provenance relationships among results, activities, and actors without mandating PROV serialization.

## S-AGENTSKILLS — Agent Skills specification

- **URL:** https://agentskills.io/specification
- **Last reviewed:** 2026-07-14
- **Used for:** `SKILL.md`, `references/conformance.md`
- **Aspects extracted:** Progressive-disclosure skill packaging; a skill is not an enforcement boundary.

## S-HEXAGONAL — Hexagonal architecture

- **URL:** https://alistair.cockburn.us/hexagonal-architecture
- **Last reviewed:** 2026-07-14
- **Used for:** `references/architecture.md`, `references/results.md`, `references/inspect.md`, `references/conformance.md`
- **Aspects extracted:** Provider ports and technology-specific adapters with dependencies directed toward semantic contracts.

## S-SUPPLY-CHAIN-INVENTORY — Optional software and AI inventories

- **URL:** https://spdx.dev/learn/areas-of-interest/ai/
- **Related URLs:** https://cyclonedx.org/capabilities/mlbom/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/results.md`, `references/providers.md`, `references/inventory-generate.md`, `references/conformance.md`
- **Aspects extracted:** Optional inventories may cover AI/ML and agent/governance components with identities, versions, relationships, and provenance. No one serialization is required by the workflow contract.

## S-SECURE-DEVELOPMENT-ASSURANCE — Secure development and verification

- **URL:** https://csrc.nist.gov/pubs/sp/800/218/final
- **Related URLs:** https://csrc.nist.gov/pubs/sp/800/218/a/final · https://owasp.org/www-project-application-security-verification-standard/ · https://genai.owasp.org/llm-top-10/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/development-contracts.md`, `references/develop-run.md`, `references/develop-consolidate.md`, `references/develop-abandon.md`, `references/assurance-contracts.md`, `references/assessment-run.md`, `references/review-run.md`, `references/qa-run.md`
- **Aspects extracted:** Secure-development evidence, versioned verification criteria, bounded model/agent risks, prompt-injection resistance, evidence origin, and adversarial evaluation. These mappings are engineering guidance, not certification or legal compliance.

## S-SUPPLY-CHAIN-EVIDENCE — Provenance and verification

- **URL:** https://slsa.dev/spec/v1.2/
- **Related URLs:** https://in-toto.io/ · https://docs.sigstore.dev/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/deliver-publish.md`, `references/inventory-generate.md`, `references/assurance-contracts.md`
- **Aspects extracted:** Exact subject and artifact revisions, producing-step provenance, optional signing/verification, and evidence freshness. Implementations pin the selected specification and trust policy.

## S-WCAG22 — Accessibility assessment

- **URL:** https://www.w3.org/TR/WCAG22/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/assessment-run.md`, `references/qa-run.md`
- **Aspects extracted:** Accessibility assessments pin the applicable criteria, version, and conformance level and preserve automated and human evidence separately.

## S-NATIVE-CI-ENFORCEMENT — Reusable GitHub and GitLab controls

- **URL:** https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
- **Related URLs:** https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets · https://docs.gitlab.com/ci/components/ · https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/assurance-contracts.md`, `references/qa-run.md`, `references/integration-validate.md`, `references/integration-run.md`, `references/release-run.md`
- **Aspects extracted:** Provider-native reusable modules, immutable revision pinning, stable required-check/job identity, policy enforcement, native evidence, and platform-specific bypass/precedence behavior. Provider skills own detailed mechanics and capability checks.

## S-OPERATIONAL-AUTHORITY — Access, incident, recovery, and protected change

- **URL:** https://csrc.nist.gov/pubs/sp/800/53/r5/upd1/final
- **Related URLs:** https://csrc.nist.gov/pubs/sp/800/61/r3/final · https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments
- **Last reviewed:** 2026-07-14
- **Used for:** `references/operational-contracts.md`, `references/acceptance-validate.md`, `references/acceptance-decide.md`, `references/integration-validate.md`, `references/integration-run.md`, `references/transition-run.md`, `references/release-run.md`, `references/incident-run.md`, `references/corrective-action-run.md`, `references/retirement-run.md`
- **Aspects extracted:** Accountable authorization, segregation of duties, protected target changes, least privilege, incident preparation/detection/response/recovery, contingency and rollback evidence, scoped emergency access, and post-use review. The workflow contract is an architectural synthesis; provider features and organization-specific control mappings require conformance and qualified review.

## S-OPERATIONAL-OBSERVATION — Monitoring and accessibility evidence

- **URL:** https://dora.dev/capabilities/
- **Related URLs:** https://www.w3.org/TR/WCAG22/ · https://www.nist.gov/itl/ai-risk-management-framework
- **Last reviewed:** 2026-07-14
- **Used for:** `references/observe-run.md`, `references/incident-run.md`, `references/corrective-action-run.md`
- **Aspects extracted:** Delivery/outcome feedback, user and system monitoring, testable accessibility evidence, AI-risk observation, incident escalation, and effectiveness feedback. Delivery research is not EU DORA law; every source remains separately identified.

## S-REGULATED-APPLICABILITY — EU lifecycle and incident obligations

- **URL:** https://eur-lex.europa.eu/eli/reg/2024/1689/oj
- **Related URLs:** https://eur-lex.europa.eu/eli/reg/2022/2554/oj · https://eur-lex.europa.eu/eli/reg/2024/2847/oj/eng · https://eur-lex.europa.eu/eli/dir/2022/2555/oj/eng · https://eur-lex.europa.eu/eli/reg/2016/679/oj
- **Last reviewed:** 2026-07-14
- **Used for:** `references/observe-run.md`, `references/incident-run.md`, `references/retirement-run.md`
- **Aspects extracted:** Applicability-dependent monitoring, lifecycle, security, resilience, privacy, incident, notification, and retirement concerns. The skill deliberately gives no universal deadline or legal conclusion; current consolidated text, jurisdiction/national transposition, role and system scope, and qualified legal/privacy review remain required.

Decision and control mappings are maintained in the repository plan traceability artifacts.
