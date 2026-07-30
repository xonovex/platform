# create: open a pull request on GitHub

## Guideline

Push the branch (that's `git-guide`'s job, prerequisite: clean branch rebased onto the latest target), then open the PR with `gh pr create`, which auto-pushes the source branch if the remote ref is missing (prompting where to push / offering to fork) and sets reviewers/labels/assignees in one call. The raw `POST /repos/{owner}/{repo}/pulls` does none of that. The PR description content is `pull-request-guide`'s.

## `gh pr create` flags

```bash
gh pr create \
  --base main \                    # -B target branch
  --head feat/x \                  # -H source; OWNER:branch for a cross-fork head
  --title "feat: x" \              # -t
  --body-file pr-body.md \         # -F; use - for stdin, avoids multiline shell-escaping
  --reviewer org/team-slug \       # -r; a TEAM is org/team-slug (a bare slug is read as a username and fails)
  --assignee @me \                 # -a
  --label enhancement \            # -l
  --milestone v2 \                 # -m
  --draft                          # -d; open as a draft
```

- `-f/--fill` auto-fills title+body from commits (`--fill-first` / `--fill-verbose`); `--fill` OVERRIDES any inline `--title`/`--body` for the filled fields.
- A team reviewer must already have repo access; via REST it's `POST .../pulls/{n}/requested_reviewers` with `team_reviewers[]`.

## Draft

`--draft` opens a draft PR (mark ready later with `gh pr ready N`).

## Issue linking and close semantics

Closing keywords live in the PR **body** only. There is no `gh` flag:

- Keywords: `close`/`closes`/`closed`, `fix`/`fixes`/`fixed`, `resolve`/`resolves`/`resolved`.
- `Closes #123` same-repo; `Closes owner/repo#100` cross-repo.
- They auto-close the linked issue **only when the PR merges into the default branch**.

## Additive-body semantics (the edit trap)

`gh pr edit --body` / `--body-file` **replaces** the entire description: it never appends. To preserve existing text, fetch first and recombine:

```bash
body=$(gh pr view 123 --json body -q .body)
gh pr edit 123 --body "$body"$'\n\n## Update\n...'
```

Metadata, by contrast, IS incremental via paired flags: `--add-reviewer`/`--remove-reviewer`, `--add-label`/`--remove-label`, `--add-assignee`, `--add-project`.

## Idempotency

`gh pr create` is NOT create-or-update. It aborts non-zero with "a pull request already exists for the 'X' branch" if an open PR exists. Guard it:

```bash
gh pr create --fill --base main || gh pr view --json url -q .url
```

Only `--web` proceeds anyway; `--dry-run` prints details but "may still push git changes" (not a pure no-op).

## Raw REST equivalent

```bash
gh api --method POST repos/{owner}/{repo}/pulls \
  -f title="feat: x" -f head="feat/x" -f base="main" \
  -f body="..." -F draft=false
```

- Requires `head` + `base` and (`title` OR an issue); optional `body`, `draft`, `head_repo`, `maintainer_can_modify`; cross-fork head is `username:branch`.
- REST does **not** push your branch. The ref must already exist or you get `422 "Head sha can't be blank"`.
- REST cannot set reviewers/assignees/labels in this call (separate endpoints).

### Related

[review-post.md](./review-post.md), [auth.md](./auth.md)
