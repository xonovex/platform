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
Core commands accept exact subject revisions and conditionally require them at
provider-native protected boundaries. Results stay inline except for `publish`. Each
command's `--effect` argument documents its own default; the workflow guide's
`references/effects.md` owns the table those defaults come from.

Use `--request <file>` with the workflow guide's Markdown handoff contract for
cross-role traceability, explicit required or preferred capability needs, evidence
bundles, relationships, and retry identity. Review and validation can run in fresh
independent context; externally submitted apply operations use an idempotency key when
their provider supports one.

Use repeatable `--context <text-or-reference>` arguments on Create, Review, Revise,
Decide, Execute, Validate, Abandon, Workspace Create, Workspace Merge, and Workspace
Abandon for smaller handoffs. Context explains intent, constraints, and tradeoffs; it
remains distinct from evidence and authority, and it never carries digests, version
counters, audience taxonomies, or visibility labels. Persist selected context for
future sessions through a separate Publish operation, such as a provider-native issue,
pull-request, or merge-request note.
