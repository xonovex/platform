# Workflow Result Contracts

## Contents

[Shared semantics](#shared-semantics) · [Canonical result kinds](#canonical-result-kinds) · [Ephemeral handle](#ephemeral-phaseresulthandle) · [Publication and freshness](#publication-and-freshness)

## Shared semantics

Every result communicates, in provider-native form:

- its canonical `kind` and subject;
- source result or evidence references;
- material inputs, scope, and assumptions;
- outcome, status, limitations, and unresolved gaps;
- actor or executor origin and the authority of any decision;
- native reference and revision when published;
- policy/control/profile versions relevant to its interpretation;
- supported follow-up capabilities.

These are semantic requirements, not field names or a required serialization. A provider may represent them with work-item fields, comments, links, commits, checks, approvals, database records, API resources, or another native mechanism.

## Canonical result kinds

| Result kind             | Minimum kind-specific meaning                                                                                     |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Discovery               | Problem or opportunity, affected context and stakeholders, observations, assumptions, unknowns                    |
| Research                | Question and scope, sources and method, findings, limitations, confidence, open questions                         |
| Formulation             | Candidate behavior or requirements, examples and constraints, ambiguities, unresolved decisions                   |
| Experience Design       | Users and tasks, interaction and accessibility decisions, prototypes or representations, validation evidence      |
| Solution Design         | Boundaries, interfaces, data/control flows, qualities, trade-offs, constraints, unresolved risks                  |
| Decision                | Decision, options considered, rationale, consequences, authority, status                                          |
| Planning                | Objective, scope, tasks, dependencies, risks, validation, success criteria, status                                |
| Development             | Exact workspace or subject revision, changes, validation performed, unresolved findings                           |
| Deliverable Publication | Reviewable native candidate reference and immutable revision, intended target, sources, publication status        |
| Review                  | Exact subject revision, reviewer and independence, findings, severity, disposition                                |
| QA                      | Exact subject revision and environment, test scope, results, defects, coverage or assurance gaps                  |
| Assessment              | Criteria/framework and version, applicability, method, evidence, results, gaps, reviewer qualification            |
| Inventory               | Scope and subject revision, component identities and versions, relationships, provenance, gaps                    |
| Data Governance         | Data sets and flows, authorization and lineage, quality/representativeness, gaps, controls                        |
| Acceptance              | Exact deliverable revision and target, accountable actor and authority, evidence, decision, conditions and expiry |
| Integration             | Accepted revision and target, authorization reference and freshness, target-changing action, outcome and rollback |
| Transition              | Source and target context, readiness, migration or transfer, rollback, execution outcome                          |
| Release                 | Integrated revision, release/deployment target, approvals, action, outcome, rollback or recovery reference        |
| Observation             | Observed subject and time window, signals and baselines, findings, anomalies, evidence freshness                  |
| Incident                | Scope, severity, timeline, detected impact, actions, notifications, current state                                 |
| Corrective Action       | Source incident/finding, cause, action and owner, due state, verification and effectiveness                       |
| Retirement              | Subject and scope, authorization, archive/migration/deletion actions, verification, residual risk                 |
| Learning                | Source results, lesson and effectiveness evidence, proposed change, reviewer, promotion/rollback state            |

Profiles may omit or compose capabilities, but they never rename away the underlying result kind or erase its independent publication and pickup boundary. Where a result records an actor, authority, or independence, **agent-governance-guide** defines those terms and states which of them code enforces.

An `Inventory` result may specialize as an SBOM, AIBOM/AI-SBOM, ML-BOM, CBOM, service inventory, or agent-environment inventory. It may describe software, models, datasets, prompts, agents, cryptographic assets, services, installed skills, plugins, hooks, extensions, MCP servers, CI components, policy bundles, tools, and requested/effective permissions. Deterministic sources establish component identities, versions, relationships, and provenance; model enrichment is non-authoritative and limited to sourced descriptions. Each entry retains native identity/version, provenance, relationships, authority zone, observed state, and gaps. This optional inventory is not required to use one serialization; an adapter may link a provider-native SPDX, CycloneDX, package, or platform inventory by opaque reference.

## Ephemeral `PhaseResultHandle`

The runtime handoff contains only:

```text
kind
provider_context
native_reference
native_revision?          # immutable revision or provider-derived version token
source_references[]       # opaque provider-native references
available_capabilities[]  # supported follow-up operations
```

The handle is a logical operation result, not a universal document. Providers may use different names if conformance proves equivalent semantics. After session loss or compaction, resolve the native reference and reconstruct the handle; never depend on conversation memory or infer identity from file names.

## Publication and freshness

- Publish each result independently when its profile requires a pickup boundary.
- Bind review, QA, assessment, acceptance, and privileged operations to exact native subject revisions.
- Re-evaluate evidence when the subject, policy, profile, criteria, evaluator/tool/module, environment, capability matrix, or relevant source changes.
- Keep artifact validity, capability exit, human Acceptance, Integration execution, Release, and cumulative completion as distinct evaluations.
- Treat runtime trace and session identifiers as correlation only.
- Resolve provider operations and version strength through [providers.md](providers.md); do not infer provider capabilities from a native-reference shape.
