# @xonovex/workflow-runtime

Small composition runtime for independent workflow triggers, executors, controls,
evidence sinks, and caller-defined capability assessments.

The runtime has no mandatory controls or maturity model. Callers register plugins and
explicitly select the behavior required for each invocation.

## Modules

- `@xonovex/workflow-runtime/runtime` — invocation schemas, plugin ports, execution, and explanation
- `@xonovex/workflow-runtime/command-runtime` — script, script-plus-model, agent-skill, command-control, and JSONL adapters
- `@xonovex/workflow-runtime/trigger-adapters` — normalized event to invocation adaptation
- `@xonovex/workflow-runtime/maturity` — caller-defined capability assessment

## CLI

```sh
xonovex-workflow explain registry.json < invocation.json
xonovex-workflow run registry.json < invocation.json
xonovex-workflow trigger registry.json template.json < event.json
```
