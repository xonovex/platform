# Xonovex Workflow Commands

Twelve thin commands for explicit workflow operations. Each command owns its
arguments and delegates its procedure to `workflow-guide`.

Core commands:

- `create`, `review`, `revise`, `decide`
- `execute`, `validate`, `publish`, `abandon`

Workspace commands:

- `workspace-create`, `workspace-merge`
- `workspace-abandon`, `workspace-cleanup`

Operations remain separate: a command never implies the next lifecycle step.
Core commands accept an optional exact subject revision. Results stay inline except
for `publish`; `execute` defaults to inspect, while publishing and workspace mutations
default to preview and require explicit apply.
