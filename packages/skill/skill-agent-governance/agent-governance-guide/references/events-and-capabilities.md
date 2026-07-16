# Semantic Events and Harness Capabilities

## Event intent taxonomy

Map these stable intents to native harness or external mechanisms:

- session start, resume, and end;
- prompt before submission, after submission, and rejection;
- model before call, after call, failure, and cancellation;
- tool before use, after use, failure, and permission request;
- capability before execution, after result, failure, and cancellation;
- result before publication, after publication, and publication failure;
- configuration before change, after change, failure, and observed drift;
- context or compaction before and after;
- subagent before launch, after completion, failure, cancellation, and depth limit;
- workspace creation, switch, mutation, and removal;
- privileged operation before authorization, before execution, after execution, failure, and rollback.

The taxonomy expresses intent only. It does not prescribe hook names, files, handler languages, payloads, or configuration precedence.

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
