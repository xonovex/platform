# Durable Issue and Pull-Request Context Comments

Use a top-level issue comment for selected decision, rationale, assumption, constraint,
or tradeoff context that should remain visible to future human and agent sessions.
Pull-request timeline comments use the same issue-comment endpoints. A context comment
is an explanatory projection; it is not a review verdict, approval, check, merge gate,
or replacement for an authoritative decision record.

## Comment Shape

Give each canonical context version an exact marker:

```markdown
<!-- xonovex-context id=retry-protection-decision version=3 digest=sha256:<digest> -->

## Implementation decision

Decision: Use a transaction marker for retry protection.

Why: The marker commits atomically with the charge.

Alternatives: A deduplication queue adds delivery and operating complexity.

Tradeoffs: Marker retention needs an explicit cleanup policy.

Applies to: `<repository, issue/PR, and subject revision>`
Source: `<authoritative decision reference>`
Supersedes: `<prior context reference>`
Audience: implementers, reviewers, QA, and service owners
Visibility: provider-visible
```

The ID, positive version, and digest come from the canonical workflow context record.
The digest covers its semantic payload, not the provider wrapper. Keep the comment
concise and suitable for repository visibility. Never publish secrets, personal data,
privileged security details, raw chain-of-thought, or unrelated session history.

## Append-Only Reconciliation

`gh issue comment` and `gh pr comment --edit-last` cannot safely identify a logical
context record. Reconcile with the issue-comments REST API:

1. Verify the intended host, repository, authenticated login, issue or PR number, and
   current subject revision.
2. List every comment page for `repos/{owner}/{repo}/issues/{number}/comments`.
3. Select comments by the authenticated author whose first marker has the exact
   context ID and version.
4. No match permits create only after an immediate second list.
5. One matching ID/version with the exact marker and body is a completed retry; return
   it without writing.
6. One matching ID/version with another digest or body is a conflict; block.
7. Multiple matching ID/version comments are a duplicate conflict; block and return
   every native reference.
8. A semantic change increments the context version and creates a new comment naming
   the old comment in `Supersedes`.

Example discovery mechanics:

```bash
NUMBER=482
CONTEXT_ID=retry-protection-decision
CONTEXT_VERSION=3
PREFIX="<!-- xonovex-context id=$CONTEXT_ID version=$CONTEXT_VERSION "
VIEWER_LOGIN=$(gh api user -q '.login')

gh api --paginate --slurp "repos/{owner}/{repo}/issues/$NUMBER/comments" |
  jq --arg prefix "$PREFIX" --arg viewer "$VIEWER_LOGIN" \
    '[.[][] | select(
      .user.login == $viewer and
      ((.body // "") | startswith($prefix))
    ) | {id, html_url, updated_at, body}]'
```

After zero matches are confirmed twice, create from the exact reviewed body:

```bash
gh api --method POST "repos/{owner}/{repo}/issues/$NUMBER/comments" \
  -F body=@context-note.md \
  --jq '{id, html_url, updated_at, body}'
```

Re-read the returned comment by ID and compare the complete body. Do not PATCH or
delete an earlier context version. If the create response is lost, the next attempt
discovers the exact comment and becomes a no-op.

GitHub does not enforce marker uniqueness and comment creation has no idempotency key.
Two concurrent publishers can create duplicates between the final list and POST.
Manual flows need one writer per destination/context version; duplicates block until a
human chooses recovery.

## Issue and Pull-Request Destinations

- For an issue, the destination identity is host + repository + issue number.
- For a pull request, the destination identity is host + repository + PR number and
  the payload's `Applies to` should include the reviewed HEAD SHA.
- Publishing the same context version to both destinations is two independently
  previewed effects with two native comment IDs and URLs.
- A changed PR HEAD does not erase context. Re-evaluate applicability and publish a
  successor when the meaning or applicable revision changes.

## Permissions and Effects

- Fine-grained credentials need `Issues: write` or `Pull requests: write`; use the
  narrowest permission appropriate to the destination and integration.
- Comment creation triggers notifications and may cause secondary rate limiting.
- Preview shows exact body, marker, destination, authenticated identity, current
  matches, and expected create-or-no-op result.
- Apply re-lists immediately before POST and verifies the created or existing comment.
- Return repository, issue/PR number, comment ID, `html_url`, `updated_at`, canonical
  context identity, and authoritative source.
