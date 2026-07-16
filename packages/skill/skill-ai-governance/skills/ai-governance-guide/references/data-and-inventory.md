# Data, Inventory, and Human Oversight

## Inventory the effective AI system

Use deterministic sources to inventory exact identities, versions or digests, relationships, provenance, owners, lifecycle status, permissions, and observed configuration for selected components:

- models, adapters, weights, endpoints, routing and fallback;
- datasets, records, labels, features, retrieval indexes, embeddings, synthetic data, and evaluation sets;
- system/developer/user prompt templates, policies, guardrails, agents, orchestrators, tools, MCP servers, plugins, hooks, and CI modules;
- code, dependencies, infrastructure, providers, regions, secrets-access classes, telemetry, evaluators, and human tasks.

Generate a pinned representation through the selected inventory provider when required. Missing categories are gaps, not proof of absence. A model may enrich a non-authoritative description but never invent component identities, relationships, provenance, versions, or completeness.

## Govern data through its lifecycle

Record authority or lawful basis where required, purpose, collection, source, lineage, transformations, quality criteria, relevance, representativeness, known bias/gaps, labeling, separation, access, consent/notice, retention, residency, external transfer, deletion, and downstream restrictions.

Prevent evaluation contamination: separate training, tuning, validation, acceptance, monitoring, and red-team datasets as required by the profile. Track leakage, memorization, duplicate data, hidden proxies, subgroup coverage, temporal drift, and provider reuse terms.

Prompts, context, tool results, traces, and human feedback are data. Apply classification and minimization before model routing or telemetry export.

## Design effective human oversight

For each oversight task, define the human's identity/role, competence, independence, information, time, workload, accessible interface, authority, available actions, intervention point, fallback, escalation, and evidence. Test that the person can detect the relevant failure, understand uncertainty and limitations, stop or override safely, avoid automation bias, and recover the service.

Do not call a passive dashboard, generic disclaimer, nominal approval, or inaccessible kill switch human oversight.

## Maintain documentation and records

Bind system purpose, architecture, versions, data, prompts, tools, risk decisions, evaluations, instructions, limitations, oversight, changes, incidents, monitoring, and provider-native evidence to the effective system revision. Retention and access follow the data classification and applicable profile; do not centralize sensitive payloads merely for normalization.
