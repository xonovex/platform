# Invocation and Execution

Invocation answers how a command call starts. Execution answers who or what performs
it. They are independent of the selected operation and of each other.

## Invocation mechanisms

The same command contract can be invoked by:

- a person issuing a manual command;
- an agent harness hook;
- a CI/CD hook;
- a webhook handler; or
- a scheduler.

Each mechanism supplies the same subject and optional kind, perspective, criteria,
method, capability, provider, reference, revision, and result selections. The trigger
does not add a lifecycle stage, change an operation, or authorize its side effects.

## Executors

A human, deterministic script, LLM, or agent can execute the call. The executor must
honor the same argument contract, capability availability checks, provider boundary,
and explicit side-effect rules. Changing executor does not turn `review` into
`validate`, make `execute` publish, or give `decide` authority.

Harnesses and providers may attach labels such as A1, A2, or A3 as optional executor
metadata. The command package does not define those labels, pass them as command modes,
derive permissions from them, or map them to required stages. Any enforcement belongs
to the invoking harness or provider and leaves operation semantics unchanged.

## Examples

A manual caller and a CI hook can request the same validation:

```text
/xonovex-workflow:validate <subject> --criteria <criteria>
```

A webhook handler and a scheduler can request the same bounded execution:

```text
/xonovex-workflow:execute <subject> --criteria <criteria>
```

Their authentication, retry, scheduling, and delivery mechanisms differ. The command
contract does not.

## Kubernetes callers

A Kubernetes-based manual tool, harness hook, CI/CD system, webhook handler, or
scheduler submits an `AgentRun` directly through the Kubernetes API. The
[agent operator submission boundary](../../../agent/agent-operator-go/README.md#submission-boundary)
admits and reconciles that execution request. The operator is not a workflow runtime
and does not interpret trigger-specific workflow stages.

## Related guides

- [Command inventory](../README.md)
- [Role lenses](role-lenses.md)
- [Provider-native references](references.md)
- [Operation model](../../../diagram/diagram-agent-workflow/operation-model.png)
