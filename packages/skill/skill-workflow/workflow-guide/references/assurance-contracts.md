# Inventory and Assurance Contracts

Inventory, Review, QA, and Assessment publish independent results against exact native subject revisions. They may run concurrently and may appear under one profile label, but no composite erases their evidence or exit status.

## Capability boundaries

- **Inventory** identifies components and relationships within a declared exact scope. It may apply to a Deliverable Publication or any other versionable subject.
- **Assessment** evaluates any exact workflow result against a pinned criterion, framework, policy, risk model, or control set.
- **Review** is a deliverable-specific assessor judgment that publishes findings, severity, disposition, reviewer origin, and independence.
- **QA** is deliverable-specific verification in recorded environments, publishing test scope, results, defects, and assurance gaps.

## Evidence composition

| Executor or source                 | Suitable evidence                                                  | Authority limit                                                               |
| ---------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| Deterministic tool, scanner, or CI | Reproducible observations, tests, checks, inventory facts          | Authoritative only for declared tool scope, version, subject, and environment |
| Bounded model                      | Triage, clustering, explanation, candidate findings                | Inference remains non-authoritative and must link source evidence             |
| Independent bounded agent          | Adaptive investigation across permitted tools                      | Result needs validation; independence and authority are explicit              |
| Human or qualified assessor        | Accountable disposition, contextual or role-bound judgment         | Record identity, role, qualification, independence, and scope                 |
| External system                    | Provider-native check, artifact, report, approval, or audit record | Preserve subject revision, evaluator/policy version, outcome, and freshness   |

Static tools, scanners, CI, bounded model review, independent agents, and humans may compose in one capability. Record each origin separately; do not normalize copied logs into a stronger authority class. Identity, role, and independence are workflow evidence fields, not runtime-control selections.

## Freshness binding

Evidence binds to all interpretation-relevant values:

```text
subject native reference and revision or digest
criteria, policy, profile, and control versions
evaluator, scanner, model, module, and ruleset versions
environment identity and relevant configuration versions
actor identity and qualification, with role and independence where the profile binds them
evaluation time, expiry, native evidence references, and limitations
```

Re-evaluate when any required value changes. A changed subject always invalidates revision-sensitive assurance. Policy, evaluator, or environment changes invalidate evidence when the profile or criterion includes them. Never re-label stale, skipped, neutral, cancelled, timed-out, missing, or partial evidence as passing.

## External enforcement and CI

Use provider-native reusable CI modules and protected policy mechanisms through their provider skills. Bind GitHub checks to the exact head revision and stable source/check identity; bind GitLab jobs to the exact commit, component/policy revision, and native job/pipeline identity. Immutable module pins, least privilege, trusted runners, independent required checks, and bypass tests belong to the provider adapter.

A harness hook or command can provide advisory evidence. If a separate runtime composition selects a control, preserve its native evidence reference without treating it as workflow-profile enforcement.

## Adversarial and failure behavior

- Resolve and verify evidence origin, subject binding, digest/signature where selected, and provider state before use.
- Declare concurrency, timeout, retry, idempotency, cancellation, partial-result, and outage behavior. Preserve successful independent evidence when another assessor fails, but keep the aggregate incomplete until required evidence is satisfied or an authorized exception exists.
