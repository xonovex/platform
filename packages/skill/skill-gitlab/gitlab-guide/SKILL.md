---
name: gitlab-guide
description: "Use when delivering a merge request and its review on GitLab (gitlab.com, self-managed, or Dedicated) with the official glab CLI — opening an MR with `glab mr create`, posting a structured review (summary note + line-anchored inline discussions), resolving review threads, approving/withholding approval, detecting the GitLab host from the git remote, and choosing token type and scopes (the broad `api` scope for any write, `read_api` for reads). Triggers on a gitlab.com or self-managed GitLab remote, `glab`, a merge request / MR, opening an MR against main, posting an inline MR comment, a position object with base/start/head SHAs, resolving a discussion / thread, approving an MR, or a GitLab token scope question — even when the user doesn't say 'glab'."
---

# GitLab MR & review delivery — quick reference

How to realize a merge request and its review on GitLab from the shell with the official `glab` CLI (and raw REST where glab has no flag). This is the host-DELIVERY tier and nothing more: the review craft (what a finding says, Conventional Comments labels, blocking vs non-blocking) is `code-review-guide`'s; the MR description craft (what/why/how, sizing, test notes) is `pull-request-guide`'s; the local-git push and rebase are `git-guide`'s push reference. This skill only realizes those on GitLab — the wiring, scopes, anchors, and gating.

The one thing to internalize: **a GitLab "review" is not one object — you assemble it from a plain summary note, individual position-anchored discussion threads, and a separate approve/withhold signal, and almost every write needs the coarse, all-or-nothing `api` scope plus a sufficient role, so plan around the missing atomic-review endpoint and the absent write-only scope.**

When this skill fires:

1. Run `glab auth status` and confirm the intended identity on the intended host BEFORE any write — a stale env token silently acts as the wrong user.
2. Re-fetch the three diff SHAs after every push and verify each inline comment came back as a `DiffNote`, never trust a 201 alone.
3. Load the `references/*.md` file matching the task, not everything upfront.

## Requirements

- `glab` on `PATH`, authenticated to the target host. First-time machine setup (install, `glab auth login`, protocol, clone, verify) is in [references/onboarding.md](references/onboarding.md); token types and exact per-operation scopes are in [references/auth.md](references/auth.md).
- A token with `read_api` for read-only work or `api` (plus `write_repository` for git ops) for any write, carried by an identity with **>= Developer** role on the project (or the MR author).
- `jq` for reading API responses (verifying note type, merge status, discussion ids).

## Essentials

- **MR object & branch flags** — `glab mr create --source-branch <b> --target-branch main` (`-s` defaults to current branch, `-b` to project default); glab addresses MRs by per-project `iid`. See [references/create.md](references/create.md).
- **glab does not push** — `git push -u origin HEAD` first (see `git-guide`'s push reference), then create; or pass `--push`. Always pass `--yes` in scripts or create hangs on the confirmation prompt. See [references/create.md](references/create.md).
- **Summary note** — the review's prose is a plain non-anchored note: `glab mr note <iid> -m "## Summary…"` (`POST .../notes`). See [references/review-post.md](references/review-post.md).
- **Inline comment** — a position-anchored discussion: `glab mr note create <iid> --file path --line N -m "…"`, or raw `POST .../discussions` with a position object carrying three SHAs and conditional line keys. Realizes `code-review-guide`'s findings on GitLab. See [references/review-post.md](references/review-post.md).
- **Blocking** — there is no REST `REQUEST_CHANGES`; gate by withholding `/approve` against approval rules or by leaving resolvable threads unresolved with `only_allow_merge_if_all_discussions_are_resolved`. See [references/review-post.md](references/review-post.md).
- **Resolve a thread** — `PUT .../discussions/:discussion_id?resolved=true` (REST is enough on GitLab; no GraphQL needed); match by discussion `id`, never by line. See [references/review-resolve.md](references/review-resolve.md).
- **Deep-link** — MR notes carry no `web_url`; build `<mr_url>#note_<note_id>` from the returned note `id`. See [references/review-post.md](references/review-post.md).
- **Auth & host** — `api` scope for writes (`read_api` reads), `GITLAB_HOST` targets self-managed, `GITLAB_TOKEN` is the general auth token. See [references/auth.md](references/auth.md).

## Gotchas

- **No write-only scope.** Create MR, comment, approve, and resolve ALL require the broad `api` scope (full read+write to every group/project the identity can reach, plus Git-over-HTTP). `read_api` is read-only and 403s on POST; `write_repository` covers only git push / repo files, NOT discussions. Use separate read vs write tokens; prefer a project access token scoped to one project.
- **Scope is necessary but not sufficient.** An `api`-scoped token still 403s without the role: create MR / open a diff thread / resolve needs >= Developer (or MR author); approving needs being an eligible approver; unapprove / reset_approvals work only for bot users (humans 401). A 403 with `api` scope is usually a role or config problem, not a scope one.
- **CI_JOB_TOKEN is read-only on MRs** (GET list/get/notes only) — it cannot create MRs, post notes, or approve. CI that opens or reviews MRs must use a PAT or project/group access token in a masked CI/CD variable.
- **No atomic review object.** Summary = a plain note, each inline comment = its own `POST /discussions`, verdict = a separate `/approve`. The batched "Submit review with summary" exists only in the UI / GraphQL; over REST you publish each piece immediately.
- **Inline anchoring silently fails.** A wrong/stale `base_sha`/`start_sha`/`head_sha` or the wrong line key returns 201 but downgrades the `DiffNote` to a floating `Note` on the Overview tab. ALWAYS verify the response `notes[0].type == "DiffNote"`; re-fetch the three SHAs after every push; remember ADDED→`new_line` only, REMOVED→`old_line` only, UNCHANGED→both, with `old_path` AND `new_path` always required.
- **Not idempotent; one open MR per branch.** A second `create` on the same source branch returns HTTP 409 "Another open merge request already exists for this source branch: !N". List first, then push commits to the existing MR (they attach automatically) or `glab mr update <iid>`.
- **Body and labels are replace-only on update.** `--description` / `description=` overwrite the whole body (no append; read-modify-write); only labels have additive `add_labels` / `remove_labels`.
- **Several review subcommands are EXPERIMENTAL** (`glab mr note list` / `resolve` / `reopen`) and may change or be removed — for stable scripted inline review and resolution drive the Discussions API via `glab api` instead.
- **Identity & host traps.** Env tokens (`GITLAB_TOKEN` / `GITLAB_ACCESS_TOKEN` / `OAUTH_TOKEN`) override stored config; self-managed needs `--hostname` / `GITLAB_HOST` or glab silently hits gitlab.com; `git_protocol` and `api_protocol` are separate per-host settings; the token sits in plaintext `~/.config/glab-cli/config.yml` unless `--use-keyring`. Run `glab auth status` before every write.

## Example — open an MR, post a review with an inline comment, resolve the thread

```bash
# 0. confirm identity + host before any write (a stale env token acts as the wrong user)
glab auth status

# 1. push the branch (see git-guide), then open the MR (--yes or it hangs in scripts)
git push -u origin HEAD
iid=$(glab mr create -R group/project \
  --source-branch "$(git branch --show-current)" --target-branch main \
  --title "feat: x" --description "<what/why/how, see pull-request-guide>" \
  --reviewer alice --label backend --yes \
  | grep -oE '!\d+' | tr -d '!')

# 2a. summary note (plain, non-anchored)
glab mr note "$iid" -R group/project -m "## Review summary
Two issues inline; withholding approval until addressed."

# 2b. inline comment on an added line (glab resolves the three SHAs for you)
glab mr note create "$iid" -R group/project \
  --file src/api/users.go --line 42 \
  -m "**issue (blocking):** unchecked error — see code-review-guide for labels"

# 2c. block by withholding approval (do NOT approve); verify gating
glab api "projects/group%2Fproject/merge_requests/$iid" | jq -r .detailed_merge_status

# 3. later: list threads, match by id, resolve, then approve to clear the gate
glab api "projects/group%2Fproject/merge_requests/$iid/discussions" \
  | jq -r '.[] | select(.notes[0].position.new_path=="src/api/users.go") | .id'
glab api --method PUT \
  "projects/group%2Fproject/merge_requests/$iid/discussions/<discussion_id>?resolved=true"
glab mr approve "$iid" -R group/project
```

## Progressive Disclosure

Each reference is a trigger — read it only when the user's intent matches; do not preload everything.

- Read [references/onboarding.md](references/onboarding.md) — Load when setting up a fresh machine or account: installing glab, `glab auth login` (web / device / stdin), choosing `git_protocol` vs `api_protocol`, cloning, and verifying with a read call.
- Read [references/auth.md](references/auth.md) — Load when auth fails, choosing a token type (PAT / project access token / fine-grained / CI_JOB_TOKEN), scoping it to the exact operation (`read_api` vs `api`), targeting a self-managed host, storing it (keyring / CI), or rotating a leak.
- Read [references/create.md](references/create.md) — Load when opening or updating an MR: push, `glab mr create` flags, draft, reviewers / labels, issue-link / close semantics, the one-open-MR-per-branch 409, additive-body rules, and the raw REST equivalent.
- Read [references/review-post.md](references/review-post.md) — Load when publishing a review: the summary note, the exact inline position model (three SHAs + conditional line keys), verifying `DiffNote`, the blocking mechanism, and the `#note_<id>` deep-link.
- Read [references/review-resolve.md](references/review-resolve.md) — Load when listing threads, matching a finding to a discussion by id, resolving via the REST PUT, replying in-thread, and the merge-gating effect.
