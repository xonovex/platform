# GitLab Issues

Use project issues for provider-native tickets. GitLab's `iid` is unique only within
its project; preserve host + project full path or numeric project ID + issue IID.
Also retain the global issue `id`, `web_url`, `references.full`, and observed
`updated_at`.

## Read and Discover

Confirm the intended host and identity before a write:

```bash
glab auth status
glab api "projects/<url-encoded-project-path>"
glab api "projects/<url-encoded-project-path>/issues/<issue-iid>"
```

Project paths in REST URLs must be URL-encoded. Do not interchange the project-scoped
IID, global issue ID, GraphQL Global ID, or work-item reference.

Use `glab issue list/view` for interactive work and the paginated REST API for exact
machine reconciliation. Instance-wide issue listing defaults to the current user's
scope unless `scope=all` is explicit.

## Create Once

`glab issue create` supports title, description, assignees, labels, milestone,
confidentiality, due date, weight, and linked issues. Use `--yes` in non-interactive
flows. Use the Issues REST API when exact type, IDs, or response fields matter:

```bash
glab api --method POST "projects/<project>/issues" \
  -f title="<title>" \
  -F "description=@body.md" \
  -f issue_type=task
```

The current Issues API supports `issue`, `incident`, `test_case`, and `task` on
create. Discover the target host before relying on a type or field.

Neither CLI nor REST create is create-or-update. For an externally retried create,
include a stable marker in the description:

```markdown
<!-- xonovex-ticket:<stable-idempotency-key> -->
```

Before create and after an unknown result, paginate the project's issues and compare
the exact marker, title, description, and intended type. Zero matches permits create
after an immediate second read; one exact match is the existing result; divergence or
multiple matches blocks. GitLab does not enforce marker uniqueness, so concurrent
creators still require one writer or an external lock.

## Update Without Losing Metadata

Use `PUT /projects/:id/issues/:issue_iid`. Prefer `add_labels` and `remove_labels`;
`labels` replaces the complete label set. Assignee arrays, title, description,
milestone, iteration, weight, and dates also require a read/merge/preview/write cycle
when unrelated values must survive.

Immediately before apply, re-read `updated_at` and every protected field. Block if
they differ from preview. GitLab issue updates expose no general atomic
compare-and-swap precondition, so report the remaining race window.

Close and reopen with `state_event=close` or `state_event=reopen`. Do not close an
issue solely to represent a board column unless the target is explicitly the Closed
list or native Status semantics require the corresponding closed category.

## Relationships

Create or inspect cross-project links with the Issue Links API:

```text
POST /projects/:id/issues/:issue_iid/links
  target_project_id=<project>
  target_issue_iid=<iid>
  link_type=relates_to|blocks|is_blocked_by
```

The relationship is bidirectional, but its direction is relative to the source issue.
Resolve both project identities and IIDs, reconcile existing links before create, and
preserve the returned link ID when the host version supplies one.

Use the Work Item GraphQL API for work-item types and widgets that the Issues REST API
does not expose. Introspect the target schema first: work-item fields can be
experimental or version/tier dependent. Do not silently downgrade an epic, objective,
task hierarchy, or custom status to a plain issue.

Time estimate, spent time, reset, and statistics are separate Issues API effects.
Reconcile `/time_stats`, preview the duration delta, use the dedicated
`time_estimate`, `add_spent_time`, `reset_time_estimate`, or `reset_spent_time`
endpoint, then re-read the statistics. Reset is destructive historical/accounting
behavior and requires explicit authority.

Moving or cloning an issue to another project changes or creates project-scoped
identity. Return both source and resulting project/IID references, and never continue
using the old short reference as if it identified the destination.

## Notes and Context

Ordinary issue comments use `glab issue note` or the Issue Notes API. Keep durable
context there rather than in a merge-request review thread, which is scoped to the MR.

## Result Handoff

Return host, project path and ID, issue IID and global ID, `references.full`,
`web_url`, `updated_at`, state, type, assignees, labels, milestone/iteration/dates,
relationships, context-note references, exact effects, and any unavailable tier,
version, permission, or concurrency guarantee.
