# Agents

## Launch and submit

1. Check `test "${HERDR_ENV:-}" = 1`, load the installed guidance, and inspect `herdr agent` plus the relevant native agent help. Use `herdr status server` if client and server feature support is uncertain.
2. Inspect `herdr agent list` and the intended pane. Reuse an agent only when it matches the requested work. Starting an agent requires a pane at an interactive shell prompt, with no foreground command, editor, or agent.
3. Create the requested topology through the layout operation. Otherwise follow the installed guide's sibling-pane default. Preserve the intended working directory and focus.
4. Start a uniquely named agent in the returned pane ID. Pass native flags after `--`; preserve requested model, reasoning effort, sandbox, and approval settings without granting additional permissions.
5. Successful startup means Herdr detected the expected agent and interactive readiness. If startup reports `agent_not_ready`, inspect the existing agent instead of launching a duplicate.
6. Submit the task with `agent prompt`, then read the resulting state and output. Verify native settings and task activation when the request depends on them.

```bash
herdr agent start writer --kind codex --pane <returned-pane-id> -- <native-agent-arguments>
herdr agent prompt writer "Review the current diff and report findings."
herdr agent get writer
herdr agent read writer --source recent-unwrapped --lines 60
```

Agent names are unique among live agents and match `[a-z][a-z0-9_-]{0,31}`. Agent targets are live names or pane IDs currently hosting agents, not terminal IDs or agent-kind labels.

## Prompt files and native goals

When the user requests a native goal, submit the native goal command plus the full prompt contents through the interactive agent. Confirm that the installed agent supports the command. Herdr submits input but does not create or own the native goal.

For an agent supporting `/goal`, pass the text as one subprocess argument:

```python
from pathlib import Path
import subprocess

prompt = Path("task.md").read_text(encoding="utf-8")
subprocess.run(
    ["herdr", "agent", "prompt", "writer", "/goal " + prompt],
    check=True,
)
```

Keep actual newlines and avoid shell interpolation of prompt text. Read the native confirmation of the active goal; a successful submission alone does not prove it started. Send later corrections as follow-up prompts to the existing task unless the user requests a replacement goal.

## Wait and reconcile

For a short task, `agent prompt --wait --timeout 60000` waits for a settled `idle`, `done`, or `blocked` state. Submit long goals without `--wait`, then inspect progress or use bounded `agent wait` calls. Do not add `--until` unless a particular state is required.

A wait on an already-working agent can finish with its current turn before a queued follow-up runs. After timeout or `agent_prompt_stalled`, inspect `agent get` and `agent read` before retrying. Do not treat those errors as proof that input was never delivered.

If blocked, read the question or approval screen and resolve it within the user's existing authorization. Report any missing decision rather than sending blind keys. Verify completed files or other task outputs before reporting completion.
