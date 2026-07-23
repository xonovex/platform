# Invocation, Effects, and Execution

Invocation, execution, and authority are runtime context. They do not change the
meaning of a workflow operation.

## Invocation

A person, harness hook, CI/CD event, webhook, or scheduler may submit the same
versioned request contract. The host records the trigger and authorization context;
the public slash command does not expose them as semantic selection flags.

## Execution

A human, deterministic tool, model, or agent may execute a request. Every executor
must honor the same:

- exact resource bindings and revisions;
- method, perspective, and criterion resolution;
- effect mode;
- capability availability;
- operation-result envelope;
- policy and provider boundary.

The runtime derives the required skills, tools, and adapters, then reports their
identities, installed versions, selection reasons, and any unavailable requirement.
An ordinary caller does not choose a generic `--capability`.

Before executing an operation, normalize the request and ask the host composition
runtime to resolve it against every installed plugin root, or one explicit
inventory. A blocked result stops execution; a degraded result remains visible; a
ready or degraded result loads each guide once in its returned dependency-first
`loadOrder`.

## Effect modes

`inspect` reads and reasons without proposing or applying mutations.

`preview` resolves the exact target and produces a reviewable effect set without
applying it.

`apply` performs only the previewed, authorized effects and records each effect as
planned, applied, failed, or unknown.

Execute defaults to `inspect`. Publish and effectful workspace commands default to
`preview`. Read-only and inline-generative operations expose no effect flag.

An `apply` request is not authority. The runtime and provider still enforce
credentials, permissions, approvals, budgets, idempotency, and concurrency.

## Partial and unknown outcomes

The operation result preserves successful, failed, and unknown effects separately.
It never reports whole-call success after a partial write. The retry boundary states
whether retry is safe and includes an idempotency key when the provider supports one.

## Kubernetes callers

A Kubernetes caller submits its execution request through the Kubernetes API. The
[agent operator submission boundary](../../../agent/agent-operator-go/README.md#submission-boundary)
admits and reconciles that request. The operator may enforce runtime policy but does
not invent lifecycle stages or change operation semantics.

## Related guides

- [Command inventory](../README.md)
- [Provider-native resource bindings](references.md)
- [Role lenses](role-lenses.md)
- [Contract migration](migration.md)
- [Operation model](../../../diagram/diagram-agent-workflow/operation-model.png)
