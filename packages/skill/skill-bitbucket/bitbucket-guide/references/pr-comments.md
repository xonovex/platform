# Bitbucket Data Center PR comments — post, anchor, edit, blocker tasks, links

## Contents

[Find the PR](#find-the-pr-for-a-branch) · [Top-level comments](#post-a-top-level-comment) · [Inline comments](#post-an-inline-anchored-comment) · [Edit](#edit-a-comment) · [Blocker tasks](#blocker-tasks-merge-gating) · [Resolve a task](#resolve-a-task) · [Reply](#reply-on-a-thread) · [Deep links](#deep-link-between-comments) · [Fields](#severity--state-fields-reference)

These recipes drive a self-hosted Bitbucket Server / Data Center pull request over its REST 1.0 API; they do not apply to Bitbucket Cloud, whose REST 2.0 comment model differs (Cloud uses `content.raw` and `inline`, not the `text` + `anchor` shape below). For _what_ a review comment should say — Conventional Comments labels, blocking versus non-blocking, pairing one summary with line-anchored inline detail — see the code-review-guide skill; this file covers _how_ to deliver that on Bitbucket over REST.

All paths are under `BASE=https://<host>/rest/api/1.0/projects/<KEY>/repos/<repo>` and `PR=$BASE/pull-requests/<id>` (resolve `<host>` / `<KEY>` / `<repo>` from your instance's coordinates). How the token reaches these calls — `curl -n` versus `Authorization: Bearer` — is in [auth.md](auth.md).

Build every JSON body with a serializer (`python3` + `json`), never hand-escaped — comment bodies contain backticks, quotes, and newlines.

## Find the PR for a branch

```bash
curl -s -n "$BASE/pull-requests?state=ALL&at=refs/heads/<branch>&direction=OUTGOING&limit=25"
```

Each value has `id`, `state`, `title`, `fromRef.displayId`, `toRef.displayId`.

## Post a top-level comment

```bash
curl -s -n -X POST -H 'Content-Type: application/json' \
  "$PR/comments" -d '{"text":"<body>"}'
```

Response is JSON carrying the new `id` and `version`. HTTP `201` = created.

## Post an inline (anchored) comment

```json
{
  "text": "<body>",
  "anchor": {
    "diffType": "EFFECTIVE",
    "path": "path/relative/to/repo/root/File.kt",
    "line": 420,
    "lineType": "ADDED",
    "fileType": "TO"
  }
}
```

- `path` — repo-root-relative path of the file.
- `line` — line number in the file version named by `fileType`.
- `fileType` / `lineType`:
  - added line in the new file → `"TO"` + `"ADDED"` (most review comments).
  - removed line in the old file → `"FROM"` + `"REMOVED"`.
  - unchanged line → `"CONTEXT"` (with `"TO"`).
- After posting, check the response `anchor.orphaned`. `false` = attached to the intended diff line; `true` = it did not match, re-check `line` / `lineType`.

Python pattern (reused for every post):

```python
import json, subprocess
def post(pr_id, payload):
    out = subprocess.run(
        ["curl","-s","-n","-X","POST","-H","Content-Type: application/json",
         "-w","\nHTTP %{http_code}","--data", json.dumps(payload),
         f"https://<host>/rest/api/1.0/projects/<KEY>/repos/<repo>/pull-requests/{pr_id}/comments"],
        capture_output=True, text=True)
    return out.stdout  # parse for id, version, anchor.orphaned
```

## Edit a comment

`PUT` needs the comment's **current** `version` (it increments on each edit) or returns `409`.

```bash
curl -s -n -X PUT -H 'Content-Type: application/json' \
  "$PR/comments/<cid>" -d '{"version":<n>,"text":"<new body>"}'
```

Look up `id` + `version` after the fact by paging activities and matching on text:

```bash
curl -s -n "$PR/activities?limit=100"   # values[].comment has id, version, text, anchor, severity, state
```

## Blocker tasks (merge gating)

A comment with `severity: BLOCKER` is a **task** (versus a `NORMAL` comment). Tasks show a checkbox, count in the PR header, and — **if** a repo admin enabled the "all tasks must be resolved before merge" check — block the merge button until ticked off.

Flip an existing comment to a task in place (preserves its `id` and any deep-links):

```bash
curl -s -n -X PUT -H 'Content-Type: application/json' \
  "$PR/comments/<cid>" -d '{"version":<n>,"text":"<same body>","severity":"BLOCKER"}'
```

Verify the open task count and list them:

```bash
curl -s -n "$PR/blocker-comments?count=true"   # {"OPEN": 3}
curl -s -n "$PR/blocker-comments"              # the tasks, with anchors
```

Revert to a normal comment by `PUT`ting `"severity":"NORMAL"` with the new version. You cannot see whether the merge-gating check is enabled via the API, so describe a blocker as "a task" rather than promising it blocks merge.

## Resolve a task

Ticking a task off sets its `state` to `RESOLVED`, same in-place `PUT` + current `version`:

```bash
curl -s -n -X PUT -H 'Content-Type: application/json' \
  "$PR/comments/<cid>" -d '{"version":<n>,"state":"RESOLVED"}'
```

Reopen with `"state":"OPEN"`. Resolving is normally the **author's** action — a reviewer doing a verification pass should confirm the fix in the code first, then resolve. The `id` is preserved, so deep-links keep working.

## Reply on a thread

Reply under an existing comment by POSTing to the comments endpoint with a `parent` id — that threads it beneath the target. Omit `parent` and it is a new top-level comment instead.

```bash
curl -s -n -X POST -H 'Content-Type: application/json' \
  "$PR/comments" -d '{"text":"Resolved in <commit>.","parent":{"id":<cid>}}'
```

The response carries the reply's own `id` and `version` and a `reply: true` flag. Needs `REPO_WRITE`, like any write. Useful after resolving a task, to note where the fix landed.

## Deep-link between comments

Link a summary comment to its inline detail (better than "see comment 3", which the reader cannot resolve):

```
https://<host>/projects/<KEY>/repos/<repo>/pull-requests/<id>/overview?commentId=<commentId>
```

Use it in markdown: `([details](https://<host>/projects/<KEY>/repos/<repo>/pull-requests/<id>/overview?commentId=<commentId>))`. Cross-linking a summary to its inline anchors this way is the delivery half of the code-review-guide's "one summary plus line-anchored detail" convention.

## Severity / state fields (reference)

- `severity`: `NORMAL` (plain comment) or `BLOCKER` (task).
- `state`: `OPEN` / `RESOLVED` / `PENDING`.
- A POST/PUT response echoes both, plus `id`, `version`, `author`, and (for inline) `anchor`.
