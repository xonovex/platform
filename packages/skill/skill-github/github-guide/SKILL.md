---
name: github-guide
description: "Use when delivering a pull request and its review on GitHub (github.com or GitHub Enterprise Server) from the command line with the `gh` CLI and `gh api` — opening a PR with `gh pr create`, posting a structured review with line-anchored inline comments, resolving review threads, detecting that the host is GitHub from the git remote, and scoping a classic or fine-grained token. Triggers on a github.com / GHES remote, `gh pr create`, `gh api .../pulls/.../reviews`, an inline review comment by path+line, resolving a review thread (resolveReviewThread), REQUEST_CHANGES / branch-protection merge gating, or GH_TOKEN / GH_ENTERPRISE_TOKEN scopes — even when the user doesn't say 'gh' but the repo is hosted on GitHub."
---

# GitHub PR & review delivery — quick reference

How to realize a pull request and its review on GitHub from the shell. This is the host-**delivery** tier: it does not teach the craft, only how to land it on GitHub.

- _What a good review comment says_ (Conventional Comments labels, blocking vs non-blocking, summary plus inline, cross-linking) belongs to **`code-review-guide`**.
- _What a good PR description says_ (sizing, how-tested, tradeoffs, self-review) belongs to **`pull-request-guide`**.
- _The local-git push and rebase_ that get a clean branch onto the remote belong to **`git-guide`**'s push reference.

This skill only takes those finished artifacts and posts them through `gh` / `gh api`.

The one thing to internalize: **a GitHub review is one atomic object — a single `POST .../pulls/{n}/reviews` carries the summary body, every line-anchored inline comment, AND the APPROVE / REQUEST_CHANGES / COMMENT verdict together — but `gh pr review` posts only the summary verdict, so inline comments drop to `gh api` (REST) and resolving a thread drops to `gh api graphql` (GraphQL-only), while merge-blocking lives entirely in branch protection / rulesets, not in the review.**

When this skill fires:

1. Confirm the host is GitHub and auth works — `gh auth status` then a real read call (`gh api user`) — before any write.
2. Reach for `gh api` the moment you need an inline comment or a thread resolve; the high-level `gh pr` verbs cannot do either.
3. Load the `references/*.md` file matching the task, not everything upfront.

## Requirements

- `gh` authenticated to the target host. First-time install + `gh auth login` + protocol + clone + verify are in [references/onboarding.md](references/onboarding.md); token families, per-operation least-privilege scopes, and storage are in [references/auth.md](references/auth.md).
- The git remote points at GitHub: `gh repo view --json nameWithOwner,url` resolves it; a `github.com` host (or a GHES host) on `git remote get-url origin` is the detection signal. For GHES, set `GH_HOST` / `--hostname` and use `GH_ENTERPRISE_TOKEN` (see auth).

## Essentials

- **Open a PR** — `gh pr create --base <target> --head <source> --title … --body-file -` auto-pushes the branch if it isn't on the remote and sets reviewers/labels/assignees in one call; the raw `POST /repos/{owner}/{repo}/pulls` does NOT push. See [references/create.md](references/create.md).
- **Push / rebase first** — the clean branch is `git-guide`'s job; this skill assumes the source branch builds on the latest target. See `git-guide`'s push reference.
- **Post a structured review** — `gh pr review` only carries the summary + verdict; batch the summary, every inline comment, and the verdict into one `gh api … /pulls/{n}/reviews` object. See [references/review-post.md](references/review-post.md).
- **Write the review content with code-review-guide** — labels, severity, and blocking decoration are `code-review-guide`'s; this skill only anchors and submits them.
- **Write the PR body with pull-request-guide** — `gh pr create --body-file` takes a description authored per `pull-request-guide`.
- **Resolve a thread** — GraphQL-only `resolveReviewThread` by thread node id (`PRRT_…`), matched by id never by line; list threads with `pullRequest.reviewThreads`. See [references/review-resolve.md](references/review-resolve.md).
- **Scope the token per operation** — push needs Contents: write; open-PR / post-review need Pull requests: write; resolve also needs Contents: read & write on a fine-grained token. See [references/auth.md](references/auth.md).

## Gotchas

- A fine-grained PAT must have **Contents: write** to push commits/refs (`POST/PATCH .../git/refs`, `PUT .../contents`) — Contents: read is enough only to OPEN a PR and POST a review, never to push.
- `gh pr edit --body` / `--body-file` **replaces** the whole description (never appends) — fetch first (`gh pr view N --json body -q .body`), recombine, then set; metadata is incremental via paired `--add-*` / `--remove-*` flags.
- `gh pr create` is **not** create-or-update — it aborts non-zero if an open PR already exists for the branch; guard with `|| gh pr view --json url -q .url`. Only `--web` proceeds anyway, and `--dry-run` "may still push git changes".
- `gh pr review` has **no inline support** (cli/cli#12396) and the standalone `.../pulls/{n}/comments` endpoint 422s on `line`/`side` payloads (cli/cli#13358) — put inline comments inside one `.../reviews` object.
- `line` in the reviews API is a **file line number** + `side`, NOT the deprecated diff `position` (never compute position); a multi-line range needs `start_line`/`start_side` preceding `line`/`side` in the same hunk. Anchor to the PR HEAD sha or comments go "outdated".
- You **cannot** APPROVE or REQUEST_CHANGES your own PR (HTTP 422) — self-reviews are COMMENT-only.
- REQUEST_CHANGES blocks merge only under branch-protection / ruleset required reviews, and clears only via the **same** reviewer approving or a write-access **dismiss** (`PUT .../dismissals`) — another person's approval does not override it.
- Thread resolution is GraphQL-only; on a fine-grained PAT / App token it also silently needs **Contents: read & write** or returns "Resource not accessible by integration" (community #44650). Classic `repo` suffices.
- `Closes #N` in the PR **body** auto-closes the issue only when the PR merges into the **default** branch; cross-repo needs `Closes owner/repo#N`. There is no `gh` flag for it.
- GitHub Enterprise Server uses **GH_ENTERPRISE_TOKEN** / GITHUB_ENTERPRISE_TOKEN, not GH_TOKEN — mixing them is the classic "works on github.com, 401 on GHES".
- Insufficient permission on a **private** repo returns 404 (not 403) — GitHub hides existence; read the `X-Accepted-GitHub-Permissions` header (`gh api -i <endpoint>`) for the exact required permission.

## Example — open a PR, post a blocking inline review, resolve a thread

```bash
# 0. confirm the host is GitHub and you can read
gh repo view --json nameWithOwner,url -q '.url'   # github.com/... or your GHES host
gh auth status && gh api user -q '.login'

# 1. open the PR (gh pushes feat/x if it isn't on the remote); idempotency guard
gh pr create --base main --head feat/x \
  --title "feat: guard null user" --body-file pr-body.md \
  --reviewer org/team-slug --assignee @me --label enhancement \
  || gh pr view feat/x --json url -q .url

# 2. post ONE review object: summary + inline comment + verdict (gh pr review can't do inline)
HEAD=$(gh pr view 123 --json headRefOid -q .headRefOid)
gh api --method POST repos/{owner}/{repo}/pulls/123/reviews \
  -f commit_id="$HEAD" -f event=REQUEST_CHANGES \
  -f body=$'## Summary\nOne blocking issue, see inline.' \
  -f 'comments[][path]=src/app.ts' \
  -F 'comments[][line]=42' \
  -f 'comments[][side]=RIGHT' \
  -f 'comments[][body]=**issue (blocking):** guard against a null `user` here.'

# 3. list threads, then resolve by node id (never by line) — GraphQL only
gh api graphql -f query='query($o:String!,$r:String!,$n:Int!){repository(owner:$o,name:$r){pullRequest(number:$n){reviewThreads(first:100){nodes{id isResolved path comments(first:1){nodes{databaseId body}}}}}}}' \
  -F o=OWNER -F r=REPO -F n=123
gh api graphql -f query='mutation($t:ID!){resolveReviewThread(input:{threadId:$t}){thread{id isResolved}}}' \
  -f t=PRRT_kwDOxxxxx
```

## Progressive Disclosure

Each reference is a trigger — read only the one matching the user's intent; do not preload everything.

- Read [references/onboarding.md](references/onboarding.md) — Load when setting up a fresh machine/account: installing `gh`, running `gh auth login`, picking HTTPS vs SSH, making gh the git credential helper, cloning, and verifying with a read call (GHES included).
- Read [references/auth.md](references/auth.md) — Load when auth fails, choosing classic vs fine-grained PAT, scoping a token per operation (push / open-PR / review / resolve), the GH_TOKEN vs GH_ENTERPRISE_TOKEN split, keyring storage, or wiring tokens into CI / GitHub Actions.
- Read [references/create.md](references/create.md) — Load when opening a PR: `gh pr create` flags, draft, reviewers/labels, issue-linking and auto-close semantics, the additive-body / replace-on-edit trap, idempotency guard, and the raw `POST /pulls` REST equivalent.
- Read [references/review-post.md](references/review-post.md) — Load when publishing a review: the single `.../reviews` object, the exact path/line/side inline anchor model, the REQUEST_CHANGES blocking mechanism, and deep-linking from `html_url`.
- Read [references/review-resolve.md](references/review-resolve.md) — Load when resolving/replying on threads: listing `reviewThreads`, matching a finding to a thread by id (never line), the GraphQL `resolveReviewThread` mutation, in-thread replies, and the conversation-resolution merge gate.
