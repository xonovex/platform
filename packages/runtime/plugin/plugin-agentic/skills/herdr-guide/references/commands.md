# Commands

## Run and inspect

Check the caller environment, load current guidance, and inspect the relevant `herdr pane` help. Use the pane surface for shells, tests, servers, and other ordinary processes. Use the agent surface when the target is a recognized coding agent.

Create a sibling pane with the intended working directory and preserved focus. Read its ID from the creation response and inspect it before sending a command.

```bash
herdr pane split --current --direction right --cwd "$PWD" --no-focus
herdr pane run <returned-pane-id> "npm test"
herdr pane wait-output <returned-pane-id> --match "Tests" --timeout 60000
herdr pane read <returned-pane-id> --source recent-unwrapped --lines 80
```

`pane run` submits command text and Enter together. `wait-output` can match output already present in the selected snapshot; a match alone does not prove successful execution. Check the actual command result or produced artifact. On timeout, read current output before deciding whether another wait or intervention is necessary.

## Select output evidence

- `recent-unwrapped`: prefer for logs and transcripts; joins soft-wrapped rows.
- `visible`: inspect the currently rendered screen or an interactive question.
- `recent`: inspect rendered rows when wrapping matters.
- `detection`: use with `agent read` for the snapshot used to classify the agent.

Limit `--lines` to the evidence needed. Request ANSI output only when colors or terminal styling matter. Larger reads cannot guarantee recovery of alternate-screen application history.

If the final response is unavailable after a larger recent read, ask the existing agent to save it as Markdown and return the path. Read that file on the same machine. Use this fallback only when terminal history is insufficient.

## Preserve running work

Use `agent send-keys` for agent interface controls and `pane send-keys` for intentional raw terminal controls. Inspect state before sending keys. Interrupt or stop only the target whose requested work requires it; do not use server shutdown or process killing to recover a command timeout.
