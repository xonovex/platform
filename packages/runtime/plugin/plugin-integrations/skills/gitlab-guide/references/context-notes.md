# Durable Issue and Merge-Request Context Notes

Use top-level Notes for decision, rationale, assumption, constraint, or tradeoff
context that future human and agent sessions must recover. Issue and merge-request
Notes have separate endpoints but the same append-only identity rules. Do not use an
inline diff discussion, approval, system note, or quick action as the context record.

## Note Shape

```markdown
<!-- xonovex-context id=retry-protection-decision version=3 digest=sha256:<digest> -->

## Implementation decision

Decision: Use a transaction marker for retry protection.

Why: The marker commits atomically with the charge.

Alternatives: A deduplication queue adds delivery and operating complexity.

Tradeoffs: Marker retention needs an explicit cleanup policy.

Applies to: `<project, issue/MR IID, and subject revision>`
Source: `<authoritative decision reference>`
Supersedes: `<prior context reference>`
Audience: implementers, reviewers, QA, and service owners
Visibility: provider-visible
```

The canonical context ID, positive version, and semantic digest form the marker. Keep
the note suitable for project visibility. Never publish secrets, personal data,
privileged security detail, raw chain-of-thought, or unrelated session history.

## Append-Only Reconciliation

Use these collections:

```text
GET /projects/:id/issues/:issue_iid/notes
POST /projects/:id/issues/:issue_iid/notes

GET /projects/:id/merge_requests/:merge_request_iid/notes
POST /projects/:id/merge_requests/:merge_request_iid/notes
```

Before POST:

1. Verify host, project, authenticated username, destination kind/IID, and exact
   subject revision.
2. Paginate all notes and ignore system notes.
3. Select notes authored by the acting identity whose first marker has the exact
   context ID and version.
4. No match permits create only after an immediate second list.
5. One exact marker and body is a completed retry; return it without writing.
6. One same-version note with another digest/body or multiple matches is a conflict.
7. A semantic change increments the version and creates a successor naming the prior
   canonical and native note reference.

Do not PUT or DELETE historical context notes even though the Notes API exposes those
operations. GitLab note updates provide no conditional body precondition. Append-only
successors preserve the decision history and avoid silent overwrite.

After POST, re-read the note by ID and compare its complete body. If the response is
lost, the next attempt reconciles the exact note and becomes a no-op.

GitLab does not enforce marker uniqueness. Concurrent publishers can both pass the
final list and POST duplicates, so manual flows require one writer per
destination/context version. Duplicate discovery blocks and returns every note.

## References and Visibility

- Preserve project full path/ID, issue or MR IID, note ID, author, `updated_at`, and
  `<resource-web-url>#note_<note-id>`.
- Publishing the same context to an issue and MR is two independently previewed
  effects and returns two note identities.
- Treat the note as a projection unless it is explicitly the authoritative decision
  record; retain `Source` in the payload and handoff.
- Use `internal=true` only when the destination supports it and the requested audience
  requires it. An internal note's narrower visibility is part of the returned
  context.
- Creating a note triggers notifications and instance rate limits. Include those
  effects in preview.
