---
name: ai-governance-guide
description: "Use when governing an AI or agentic system across applicability, roles, risk and impact, inventories, data, documentation, human oversight, evaluations, monitoring, incidents, or meaningful change. Triggers on AI risk registers, model or dataset provenance, AIBOM/AI-SBOM, bias or representativeness, accuracy and robustness, prohibited or high-risk classification, EU AI Act duties, NIST AI RMF, ISO/IEC 42001, post-market monitoring, or AI assurance evidence, even when the user doesn't say 'AI governance'."
---

# AI Governance and Assurance

Govern AI-system decisions and evidence without treating a framework, model card, evaluation score, or agent assertion as legal applicability or accountable acceptance.

## Essentials

- **Resolve applicability first** — record jurisdiction, dates, intended purpose, role, classification, affected people, exemptions, sources, assumptions, and qualified review, see [references/applicability-and-accountability.md](references/applicability-and-accountability.md)
- **Inventory the effective system** — pin models, data, prompts, agents, tools, evaluations, providers, permissions, versions, provenance, and relationships, see [references/data-and-inventory.md](references/data-and-inventory.md)
- **Manage contextual risk** — identify benefits, affected people, harms, misuse, controls, residual risk, owners, acceptance authority, and change thresholds, see [references/risk-and-impact.md](references/risk-and-impact.md)
- **Govern data and oversight** — document lineage, authority, quality, representativeness, gaps, privacy, human competence, intervention, fallback, and escalation, see [references/data-and-inventory.md](references/data-and-inventory.md)
- **Evaluate exact revisions** — test validity, accuracy, robustness, security, abuse, harmful outcomes, resource controls, and control effectiveness with independent origins, see [references/evaluation-and-monitoring.md](references/evaluation-and-monitoring.md)
- **Monitor meaningful change** — observe production, detect drift, reassess, handle incidents, update safely, and promote learning only through reviewed reversible change, see [references/evaluation-and-monitoring.md](references/evaluation-and-monitoring.md)

## Gotchas

- Framework mappings are contextual and often many-to-many; they do not establish equivalence, conformity, certification, or legal compliance.
- Legal role and risk classification depend on current facts, dates, jurisdiction, intended purpose, deployment context, and value-chain relationships. Require qualified legal review for article-level conclusions.
- A model version is not the AI-system version. Prompts, tools, data, retrieval, policies, orchestration, providers, permissions, and runtime configuration can materially change behavior.
- Aggregate benchmark success can hide harmful subgroup, edge-case, abuse, and operational failures. Preserve scenario-level results and known gaps.
- Human oversight is ineffective when the person lacks time, information, competence, authority, accessible controls, or a tested intervention path.

## Example

```text
System: refund-agent@sha256:abc · provider role unresolved · EU use assumed
Inventory: model + prompt + retrieval index + transcript dataset + refund tool + policies
Risks: incorrect refund, discriminatory denial, prompt injection, data disclosure, excessive agency
Evidence: exact-revision evals + denied-action probes + redaction tests + oversight exercise
Decision: candidate obligations need legal review; release requires residual-risk acceptance
Monitor: drift, harmful outcomes, overrides, incidents, provider changes, and reassessment triggers
```

## Progressive Disclosure

- Read [references/applicability-and-accountability.md](references/applicability-and-accountability.md) - Load when resolving legal or policy applicability, value-chain role, prohibited or high-risk classification, accountability, documentation authority, or crosswalk limits
- Read [references/risk-and-impact.md](references/risk-and-impact.md) - Load when creating an AI risk or impact assessment, identifying affected people and harms, selecting controls, or accepting residual risk
- Read [references/data-and-inventory.md](references/data-and-inventory.md) - Load when inventorying models, datasets, prompts, agents, tools, provenance, permissions, data quality, lineage, privacy, or human oversight
- Read [references/evaluation-and-monitoring.md](references/evaluation-and-monitoring.md) - Load when evaluating exact revisions, red-teaming, monitoring production, handling AI incidents, detecting drift, reassessing change, or promoting learning
