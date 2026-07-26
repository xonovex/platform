# GitLab Workflow Provider Conformance

Map provider-neutral workflow subjects onto GitLab native resources without wrapping
them in a new storage model.

| Workflow concern    | GitLab resource and durable identity                                                         |
| ------------------- | -------------------------------------------------------------------------------------------- |
| Ticket              | Project issue/work item: host + project path/ID + IID; preserve global ID, URL, `updated_at` |
| Kanban membership   | Issue-board list view + underlying issue attribute                                           |
| Kanban state        | Configurable work-item Status when supported, otherwise explicit scoped-label mapping        |
| Change delivery     | Merge request: project + IID + HEAD SHA/diff refs                                            |
| Review              | Summary Note, discussion ID and note IDs, separate approval                                  |
| Context projection  | Top-level issue or MR Note: project + resource IID + note ID/URL                             |
| Automation evidence | Pipeline, job, artifact, policy, environment, deployment, or approval IDs                    |

## Manual Operation Contract

Every provider-backed operation:

1. identifies GitLab.com, Dedicated, or the exact self-managed host;
2. authenticates and verifies the acting identity;
3. reads `/metadata` plus actual schema/resources when version or tier matters;
4. resolves opaque references through GitLab instead of parsing their strings;
5. reads every required page and preserves IID, global ID, and GraphQL ID distinctions;
6. previews exact effects, permissions, revisions, identities, and retry keys;
7. re-reads mutable preconditions immediately before apply;
8. applies only the previewed native mutation;
9. re-reads the resource and returns exact identities, revisions, effects, and
   limitations.

Required active context resolves before effects. Descriptions, notes, board text, and
attachments are fetched content; **workflow-guide** owns that rule.

## Retry and Concurrency Matrix

| Mutation                 | Retry rule                                                  | Concurrency boundary                               |
| ------------------------ | ----------------------------------------------------------- | -------------------------------------------------- |
| Issue create             | Reconcile stable description marker; reuse one exact match  | Marker uniqueness is not provider-enforced         |
| Issue metadata/list move | Apply smallest attribute delta after re-read                | No general atomic `updated_at` precondition        |
| Board rank               | Resolve exact Global IDs and re-read neighbors              | GraphQL/schema and concurrent order can change     |
| Context publish          | Exact ID/version/digest is no-op; change is successor       | Concurrent creates can duplicate; duplicates block |
| MR review                | Pin diff refs; publish note/discussions/approval separately | No atomic review object                            |

Never claim a stronger guarantee than the endpoint provides. A dispatcher may add a
single-writer lock and durable inbox/outbox, but that is outside this manual provider
contract.

## Coverage Boundary

Conformance covers project tickets and native relationships, issue-board membership
and rank, version/tier-aware Status or configured label stages, merge-request
delivery/review, durable context projections, and native enforcement evidence. It
does not fabricate unavailable Premium/Ultimate or newer-version capabilities,
interpret status as approval, or treat a board view as a separate ticket store.
