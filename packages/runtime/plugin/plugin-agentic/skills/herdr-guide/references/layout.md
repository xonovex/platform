# Layout

## Inspect before changing

Check the caller environment and inspect the live workspace, tab, pane, and agent lists. Use `herdr pane current --current` to resolve the caller after a move; inherited environment IDs can refer to its previous location.

```bash
herdr workspace list
herdr agent list
herdr tab list --workspace <workspace-id>
herdr pane list --workspace <workspace-id>
```

A workspace is a space containing tabs. A tab contains one or more panes. An agent occupies a pane; starting one does not create layout.

## Create the requested layout

Use the requested working directory explicitly and `--no-focus` unless the user asks to switch. Default agent work uses a sibling pane, not a new space or tab, unless the user requests that organization. Inspect `pane layout` and split right for a wide pane or down for a narrow or tall pane.

```bash
herdr pane split --current --direction right --cwd "$PWD" --no-focus
herdr workspace create --cwd "$PWD" --label "Project" --no-focus
herdr tab create --workspace <workspace-id> --cwd "$PWD" --label "Batch 01" --no-focus
```

Choose the needed command rather than executing every example. Creation responses provide `.result.pane` for a split, or `.result.root_pane` for a workspace or tab. Workspace and tab responses also provide `.result.workspace` and `.result.tab` respectively. Read their identifier fields instead of predicting IDs from sidebar order.

For a requested organization of many agents, group spaces by project or task and use short tab labels. Put shared context or batch ranges in the space label. Preserve existing processes while applying the requested organization.

## Move running agents

Read the target agent and pane first; record its working directory and process or native session identity where available. Inspect `herdr pane move --help` before selecting the destination form.

```bash
herdr pane move <live-pane-id> --new-workspace --label "Review" --tab-label "Reviewer" --no-focus
herdr pane move <live-pane-id> --new-tab --workspace <workspace-id> --label "Reviewer" --no-focus
```

Use the move result's `.result.move_result.pane.pane_id` or the agent's live name for subsequent commands. Cross-workspace moves change pane IDs. The previous ID is not a general target, even when inherited caller context still resolves it for the moved process.

Inspect the destination and agent after the move. Confirm that the same process or native session continues in the intended location with the same working directory. Report the resulting space and tab names.

Do not close or restart an agent to move it. Leave unrelated spaces, tabs, and panes intact. A request to organize work does not authorize stopping the server or closing linked workspace groups.
