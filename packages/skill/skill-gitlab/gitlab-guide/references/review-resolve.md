# review-resolve: List, match, and resolve threads

## Guideline

Resolution is a plain REST PUT on the discussion — no GraphQL needed on GitLab. List the threads, match a finding to a thread by its `id` (never by line number), resolve, and reply in-thread. Resolution gates merge only when the project opts in.

## List and match

Enumerate the threads and match each finding to a thread by its `id` (a 40-char hex string) or, for diff threads, by `notes[0].position.new_path` + `new_line` — **never by line number alone**:

```bash
glab api "projects/group%2Fproject/merge_requests/<iid>/discussions" \
  | jq -r '.[] | {id, path: .notes[0].position.new_path, line: .notes[0].position.new_line, resolved: .notes[0].resolved}'
```

The resolve and reply endpoints accept only ids. A `note_id` can stand in for resolving (the parent discussion is looked up), but **reply needs the `discussion_id`**.

## Resolve mechanism (REST is enough)

```bash
glab api --method PUT \
  "projects/group%2Fproject/merge_requests/<iid>/discussions/<discussion_id>?resolved=true"
# resolved=false reopens
```

- This is the canonical mechanism on GitLab — **unlike GitHub, no GraphQL is required.** (A GraphQL `discussionToggleResolve(input:{id,resolve})` mutation also exists, but it needs a Global ID `gid://gitlab/Discussion/<hash>`, NOT interchangeable with the REST 40-char hex `discussion_id`.)
- **Resolvability is thread-vs-single-comment, not diff-vs-general.** Any MR thread started as a _thread_ — including general threads on the Overview tab — is resolvable; a single comment is not. It is not limited to diff/commit threads.
- Resolve a single note within a thread: `PUT .../discussions/:discussion_id/notes/:note_id?resolved=true` (`resolved` and `body` are mutually exclusive).
- `glab mr note resolve <iid> <discussion_id|8char_prefix|noteId>` exists but is **EXPERIMENTAL** (may change or be removed) — prefer the raw `glab api --method PUT` above for stable automation.

## Reply in-thread

```bash
glab api --method POST \
  "projects/group%2Fproject/merge_requests/<iid>/discussions/<discussion_id>/notes" \
  -f body="Fixed in abc1234."
```

Use this, not `glab mr note -m` — that posts a NEW standalone unthreaded comment that never blocks and never attaches to the thread.

## Merge-gating effect

Thread resolution gates merge ONLY when the project (or group) enables `only_allow_merge_if_all_discussions_are_resolved` ("All threads must be resolved"). A group-level enable LOCKS the project setting. Only **threads** gate — single comments never do. Check it:

```bash
glab api "projects/group%2Fproject/merge_requests/<iid>" | jq -r .detailed_merge_status
# "discussions_not_resolved" -> unresolved threads are blocking the merge
```

Resolving needs `api` scope + Developer role (or the MR author); a read-only / Reporter identity can list but not resolve (see [auth.md](auth.md)).

## Gotcha

Beware the project "automatically resolve threads when they become outdated" setting: a push that moves the anchored lines silently flips those threads to resolved — re-check `detailed_merge_status` after a push rather than assuming an unresolved thread stayed blocking.

### Related

[review-post.md](review-post.md)
