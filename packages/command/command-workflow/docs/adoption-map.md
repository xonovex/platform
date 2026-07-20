# Adoption Map

Adopt only the dimensions needed by one workflow. There is no required control stack
and no maturity level that enables behavior.

| Dimension | Examples                                                     | Required by the runtime                |
| --------- | ------------------------------------------------------------ | -------------------------------------- |
| Trigger   | Manual command, harness hook, CI/CD hook, schedule, webhook  | An invocation, not a particular origin |
| Executor  | Script, script plus LLM, agent plus workflow skill           | Exactly one registered plugin          |
| Host      | Local process, agent harness, CI runner, Kubernetes operator | No; the host is outside the runtime    |
| Controls  | Approval, budget, protected target, critique, escalation     | No; select zero or more                |
| Evidence  | JSONL, provider event, audit service                         | No; select zero or more                |
| Maturity  | A1/A2/A3 or an organization-specific model                   | No; derived after composition          |

## Small compositions

| Need                              | Composition                                              |
| --------------------------------- | -------------------------------------------------------- |
| Run a script manually             | Manual trigger + script executor                         |
| Add an LLM transform in CI        | CI trigger + script-LLM executor                         |
| Run an agent skill from a harness | Harness hook + agent-skill executor                      |
| Observe a policy without blocking | Any composition + control in `observe` mode              |
| Block before execution            | Any composition + before-phase control in `enforce` mode |
| Require an audit write            | Evidence sink with failure behavior `fail`               |
| Report team maturity              | Any composition + caller-owned maturity assessment       |

Start with trigger plus executor. Add a host adapter when deployment needs one. Add each
control and evidence sink separately, with its effect stated at the selection site.

## What selection does not imply

- A hook does not imply a control.
- An agent executor does not imply A1, A2, or A3.
- Kubernetes hosting does not imply approvals, escalation, or evidence capture.
- A capability declaration does not enforce anything unless it is explicitly required.
- An observing denial does not block.
- A maturity assessment does not modify the composition.

Use [`workflow-onboard-advise`](../commands/workflow-onboard-advise.md) to draft a minimal
selection and [`workflow-inspect`](../commands/workflow-inspect.md) to explain it before
execution.
