# Governance Composition

Record the effective composition before execution. The record should identify:

- the caller-selected operation and exact subject and revision;
- the normalized trigger kind and native event reference;
- one selected executor capability and its owner;
- zero or more controls, each with phase, `observe` or `enforce` mode, enforcement point, and failure behavior;
- zero or more evidence sinks, each with durability, sensitivity, access, retention, and sink-failure behavior;
- required capabilities and the evidence used to establish their availability;
- host and provider boundaries, including where credentials and privileged side effects remain;
- deadlines, retry and output bounds owned by the selected adapters; and
- unresolved assumptions, unsupported capabilities, and the authority accepting each gap.

## Composition procedure

1. Accept the caller-selected operation, exact subject, revision, owner, and applicable
   policy. Stop on a missing operation instead of selecting or sequencing one.
2. Normalize the native trigger without granting it authority to select its own executor or controls.
3. Select exactly one executor capability from trusted configuration.
4. Add controls independently and state whether each observes or enforces before or after execution.
5. Add evidence sinks independently and define whether their failure is advisory or fatal.
6. Compare required capabilities with verified effective capabilities; fail visibly on missing mandatory capabilities.
7. Explain the resulting behavior, especially every denial point and every side effect that cannot be undone.
8. Bind the approved composition to the exact invocation without copying untrusted event data into privileged configuration.

## Executor capabilities

- A deterministic command performs the caller-selected bounded operation.
- Command-plus-model execution gathers authoritative facts deterministically, then applies a bounded model transformation or evaluation.
- An agent with a workflow skill performs adaptive multi-step work within explicit authority, tool, cost, time, and child-depth limits.

The governance composition does not supply or reorder an operation, executor, host,
retry policy, token budget, or protected enforcement point. Those are independently
selected or implementation-owned capabilities whose effective behavior must be
verified separately.
