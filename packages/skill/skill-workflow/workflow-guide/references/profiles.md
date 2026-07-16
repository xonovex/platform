# Workflow Profile Composition

## Profile contract

A profile declares semantics, not one config format:

- identity, version, owner, intended scope, and applicability;
- included canonical capabilities and preserved result kinds;
- allowed sequence, concurrency, iteration, backward edges, and composite presentation;
- required inputs, publication boundaries, exit rules, and cumulative completion evidence;
- independent method, provider, workspace, policy, and learning-axis requirements;
- actor, authorization, freshness, and independence requirements;
- required governance evidence and the guarantee an enforcement point must provide;
- failure behavior when a dependency, evidence item, provider, or enforcement capability is unavailable.

## Composition rules

1. Compose requirements by semantic identity, never by copying provider or harness configuration.
2. Strengthening is additive: union required evidence, authority, and guarantees.
3. Resolve incompatible requirements visibly; do not use last-writer-wins for mandatory controls.
4. Weakening requires an authorized exception with scope, owner, rationale, compensating controls, expiry, evidence, and review.
5. Reject a mandatory control when no selected enforcement point can guarantee it.
6. Preserve every included capability's result contract and publication boundary, even under a composite label.
7. Reject invalid edges, missing prerequisites, unsatisfied exact-revision requirements, and incompatible actor rules.

## Default software-delivery presentation

```text
Discover = Discovery + Formulation
Plan     = Research + Planning
Develop  = Development
Deliver  = Deliverable Publication
Review   = Review + QA
Accept   = Acceptance

Explicit post-acceptance capability:
Integrate = Integration

Optional extensions:
Release   = Release
Observe   = Observation
```

This presentation is not the ontology. Another profile may split composites, run Review and QA concurrently, loop from findings to earlier capabilities, omit optional capabilities, or stop after an accepted deliverable.

## Valid profile examples

- A workflow-only local profile explicitly selects a non-file database provider, keeps Discovery and Planning results independent, and has no harness-hook dependency.
- A governed profile requires exact-revision QA and signed evidence, selects two independent enforcement layers that guarantee those requirements, and keeps policy decisions separate from their enforcement records.

## Rejection examples

- A Review composite exposes only one result and discards independent Review and QA evidence.
- A hosted provider was explicitly selected but silently replaced by a local file.
- A runtime trace ID is used as the persistent workflow key.
- A mandatory control names an advisory or unsupported hook as its only enforcement point.
