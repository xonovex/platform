# Durable Cross-Session Work

## Administrative checkpoint

Conversation history is not workflow state. Use the provider-owned
[WorkRecord schema](../assets/workflow-work-record.schema.json) for runtime-managed
checkpoints. The
[durable example](../assets/examples/durable-work-record.json) shows exact parent,
input, child, result, skill-version, criterion, and retry bindings.

A checkpoint write is an administrative runtime effect. It is distinct from
publishing a domain result, requires policy authorization, and must not be smuggled
through a core operation's output destination.

## Checkpoint contents

Persist:

- exact work, parent, child, input, and latest-result provider bindings;
- method, perspectives, criteria, provenance, and binding status;
- exact skill identifiers and versions;
- completed effects, receipts, evidence, unresolved questions, and uncertainty;
- safe retry boundary and reconciliation requirement;
- expected record revision for optimistic concurrency.

Provider-native identity remains authoritative. The record has no mandatory global
workflow ID.

The
[child-story continuation request](../assets/examples/child-story-continuation-request.json)
shows a new session pinning child, parent, repository, and prior work-record revisions
before any effect.

## Resume

- [ ] Resolve the exact `WorkRecord` reference and revision.
- [ ] Resolve the exact latest `OperationResult`; do not reconstruct it from chat.
- [ ] Reconcile unknown effects before retry.
- [ ] Revalidate mutable input revisions and policy.
- [ ] Preserve earlier selections unless changed with explicit provenance.
- [ ] Update the checkpoint with an expected-revision precondition.

On a concurrent record change, return blocked with both observed revisions. Never
merge administrative state by prompt order.

Parallel child sessions use separate native child records and exact revisions. A
parent aggregates their bindings and coverage; it does not overwrite child evidence.
