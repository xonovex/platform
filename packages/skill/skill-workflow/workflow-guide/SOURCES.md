# Sources

The contract is an Xonovex architectural synthesis. Sources support the recorded concern or constraint; they do not mandate this exact architecture or establish equivalence, certification, or legal compliance.

## S-ISO-12207 — Software lifecycle processes

- **URL:** https://www.iso.org/standard/90219.html
- **Last reviewed:** 2026-07-14
- **Used for:** `references/results.md`, `references/profiles.md`
- **Aspects extracted:** Lifecycle breadth and flexible process application. The public abstract does not support detailed normative claims; those require licensed text.

## S-ISO-15288 — System lifecycle processes

- **URL:** https://www.iso.org/standard/81702.html
- **Last reviewed:** 2026-07-14
- **Used for:** `references/results.md`, `references/profiles.md`
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

- **URLs:** https://spdx.dev/learn/areas-of-interest/ai/ · https://cyclonedx.org/capabilities/mlbom/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/results.md`, `references/providers.md`, `references/inventory-generate.md`, `references/conformance.md`
- **Aspects extracted:** Optional inventories may cover AI/ML and agent/governance components with identities, versions, relationships, and provenance. No one serialization is required by the workflow contract.

## S-SECURE-DEVELOPMENT-ASSURANCE — Secure development and verification

- **URLs:** https://csrc.nist.gov/pubs/sp/800/218/final · https://csrc.nist.gov/pubs/sp/800/218/a/final · https://owasp.org/www-project-application-security-verification-standard/ · https://genai.owasp.org/llm-top-10/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/development-contracts.md`, `references/assurance-contracts.md`, `references/assessment-run.md`, `references/review-run.md`, `references/qa-run.md`
- **Aspects extracted:** Secure-development evidence, versioned verification criteria, bounded model/agent risks, prompt-injection resistance, evidence origin, and adversarial evaluation. These mappings are engineering guidance, not certification or legal compliance.

## S-SUPPLY-CHAIN-EVIDENCE — Provenance and verification

- **URLs:** https://slsa.dev/spec/v1.2/ · https://in-toto.io/ · https://docs.sigstore.dev/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/deliver-publish.md`, `references/inventory-generate.md`, `references/assurance-contracts.md`
- **Aspects extracted:** Exact subject and artifact revisions, producing-step provenance, optional signing/verification, and evidence freshness. Implementations pin the selected specification and trust policy.

## S-WCAG22 — Accessibility assessment

- **URL:** https://www.w3.org/TR/WCAG22/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/assessment-run.md`, `references/qa-run.md`
- **Aspects extracted:** Accessibility assessments pin the applicable criteria, version, and conformance level and preserve automated and human evidence separately.

## S-NATIVE-CI-ENFORCEMENT — Reusable GitHub and GitLab controls

- **URLs:** https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows · https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets · https://docs.gitlab.com/ci/components/ · https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/
- **Last reviewed:** 2026-07-14
- **Used for:** `references/assurance-contracts.md`, `references/qa-run.md`
- **Aspects extracted:** Provider-native reusable modules, immutable revision pinning, stable required-check/job identity, policy enforcement, native evidence, and platform-specific bypass/precedence behavior. Provider skills own detailed mechanics and capability checks.

Decision and control mappings are maintained in the repository plan traceability artifacts.
