# GitHub Projects as Kanban

GitHub Projects stores issues and pull requests as project items. A card's workflow
state is normally a project `Status` field value, not the underlying issue's
open/closed state. Preserve all four identities: project owner and number, project
node ID, project item ID, and content issue or pull-request URL/node ID.

## Discover Before Writing

Projects commands require the `project` token scope. Confirm the owner explicitly;
projects can belong to a user or organization.

```bash
gh project view <number> --owner <owner> --format json
gh project field-list <number> --owner <owner> --format json
gh project item-list <number> --owner <owner> --limit 100 --format json
```

Resolve display names to native IDs on every run:

- project number plus owner to project node ID;
- content URL/node ID to existing project item ID;
- `Status` or another field name to field ID and data type;
- single-select label to option ID;
- iteration title to iteration ID.

Names are user-editable and are never substitutes for IDs in a saved handoff. Paginate
or raise `--limit` until the intended item and field set are known; a partial first
page cannot prove absence.

## Add Once

Before adding, search the authoritative item list for the exact content URL or node
ID. Reuse one existing item, block duplicates, and add only when absent:

```bash
gh project item-add <number> --owner <owner> \
  --url <issue-or-pull-request-url> --format json
```

The GraphQL equivalent is `addProjectV2ItemById`. Preserve the returned project item
ID. Do not create a draft issue when the subject is an existing repository issue or
pull request.

## Set or Clear Fields

Update one field per invocation:

```bash
gh project item-edit \
  --id <project-item-id> \
  --project-id <project-node-id> \
  --field-id <status-field-id> \
  --single-select-option-id <in-progress-option-id>
```

Use the matching typed flag for text, number, date, iteration, or single-select
fields; use `--clear` to remove a value. Never send a display label where an option ID
is required.

For a manual preview/apply transaction:

1. preview the exact project, item, field, old value, and target value;
2. immediately before apply, re-read that item and field;
3. block if the item, field definition, option set, or observed old value differs;
4. perform the one-field mutation;
5. re-read and return the actual value plus its timeline/effect reference when
   available.

Projects field mutations do not provide a general compare-and-swap precondition. The
pre-write read detects many races but cannot make the write atomic. Built-in
automations may also change `Status`; report this limitation rather than claiming
strong concurrency control.

## Board Operations

- Move a card between kanban columns by changing its resolved `Status` option.
- Add/remove issue metadata through the Issues operation, not through project fields.
- Archive a completed item with `gh project item-archive`; archive is not deletion and
  does not close its issue.
- Remove an item with `gh project item-delete` only under an explicit destructive
  effect; it does not delete the issue.
- Create draft items only when the user explicitly wants a project-local draft rather
  than a repository ticket.
- Configure built-in workflows separately. Their automatic add, status, archive, or
  close effects must be included in preview, verification, and drift checks.

## Compatibility

Discover `gh project --help` and the target GraphQL schema on GitHub Enterprise Server
before promising a command or field. For example, advanced `gh project item-list
--query` depends on host support. If Projects or a field type is unavailable, keep the
ticket operation usable, report degraded kanban coverage, and do not invent a label or
issue-state mapping unless the caller explicitly configured it.

## Result Handoff

Return the host, owner, project number and node ID, item ID, content URL/node ID, field
ID/type, selected option or value ID, observed previous and resulting values, and
every applied or unknown effect. Keep the underlying issue identity and state
separate.
