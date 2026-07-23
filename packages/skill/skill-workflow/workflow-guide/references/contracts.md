# Workflow Request and Result Contracts

## Normative assets

- [WorkflowRequest JSON Schema](../assets/workflow-request.schema.json) defines one
  normalized operation request.
- [OperationResult JSON Schema](../assets/workflow-result.schema.json) defines the
  machine-readable result returned by every operation and workspace transaction.
- [WorkRecord JSON Schema](../assets/workflow-work-record.schema.json) defines a
  provider-owned administrative checkpoint for continuation.
- [Multi-provider review request](../assets/examples/multi-provider-review-request.json)
  and
  [matching OperationResult](../assets/examples/review-operation-result.json) show an
  exact request/result pair.
- [Publish request](../assets/examples/publish-request.json) and
  [matching OperationResult](../assets/examples/publish-operation-result.json) show
  the separate authorized persistence effect.

The schemas are the stable machine contract. Prose defines semantic constraints that
JSON Schema cannot express.

## Request normalization

Normalize conversational or command shorthand into `WorkflowRequest` before doing
work:

- Set exactly one operation.
- Bind the subject independently.
- Put supporting resources in named `inputs` slots; a slot can contain one binding or
  an ordered list.
- Set a semantic output kind without an external output destination.
- Resolve command defaults explicitly: assisted selection and the operation's allowed
  effect mode.
- Reject a request document whose operation conflicts with the invoked operation.
- Keep exact expert skill or capability overrides only in
  `implementationOverrides`; ordinary callers request semantics, not mechanisms.
- Put operation-conditional semantic needs in `selection.skillRequirements` and
  applicable preference conventions in `selection.preferenceOverlays`; use empty
  arrays when neither applies.

`destination` is valid only for Publish. A destination hidden in `inputs`, output
metadata, method text, or conversational context never authorizes persistence.

## Required result

Return two sibling representations:

1. A concise human-readable summary.
2. One `OperationResult` object conforming to the normative schema.

When the harness supports structured output, emit the object through that channel.
Otherwise render the exact JSON object in a fenced `json` block headed
`OperationResult`. Never require parsing the human summary.

The result records:

- normalized request and contract version;
- completed, partial, blocked, or failed status;
- resolved method, perspectives, criteria, proposals, skills, preference overlays,
  and capabilities;
- unavailable, incompatible, ambiguous, and conflicting skill selections without
  fabricated implementation identities;
- provenance, reason, confidence, and binding status;
- primary inline output or the exact provider binding returned by Publish;
- evidence, observed or proposed effects, and authorization status;
- unresolved questions and uncertainty;
- retry and reconciliation boundary;
- observed revisions and concurrency state.

Use empty arrays rather than omitting required collections. Do not invent a revision,
receipt, approval, confidence score, or idempotency outcome.

## Operation constraints

| Operation group                                   | Output and effect rule                                                                                                   |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Create, Review, Revise, Decide, Validate, Abandon | `effect.mode` is `inspect`; return the domain result inline and never persist it.                                        |
| Execute                                           | Accept `inspect`, `preview`, or `apply`; return its operation result inline even when apply changes the bounded subject. |
| Publish                                           | Accept `preview` or `apply`; require an exact domain result/artifact and a write-intent destination binding.             |
| Workspace create, merge, cleanup                  | Accept `preview` or `apply`; report every provider effect inline.                                                        |
| Workspace abandon                                 | `effect.mode` is `inspect`; return an inline abandonment record and preserve resources.                                  |

Publish may write a successor to the same logical provider reference as an input only
when the destination carries an exact concurrency precondition. It never rewrites the
pinned source revision.

## Handoffs

Within one conversation, a caller may refer to the exact immediately preceding
`OperationResult`. If more than one candidate exists, require an exact binding or
digest. Across conversations or sessions, require a provider-native binding to a
persisted result or administrative `WorkRecord`.

An automated host may run operation then Publish under predefined policy, but it
must preserve two requests, two results, and two effect decisions.
