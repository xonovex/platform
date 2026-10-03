# GitLab Issue Boards as Kanban

An issue board is a view. Its cards are issues, and most lists select issues by one
attribute: label, assignee, milestone, iteration, or configurable work-item Status.
There is no separate REST card object to mutate for a normal cross-list move.

## Discover Host and Board Shape

Self-managed versions and editions differ. Read metadata and the board before choosing
a mechanism:

```bash
glab api metadata
glab api "projects/<project>/boards"
glab api "projects/<project>/boards/<board-id>/lists"
```

For a group board, use the parallel `groups/:id/boards` and
`groups/:id/boards/:board_id/lists` resources and preserve the group plus each card's
own project identity. Group-board creation/configuration and several scopes remain
tier-dependent.

The REST list collection omits the built-in Open and Closed lists. Preserve project
path/ID, board ID, list ID, list type, and the underlying label/user/milestone/
iteration/status identity. A list display name is not a durable ID.

For GraphQL-only list types or position mutations, introspect the target host's
current schema. GraphQL IDs are Global IDs, not REST database IDs.

## Move Between Lists

Translate the source and target list types into the smallest issue mutation:

| Move                       | Native issue effect                           |
| -------------------------- | --------------------------------------------- |
| Open to label              | add target label                              |
| Label A to label B         | remove A, add B                               |
| Label to Open              | remove source label                           |
| Any open list to Closed    | close issue                                   |
| Closed to open list        | reopen, then set the target attribute         |
| Assignee A to assignee B   | replace only the board assignment as intended |
| Milestone/iteration A to B | set the target milestone/iteration            |
| Status A to Status B       | set the work-item status when supported       |

For label boards, use `add_labels` and `remove_labels`; never replace the full label
array. Preserve unrelated labels, including security, component, release, and other
board labels. An issue can legitimately appear in multiple label lists.

Preview every native attribute change and re-read the issue immediately before apply.
After apply, re-read both the issue and target list membership. Board filters and
permissions can keep a correctly mutated issue out of the visible result, so verify
the view as well as the issue.

## Native Status Versus Label Workflow

Configurable work-item Status is a separate, versioned capability. It became generally
available in GitLab 18.4 for issues/tasks on Premium or Ultimate; older and Free hosts
use binary open/closed state plus an explicitly configured label workflow.

1. Read `/metadata` for version/edition and use configured license knowledge.
2. Query the actual board/list and Work Item GraphQL schema for status support.
3. Use native status IDs only when the status widget and mutation exist for that
   subject and identity.
4. Otherwise use the repository's explicit label-to-stage mapping.
5. If neither is configured, block rather than invent `To do`, `Doing`, or `Done`
   labels.

Statuses in done or canceled categories can close the work item. Re-read both status
and state after mutation and report the coupled effect.

## Vertical Ranking and Board Configuration

Cross-list membership and within-list rank are different operations. Use the
GraphQL `issueMoveList` mutation only for an explicit rank/move request after resolving
the current board/list/issue Global IDs and the target host's input schema. Preserve
the returned `positionInList` and re-read neighbors. Do not call it merely to add or
remove a label.

Board-list configuration uses:

- `POST /projects/:id/boards/:board_id/lists` for supported label, assignee,
  milestone, or iteration lists;
- `PUT .../lists/:list_id` for list position;
- `DELETE .../lists/:list_id` to remove the view, not its issue attributes;
- GraphQL for status lists or other fields only when exposed by the target schema.

Creating, deleting, or reordering lists is a separate board-configuration effect from
moving a ticket.

## Result Handoff

Return host version/edition and known tier, project and board IDs, source and target
list IDs/types, issue project + IID/global ID, exact attribute delta, observed
`updated_at`, resulting state/status/list membership/position, and every degraded or
unsupported capability.
