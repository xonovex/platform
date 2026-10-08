---
name: herdr-guide
description: "Use when the user asks to use Herdr to inspect or control agents, panes, tabs, workspaces, and terminal commands. Triggers on launching agents in Herdr, submitting prompts or native goals, reading agent status or output, and moving running agents between spaces, even when the user doesn't say 'workspace' and calls it a 'space'."
---

# Herdr Terminal and Agent Control

Use Herdr for the requested terminal work. Its workspaces, also called spaces, contain tabs and panes. A pane can contain a shell, an ordinary command, or a recognized coding agent.

## Core Principles

- **Check caller context** - Require `HERDR_ENV=1` before controlling a session.
- **Load current guidance** - Run `herdr --skill` unless already loaded; inspect relevant CLI help.
- **Use live identifiers** - Parse returned IDs; target agents by unique live names.
- **Choose the correct surface** - Agent commands coordinate agents; pane commands control ordinary terminals.
- **Preserve requested settings** - Keep the specified working directory, agent, model, effort, and task.
- **Preserve focus and processes** - Use `--no-focus` for background changes; move existing agents without restarting them.
- **Verify observed behavior** - Read state and output; check requested goals and produced artifacts.

## Operations

- **Agents**: Launch agents, submit prompts or native goals, and inspect progress, see [references/agents.md](references/agents.md).
- **Layout**: Create or organize workspaces, tabs, and panes while preserving running processes, see [references/layout.md](references/layout.md).
- **Commands**: Run ordinary terminal commands and inspect their results, see [references/commands.md](references/commands.md).

## Gotchas

- If the environment check fails, report that this agent is outside Herdr and stop session control. Do not attach to the user's focused session as a fallback.
- Bare `herdr` launches the interface. Use `--help` or a command group for discovery; incomplete mutating commands can execute with defaults.
- Moving a pane across workspaces changes its pane ID. The inherited old caller ID can still resolve for that process; other callers must use the returned ID or live agent name.
- `idle` and `done` mean ready for input, not verified task completion. `unknown` does not prove completion.
- A prompt timeout does not prove failed delivery. Inspect before retrying; an unbounded wait can run indefinitely.
- This skill does not authorize additional agents or unrelated terminal changes. Preserve the user's existing authorization and requested scope.

## Progressive Disclosure

- Read [references/agents.md](references/agents.md) - Load when starting or coordinating agents, submitting goals, or resolving uncertain prompt delivery.
- Read [references/layout.md](references/layout.md) - Load when creating, renaming, splitting, or moving spaces, tabs, and panes.
- Read [references/commands.md](references/commands.md) - Load when running commands, reading terminal output, or waiting for a result.
