# create: Push and open the merge request

**Guideline:** Push the branch yourself, then open the MR with `glab mr create`, capturing the `iid`. glab does NOT push by default, MR creation is NOT idempotent, and the description / labels are replace-only on update. The MR description content is `pull-request-guide`'s craft; the pure-git push and rebase are `git-guide`'s push reference — name them, don't restate them here.

## Push first

`glab mr create` opens an MR from an already-pushed source branch; it does not push on its own. Push with git (see `git-guide`'s push reference), or pass `--push` to make glab push instead:

```bash
git push -u origin HEAD          # see git-guide's push reference
```

## Create

```bash
glab mr create -R group/project \
  --source-branch "$(git branch --show-current)" \
  --target-branch main \
  --title "feat: x" \
  --description "<what/why/how — see pull-request-guide>" \
  --reviewer alice,bob --label backend --yes
```

- `-s` / `--source-branch` defaults to the current branch; `-b` / `--target-branch` defaults to the project default.
- **`--yes` / `-y` is mandatory in scripts** — without it create hangs on an interactive confirmation prompt.
- `--push` makes glab push the branch for you instead of requiring a prior `git push`.
- **Draft:** `--draft` (alias `--wip`) OR a `Draft:` title prefix — equivalent, don't double-apply.

## Metadata

- `--assignee` / `--reviewer` take **usernames** (comma-separated or repeatable; `--reviewer` has no short form).
- `--label` takes label **names**; missing labels are **auto-created**.
- `--milestone`, `--remove-source-branch`, `--squash-before-merge` as needed.

## Issue link and close semantics

- `--related-issue N` links the issue (and reuses its title).
- To **auto-close** an issue on merge, put `Closes #N` (or `Fixes` / `Resolves`) in the **description**; it fires only when the MR merges into the project's **default** branch.
- Naming a branch `1234-foo` auto-appends `Closes #1234` to the MR.

## Idempotency — one open MR per source branch

NOT idempotent. GitLab hard-enforces ONE open MR per source branch; a second `create` returns HTTP 409 "Another open merge request already exists for this source branch: !N". Guard, then attach commits to the existing MR:

```bash
glab mr list --source-branch "$(git branch --show-current)" -R group/project   # find any existing MR
git push                                                                         # new commits attach automatically
glab mr update <iid> --description "<new body>"                                  # or update metadata
```

## Additive-body semantics on update

- `--description` / `description=` **overwrite the whole body** — there is no append. Read-modify-write: `glab mr view <iid> -R group/project -F json | jq -r .description`, edit, re-set.
- Only **labels** are additive: REST `add_labels` / `remove_labels`; setting `labels=` replaces.

## Raw REST equivalent

```bash
glab api --method POST "projects/group%2Fproject/merge_requests" \
  -f source_branch="feat/x" -f target_branch="main" -f title="feat: x" \
  -f description="…" -f labels="backend,bug"
```

- `:id` = numeric project ID or URL-encoded path (`group%2Fproject`). Returns 201.
- **KEY DIFFERENCE from the CLI:** REST takes numeric `reviewer_ids[]` / `assignee_ids[]` arrays (resolve via `GET /users?username=`), NOT usernames; `labels` is a comma-separated string.
- Update: `PUT /projects/:id/merge_requests/:iid` — `description` and `labels` are **replace-only**; only `add_labels` / `remove_labels` are additive.
- Needs `api` scope + Developer role (see [auth.md](auth.md)); prefer a project access token scoped to one project with a short expiry.

**Related:** [review-post.md](review-post.md), [review-resolve.md](review-resolve.md)
