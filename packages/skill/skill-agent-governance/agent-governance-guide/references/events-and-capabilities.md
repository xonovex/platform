# Semantic Events and Harness Capabilities

## Event intent taxonomy

This section owns the intent vocabulary. A policy binds to a family name, so the bold names below are the stable surface: every other file that enumerates intents is a view of this list and may not add, rename, or drop a family. `scripts/validate-event-intent-vocabulary.mjs` fails the build when a view diverges.

Map these stable intents to native harness or external mechanisms:

- **session** — start, resume, turn completion, and end;
- **prompt** — before submission, after submission, and rejection;
- **model** — before call, after call, failure, and cancellation;
- **tool** — before use, after use, and failure;
- **permission** — request and decision;
- **capability** — before execution, after result, failure, and cancellation;
- **result** — before publication, after publication, and publication failure;
- **configuration** — before change, after change, failure, and observed drift;
- **compaction** — context preservation before and after;
- **subagent** — before launch, after completion, failure, cancellation, and depth limit;
- **workspace** — creation, switch, mutation, and removal;
- **privileged operation** — before authorization, before execution, after execution, failure, and rollback.

The taxonomy expresses intent only. It does not prescribe hook names, files, handler languages, payloads, or configuration precedence.

### Family boundaries

These four boundaries decide what a policy can bind to, so they are settled here rather than per view:

- **permission is a family, not a tool sub-event.** Native permission events carry support, coverage, and blocking that vary independently of tool preflight: a surface can block tool preflight while its permission gate has no effect, and a permission event pair can be observation-only while tool preflight blocks. Nested under tool, a supported preflight would imply a permission guarantee the matrix denies.
- **file mutation is workspace mutation, not a family.** Native file create, save, and delete events report mutation of the active workspace and are recorded there. File-event coverage is not tool-interception coverage; a profile needing both records both mappings.
- **turn completion is a session sub-event, not a result event.** A native main-agent stop event ends a turn and may force another one; it does not publish a result. The result family is publication of a workflow result, and child-agent completion belongs to subagent.
- **result and configuration have no native harness event.** They stay in the taxonomy: an adapter records them as unsupported rather than approximating a nearby event, and a profile that requires them selects external enforcement.

## Capability matrix contract

For every semantic event and native mapping, record:

- harness/product and tested version/date;
- native event name and support state: supported, unsupported, experimental, or unknown;
- handler type and whether it actually executes;
- blocking, advisory, observation, and permission semantics;
- output, context injection, and compaction behavior;
- ordering, matching, concurrency, retry, reentrancy, and timeout behavior;
- organization-managed, project, and user configuration behavior;
- trust boundary, sandbox, permissions, secrets, and data exposure;
- limitations, conformance-probe reference, owner, and review date.

## Guarantee rules

- Unsupported, experimental, or unknown support cannot guarantee a mandatory control.
- A declared denial does not imply concurrent sibling hooks stopped or rolled back.
- Managed configuration is authoritative only within the native guarantee verified by the adapter.
- Profiles fail when a mandatory intent has no selected enforcement point with sufficient support and failure behavior.
- Use independent external enforcement when a harness event can be bypassed, lacks blocking semantics, or cannot cover the required operation.
