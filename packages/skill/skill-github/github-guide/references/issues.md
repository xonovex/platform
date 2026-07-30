# GitHub Issues as Workflow Tickets

Use repository issues as provider-native tickets. A durable identity is the repository
plus issue number; preserve the issue node ID, URL, and observed `updatedAt` when
available. Pull requests also appear in issue-list APIs, so exclude entries containing
`pull_request` when raw REST listing is meant to find tickets.

## Read and Discover

Confirm the exact host, repository, and identity before a write:

```bash
gh auth status
gh repo view --json nameWithOwner,url
gh issue view <number-or-url> \
  --json number,id,url,title,body,state,stateReason,assignees,labels,milestone,updatedAt
```

Use `gh issue list --search ... --json ...` for human-oriented filtering and the
paginated Issues REST API when complete machine reconciliation is required. Keep the
repository owner/name with every returned number; `#42` alone is ambiguous outside
one repository.

## Create Once

Create from a reviewed body file so quoting and newlines remain exact:

```bash
gh issue create -R <owner/repo> \
  --title "<title>" --body-file <body.md> \
  --assignee <login> --label <label> --milestone <milestone>
```

Current `gh` versions can also set `--type`, `--parent`, `--blocked-by`, and
`--blocking`. Discover support with `gh issue create --help` before relying on those
flags, especially on GitHub Enterprise Server.

`gh issue create` is not create-or-update. For a retried external create, put a stable
workflow marker in the body:

```markdown
<!-- xonovex-ticket:<stable-idempotency-key> -->
```

Before create and after an unknown result, list candidate issues and compare the exact
marker, repository, intended title, and body:

- no match permits create after an immediate second read;
- one exact match is the existing result;
- one marker with divergent content or multiple matches is a conflict;
- search-index absence alone is not proof of absence; use authoritative paginated API
  reads for retry reconciliation.

GitHub does not enforce marker uniqueness. Concurrent creators can still race, so use
one writer for a key or an external lock and report that limitation.

## Update Metadata Without Collateral Replacement

Use additive and subtractive flags for set-valued metadata:

```bash
gh issue edit <number> -R <owner/repo> \
  --add-assignee <login> \
  --add-label <label> \
  --remove-label <old-workflow-label> \
  --milestone <milestone>
```

Current `gh issue edit` also manages parent, sub-issue, blocking, blocked-by, project,
and issue-type relationships through paired add/remove flags. Use exact issue numbers
or URLs and preserve every returned relationship in the handoff.

`--body`, `--body-file`, and `--title` replace their fields. Read, combine, preview,
and re-read the issue immediately before applying a replacement. If `updatedAt`, body,
or protected metadata changed since preview, block and re-plan. This comparison is
application-level protection, not an atomic compare-and-swap guarantee.

## State and Completion

Issue state is distinct from project kanban status:

```bash
gh issue close <number> -R <owner/repo> --reason completed
gh issue close <number> -R <owner/repo> --reason "not planned"
gh issue close <number> -R <owner/repo> --duplicate-of <number-or-url>
gh issue reopen <number> -R <owner/repo>
```

Preserve `state` and `stateReason`; do not close an issue merely to move its Project
Status to Done, or set Project Status merely to claim the issue is closed. Built-in
Project workflows may synchronize the two, but that automation is an independently
configured effect which must be re-read.

Lock/unlock, pin/unpin, transfer, and delete are separate explicit ticket effects.
Preview the exact repository, issue, permissions, notifications, relationship impact,
and recovery limits. Transfer changes repository-scoped identity; return the
destination issue identity and do not keep using the old repository + number as the
active subject. Delete is destructive and must never be inferred from close.

## Comments and Context

Ordinary discussion uses `gh issue comment`. Durable versioned workflow context on
either an issue or pull request uses the append-only issue-comment procedure in
[context-comments.md](context-comments.md). Never use a review verdict or inline diff
thread as the authoritative ticket context channel.

## Result Handoff

Return:

- host and `owner/repo`;
- issue number, node ID, URL, and observed `updatedAt`;
- title, state, state reason, type, assignees, labels, milestone, and relationships;
- exact created, changed, unchanged, failed, or unknown effects;
- context-comment URLs and revisions when applicable;
- any unavailable host/version feature or concurrency guarantee.
