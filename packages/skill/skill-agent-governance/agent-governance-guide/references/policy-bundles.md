# Policy Bundles and Enforcement Mapping

## Bundle semantic controls, not one engine

A policy bundle is a versioned selection of semantic controls and their ownership. It records:

- identity/version, owner, authority zone, source/profile versions, applicability, subjects, actors, and freshness;
- domain owner, decision inputs/outcomes, explanation, evidence requests, exceptions, and failure behavior;
- enforcement intent and required guarantees, supported adapter kinds, limitations, and native references;
- data/telemetry requirements, conflicts, dependencies, compatibility, update/rollback, and retirement.

The domain skill owns the substantive guidance: **security-assurance-guide**, **accessibility-guide**, **ai-governance-guide**, and **reliability-guide** when installed. Privacy, supply-chain, data, cost, and regulated bundles may use another installed owner or organization policy. This governance skill owns composition, authority, enforcement, evidence, exceptions, and operational behavior; it does not duplicate each domain.

## Keep providers interchangeable

Start with deterministic rules for explicit facts and bounded logic. Use provider-native policy when the provider owns authoritative state or enforcement. Add an OPA/Rego adapter when centralized or declarative policy adds value. No profile may require one engine merely because fixtures demonstrate it.

Adapters conform when the same versioned facts produce the same semantic decision, explanation, evidence requirements, and exception handling. Preserve native decision references and test historical replay. Provider output never becomes the cross-platform contract.

## Map decisions to adequate enforcement

For each mandatory control:

1. resolve applicability, policy/profile version, actor, authority zone, subject, and required outcome;
2. identify an enforcement point with sufficient coverage, blocking, ordering, integrity, availability, and bypass resistance;
3. declare `fail-closed`, `fail-visible`, or `advisory` for policy and enforcement dependency failures;
4. return distinct decision, enforcement, evidence, exception, and native configuration references;
5. test intended allow/deny, missing capability, stale policy, outage, bypass, duplicate/concurrent execution, and rollback.

A hook may advise or block one harness event but cannot claim organization-wide enforcement. Use repository, CI, deployment, admission, identity, provider, or accountable human controls where the required guarantee exceeds harness capability.

Telemetry and advisory modules fail visibly or degrade by default; their outage does not automatically block unrelated work. A selected profile may require fresh telemetry evidence for a specific privileged action, but that is an explicit policy with adequate enforcement, not an observability default.

## Representative profiles

| Profile              | Typical composition                                                                       | Required proof                                                              |
| -------------------- | ----------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| Lightweight          | knowledge plus deterministic local checks; advisory external evidence                     | no mandatory guarantee is claimed; gaps visible                             |
| Secure               | security/supply-chain policy plus protected repository, CI, identity, or deployment gates | intended allow/deny, provenance, outage, exception, rollback                |
| AI                   | AI risk/data/inventory/evaluation plus security, privacy, oversight, and monitoring       | exact effective-system version, evaluation, denied actions, drift, incident |
| Regulated            | resolved applicability plus qualified review and selected legal/control obligations       | source versions, reviewer authority, gaps, independent enforcement/evidence |
| Organization-managed | managed policy/modules plus non-weakening project/user/session composition                | precedence, conflict, permission expansion, drift, emergency disable        |

Profiles include only applicable bundles. Crosswalks record relationships and gaps; they never establish equivalence, certification, conformity, or legal compliance.
