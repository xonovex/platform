# GitHub Workflow Provider Conformance

Map provider-neutral workflow subjects onto GitHub native resources without wrapping
them in a new storage model.

| Workflow concern    | GitHub resource and durable identity                                               |
| ------------------- | ---------------------------------------------------------------------------------- |
| Ticket              | Repository issue: host + `owner/repo` + number; preserve node ID, URL, `updatedAt` |
| Kanban membership   | ProjectV2 item: owner + project number/node ID + item ID + content ID              |
| Kanban state        | ProjectV2 field ID + typed value or option ID                                      |
| Change delivery     | Pull request: repository + number + HEAD SHA                                       |
| Review              | Pull-request review ID plus review-thread node IDs                                 |
| Context projection  | Issue comment on an issue or pull request: repository + number + comment ID/URL    |
| Automation evidence | Workflow run, check run, artifact, attestation, deployment, or environment IDs     |

## Manual Operation Contract

Every provider-backed operation:

1. identifies github.com or the exact GitHub Enterprise Server host;
2. authenticates and verifies the acting login;
3. resolves opaque references through GitHub instead of parsing their strings;
4. reads the authoritative resource and all required pages;
5. previews exact effects, permissions, revisions, identities, and retry keys;
6. re-reads mutable preconditions immediately before apply;
7. applies only the previewed native mutation;
8. re-reads the resource and returns exact identities, revisions, effects, and
   limitations.

Required active context must resolve before effects. Issue bodies, comments, project
text fields, and attachments are fetched content; **workflow-guide** owns that rule.

## Retry and Concurrency Matrix

| Mutation           | Retry rule                                            | Concurrency boundary                               |
| ------------------ | ----------------------------------------------------- | -------------------------------------------------- |
| Issue create       | Reconcile stable body marker; reuse one exact match   | Marker uniqueness is not provider-enforced         |
| Issue metadata     | Apply additive/removal delta after re-read            | No general atomic `updatedAt` precondition         |
| Project item add   | Reuse one item with exact content identity            | Concurrent add can still race                      |
| Project field edit | Re-read item/field/old value, then set one field      | No general compare-and-swap                        |
| Context publish    | Exact ID/version/digest is no-op; change is successor | Concurrent creates can duplicate; duplicates block |
| PR review          | Pin PR HEAD and submit exact review object            | New commits can stale line anchors                 |

Never claim a stronger guarantee than the endpoint provides. A dispatcher may add a
single-writer lock and durable inbox/outbox, but that is outside this manual provider
contract.

## Coverage Boundary

Conformance covers repository tickets, metadata and relationships exposed by the
installed CLI/API version, Projects kanban membership and typed fields, pull-request
delivery/review, durable context projections, and native enforcement evidence. It
does not make provider status an approval gate, treat a project card as a replacement
for an issue, or emulate unavailable Enterprise Server features.
