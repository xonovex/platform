# Operation Composition

## Semantic layers

| Layer                  | Contents                                                                                           | Owner                                                                      |
| ---------------------- | -------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| Operation              | Create, Review, Revise, Decide, Execute, Validate, Publish, Abandon, or a workspace transaction    | Caller selects one immediate verb.                                         |
| Semantic selection     | Subject/result kind, method, repeatable perspectives, criteria                                     | Caller selects explicitly; resolver may propose or derive with provenance. |
| Resource binding       | Named input/output slot, inline value or provider-native reference, revision, kind, schema, intent | Each provider adapter interprets only its binding.                         |
| Invocation context     | Trigger, executor, effect mode, authorization context                                              | Host supplies context; policy verifies effects.                            |
| Derived implementation | Skills, tools, adapters, capabilities                                                              | Runtime derives, validates, and reports exact selections.                  |

These are not peer axes. Supplying a trigger never selects a provider; selecting a
perspective never grants a capability; an installed skill never grants authority.

## Public and runtime concepts

Public requests use operation, semantic selections, named resource bindings, desired
output kind, resolution mode, and effect intent. Trigger and executor may be
host-supplied metadata.

The runtime owns skill/capability selection, provider adapter resolution, policy,
authorization, approvals, idempotency, retries, audit, and durable administrative
checkpoints. Expert request documents may pin exact implementation identifiers and
versions, but cannot widen semantics or authority.

## Composition procedure

- [ ] Normalize one exact `WorkflowRequest`.
- [ ] Resolve each named binding and immutable revision independently.
- [ ] Resolve semantics with provenance and mandatory policy.
- [ ] Derive and report implementation selections.
- [ ] Perform only the requested operation and allowed effect mode.
- [ ] Return one inline `OperationResult`.
- [ ] Feed a later operation only an exact prior result or provider binding.

Calls may stand alone, repeat, run in parallel, or appear in policy-defined order.
Examples such as Create → Review → Revise, Execute → Validate, or Review → Publish
are compositions, never mandatory lifecycle stages.

Publish is not a hidden tail of another operation. A host may automate the pair only
when it records two operation contracts and separately verifies Publish authority.
