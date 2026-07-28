# Composing Operations Across Roles

Order work as explicit operations with handoffs at the cold boundaries between them.
The ordering is a playbook, not a lifecycle: no command implies the next one. Select
the guides each step uses from their installed routing descriptions, as
[capability-selection.md](capability-selection.md) defines, rather than from a fixed
inventory.

## Shapes

These orderings recur across domains. Read them as shapes, not as a required sequence.

- **Scoping** — Decide settles the open questions, then a planning guide produces the
  plan itself. Slicing an increment is the same shape: Decide the split, then Validate
  the agreed examples independently.
- **Building** — Execute realizes an antecedent against the implementation guides;
  Validate evaluates the acceptance criteria as a separate invocation.
- **Authoring** — Create produces the artifact, Revise applies feedback, Publish
  persists it. A manifest, an article, and a catalog entry share this shape.
- **Delivering** — Review the change, then Publish it through the provider and release
  guides. Review never approves and never publishes.
- **Failing** — Execute reproduces the problem; when the repro does not hold, Abandon
  records the partial state and the retry boundary instead of guessing.
- **Isolating** — workspace create to isolate, merge to integrate, cleanup as its own
  later step.

## Invariants

1. Pin every provider-native artifact to its exact revision when the provider
   exposes one.
2. Run independent reviews or validations as separate invocations against the same
   subject revision.
3. Preview protected effects before apply.
4. Keep integration and cleanup separate.
5. Carry only the context the receiving role needs, in the shape
   [handoffs.md](handoffs.md) defines.

Authority, evidence, and fetched-content rules live in [governance.md](governance.md).

## Provider Mapping

Use one provider skill explicitly; do not infer it from a reference's shape.

| Concern        | GitHub                                      | GitLab                                               |
| -------------- | ------------------------------------------- | ---------------------------------------------------- |
| Ticket         | Repository issue: repo + number/node ID/URL | Project issue/work item: project + IID/global ID/URL |
| Kanban card    | ProjectV2 item with its own item ID         | Issue card rendered from the issue                   |
| Workflow stage | Project field/option ID, usually Status     | List-defining issue attribute                        |
| Change         | Pull request + HEAD SHA                     | Merge request + HEAD SHA/diff refs                   |
| Context        | Append-only comment on issue or PR          | Append-only Note on issue or MR                      |

Author a ticket body, comment, or workflow-stage change with Create, then persist it
with Publish: preview the exact write, then apply. Execute covers provider work an
antecedent already specified, such as carrying out accepted review feedback on a
change. Assignments, access control, approvals, status transitions, notifications,
and scheduling stay owned by the provider.
