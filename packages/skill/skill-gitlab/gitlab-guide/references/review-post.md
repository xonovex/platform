# review-post: Publish a structured review

## Guideline

A GitLab review has NO single object — assemble three parts and publish each immediately: a summary note, one position-anchored discussion per inline finding, and a separate approve/withhold verdict. This realizes `code-review-guide`'s findings on GitLab — what each comment says (Conventional Comments labels, blocking vs non-blocking, summary-plus-inline structure) is that skill's craft; this file is only how to deliver it.

## (1) Summary — a plain note

The review's prose is a non-anchored note. `/notes` is unthreaded; use it for the summary, not for anything that must gate merge.

```bash
glab mr note <iid> -R group/project -m "## Review summary
Two blocking issues inline; withholding approval until addressed."
# raw: POST /projects/:id/merge_requests/:iid/notes
```

## (2) Inline comments — position-anchored discussions

Each inline finding is its own `POST /projects/:id/merge_requests/:iid/discussions` with a **position object**. Easiest via glab, which resolves the SHAs for you:

```bash
glab mr note create <iid> -R group/project --file src/x.go --line 42 \
  -m "**issue (blocking):** unchecked error"
# --old-line N      anchor a removed line
# --line 10:15      anchor a range
# (omit both)       file-level comment
# --reply <discussionID>   append to an existing thread
# --unique          idempotent for bots
```

### The exact position model (raw REST)

`POST .../discussions` with `position[...]`:

- `position[position_type]=text`
- **THREE SHAs, all MANDATORY and exact:** `position[base_sha]`, `position[start_sha]`, `position[head_sha]`. Fetch them from `.../merge_requests/:iid/versions` or the MR object's `diff_refs` (`base_sha` / `head_sha` / `start_sha`). **Re-fetch after every push** — they change.
- **`position[old_path]` AND `position[new_path]` are both required** (set them equal if the file was not renamed).
- **Line key is conditional:**
  - ADDED / green line → set `new_line`, OMIT `old_line`.
  - REMOVED / red line → set `old_line`, OMIT `new_line`.
  - UNCHANGED / context line → set BOTH.

```bash
refs=$(glab api "projects/group%2Fproject/merge_requests/<iid>" | jq .diff_refs)
glab api --method POST "projects/group%2Fproject/merge_requests/<iid>/discussions" \
  -f body="**issue (blocking):** unchecked error" \
  -f position[position_type]=text \
  -f position[base_sha]="$(jq -r .base_sha <<<"$refs")" \
  -f position[start_sha]="$(jq -r .start_sha <<<"$refs")" \
  -f position[head_sha]="$(jq -r .head_sha <<<"$refs")" \
  -f position[old_path]="src/x.go" -f position[new_path]="src/x.go" \
  -f position[new_line]=42
```

### CRITICAL: verify it anchored

A wrong / stale SHA or the wrong line key does **NOT** error — GitLab returns **201** and silently downgrades the `DiffNote` to a floating `Note` on the Overview tab. Always check the response before claiming success:

```bash
# pipe the create response through jq:
... | jq -e '.notes[0].type == "DiffNote"'   # exits non-zero if it downgraded to a plain Note
```

## (3) Blocking / verdict

GitLab has **no `REQUEST_CHANGES` over REST or glab** — the real reviewer "Request changes" state is Premium/Ultimate and GraphQL/UI-only. Use a portable gate instead:

- **(a) Withhold approval** against approval rules. Approving clears the gate:
  ```bash
  glab mr approve <iid> -R group/project          # POST .../approve (optional &sha=<head>, 409 if stale)
  ```
  To block, simply do NOT approve.
- **(b) Leave resolvable threads unresolved** with the project setting `only_allow_merge_if_all_discussions_are_resolved=true` — the most reliable cross-tier gate (default discussions are resolvable). See [review-resolve.md](review-resolve.md) for the gating check.

## Deep-link

MR notes carry **no `web_url`** — build it yourself from the returned note `id`:

```
<mr_url>#note_<note_id>
```

## Scope

All writes here need `api` scope + Developer role (or MR author); see [auth.md](auth.md). The atomic "Submit review with summary" exists only in the UI / GraphQL — over REST every piece publishes immediately.

### Related

[review-resolve.md](review-resolve.md), [create.md](create.md)
