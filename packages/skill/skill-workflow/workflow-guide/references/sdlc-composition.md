# Policy-Driven SDLC Composition

An SDLC is a policy-defined graph above the stable operations. It selects required
nodes, evidence gates, loops, concurrency, and authority; it does not change operation
semantics or create a universal lifecycle.

Possible graphs include:

- [Small change](../assets/examples/small-change-work-record.json): omit unnecessary
  planning/review verbs; Execute, Validate, and separately integrate a workspace.
- [Regulated release](../assets/examples/regulated-release-work-record.json): repeat
  Review around Revise, run required validations in parallel, Decide readiness, then
  separately authorize Publish.
- [Incident response](../assets/examples/incident-response-work-record.json): start
  Review of evidence and Execute mitigation in parallel, Validate recovery, repeat
  Review for impact, then optionally Create follow-up work.

Omission, repetition, feedback edges, and parallel branches are valid. A graph node
consumes exact bindings or prior `OperationResult` references and emits its own
result. A deterministic workflow host, not model prose, owns required ordering,
transition conditions, retries, and gates.

The following never imply the next effect:

- a ready decision does not release;
- an approved review does not merge;
- a passing validation does not deploy;
- a completed execute result does not publish itself.

Record each transition's policy source and exact input result. If an optional node is
skipped, record the reason; if a mandatory node is unavailable, return blocked.

`WorkRecord.plan` is executable administrative data: nodes reference exact persisted
`WorkflowRequest` documents, edges name completion conditions and result-to-input
bindings, and runtime policy owns transition evaluation. The examples intentionally
use different graphs rather than defining a default lifecycle.
