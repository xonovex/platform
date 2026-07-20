# Composition and CLI

A trusted registry owns executable paths and adapter configuration:

```json
{
  "executors": {
    "review": {
      "adapter": "script-llm",
      "script": {"executable": "/opt/review/facts"},
      "model": {"executable": "/opt/review/evaluate", "timeoutSeconds": 30},
      "capabilities": ["execution:script", "execution:llm"]
    }
  },
  "controls": {
    "approval": {
      "command": {"executable": "/opt/controls/approval"},
      "phases": ["before"],
      "capabilities": ["oversight:approval"]
    }
  },
  "evidenceSinks": {
    "local": {
      "adapter": "jsonl",
      "path": "/var/lib/workflows/events.jsonl",
      "capabilities": ["evidence:local"]
    }
  }
}
```

An invocation selects from that registry without redefining plugins:

```json
{
  "apiVersion": "workflow.xonovex.com/v1",
  "invocationId": "workflow:review:42",
  "trigger": {
    "kind": "ci/github/pull-request",
    "reference": "github:delivery-42"
  },
  "subject": {"reference": "repository:xonovex", "revision": "commit:abc123"},
  "operation": "review",
  "executor": {"plugin": "review"},
  "controls": [{"plugin": "approval", "mode": "observe"}],
  "evidence": [{"plugin": "local", "failure": "ignore"}],
  "requiredCapabilities": ["execution:script"],
  "metadata": {}
}
```

Run or inspect by piping the invocation to the CLI:

```sh
node scripts/workflow-runtime-cli.ts explain registry.json < invocation.json
node scripts/workflow-runtime-cli.ts run registry.json < invocation.json
node scripts/workflow-runtime-cli.ts trigger registry.json template.json < normalized-event.json
```

The command adapters are owned outside the kernel:

- `script` invokes one command.
- `script-llm` invokes a deterministic collection command, then passes its result to a model command.
- `agent-skill` invokes an agent command with a selected workflow skill.

Timeout and output bounds belong to each command definition. The kernel does not impose global retry, token, cost, timeout, or child-depth policy.
