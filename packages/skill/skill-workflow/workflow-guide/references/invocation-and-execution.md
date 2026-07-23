# Invocation and Execution

Invocation answers how a request starts. Execution answers who or what performs it.
Neither changes the operation contract.

A person, harness hook, CI/CD event, webhook, or schedule may initiate the same
normalized request. A human, deterministic tool, model, or agent may perform it.
Host-supplied trigger and executor labels are metadata, not semantic selections.

The host may attach an opaque authorization-context reference. Treat it as input to
deterministic policy verification, never as credentials or a self-asserted grant.
Changing trigger, executor, identity, or maturity label cannot:

- turn Review into Validate;
- make Execute publish its result;
- give Decide approval authority;
- select a provider for an unrelated binding;
- skip an effect preview, gate, or concurrency precondition.

Every harness returns the same logical `OperationResult`. A harness without a
structured-output channel renders the object as fenced JSON while keeping the human
summary independent.
