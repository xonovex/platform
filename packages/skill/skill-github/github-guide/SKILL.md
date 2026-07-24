---
name: github-guide
description: "Use when managing GitHub tickets, Projects kanban, pull-request delivery/review, durable issue or PR context comments, or native CI/repository enforcement on github.com or GitHub Enterprise Server. Triggers on GitHub issues, issue relationships, `gh issue`, Projects v2 items/Status/fields, `gh project`, decision or rationale comments, `gh pr`, inline reviews, resolveReviewThread, reusable workflows, rulesets, protected environments, attestations, or GitHub permissions — even when the user doesn't say 'gh'."
---

# GitHub workflow delivery and enforcement — quick reference

How to realize tickets, kanban, pull-request delivery, durable context projections,
and provider-native external enforcement on GitHub. These operations do not teach
ticket, review, or PR content craft; they map finished artifacts and workflow intent
onto GitHub-native resources.

- _What a good review comment says_ (Conventional Comments labels, blocking vs non-blocking, summary plus inline, cross-linking) belongs to **`code-review-guide`**.
- _What a good PR description says_ (sizing, how-tested, tradeoffs, self-review) belongs to **`pull-request-guide`**.
- _The local-git push and rebase_ that get a clean branch onto the remote belong to **`git-guide`**'s push reference.

This skill only takes those finished artifacts and posts them through `gh` / `gh api`.

The one thing to internalize: **a repository issue, a ProjectV2 item, a project Status
field, a pull-request review, and a top-level context comment are separate native
objects. Preserve and mutate their exact IDs separately; never turn project status
into issue state, context into approval, or an issue comment into an inline review.**

When this skill fires:

1. Confirm the host is GitHub and auth works — `gh auth status` then a real read call (`gh api user`) — before any write.
2. Resolve the exact issue, project item/field, pull request, review thread, or comment
   identity before writing.
3. Reach for `gh api` when high-level verbs cannot express the exact native operation.
4. Use an append-only top-level issue comment for durable context, not a review
   comment or verdict.
5. Load the `references/*.md` file matching the task, not everything upfront.

## Requirements

- `gh` authenticated to the target host. First-time install + `gh auth login` + protocol + clone + verify are in [references/first-time-setup.md](references/first-time-setup.md); token families, per-operation least-privilege scopes, and storage are in [references/auth.md](references/auth.md).
- The git remote points at GitHub: `gh repo view --json nameWithOwner,url` resolves it; a `github.com` host (or a GHES host) on `git remote get-url origin` is the detection signal. For GHES, set `GH_HOST` / `--hostname` and use `GH_ENTERPRISE_TOKEN` (see auth).
- Projects operations require the `project` scope and explicit user or organization
  owner; refresh with `gh auth refresh -s project` when appropriate.

## Essentials

- **Manage workflow tickets** — create/reconcile repository issues; read and update
  assignees, labels, milestone, type, parent/sub-issue and blocking relationships; and
  preserve repository + number + node ID + URL + observed revision. See
  [references/issues.md](references/issues.md).
- **Manage Projects kanban** — resolve project, item, field, and option IDs; add each
  issue/PR once; set typed fields; distinguish Project Status from issue state; and
  report weak concurrency boundaries. See [references/projects.md](references/projects.md).
- **Open a PR** — `gh pr create --base <target> --head <source> --title … --body-file -` auto-pushes the branch if it isn't on the remote and sets reviewers/labels/assignees in one call; the raw `POST /repos/{owner}/{repo}/pulls` does NOT push. See [references/create.md](references/create.md).
- **Publish durable issue/PR context** — use an append-only top-level issue comment,
  give it a marker containing canonical ID/version/digest, reconcile exact retries,
  and return its `html_url` for later human or agent work. See
  [references/context-comments.md](references/context-comments.md).
- **Push / rebase first** — the clean branch is `git-guide`'s job; this skill assumes the source branch builds on the latest target. See `git-guide`'s push reference.
- **Post a structured review** — `gh pr review` only carries the summary + verdict; batch the summary, every inline comment, and the verdict into one `gh api … /pulls/{n}/reviews` object. See [references/review-post.md](references/review-post.md).
- **Write the review content with code-review-guide** — labels, severity, and blocking decoration are `code-review-guide`'s; this skill only anchors and submits them.
- **Write the PR body with pull-request-guide** — `gh pr create --body-file` takes a description authored per `pull-request-guide`.
- **Resolve a thread** — GraphQL-only `resolveReviewThread` by thread node id (`PRRT_…`), matched by id never by line; list threads with `pullRequest.reviewThreads`. See [references/review-resolve.md](references/review-resolve.md).
- **Scope the token per operation** — push needs Contents: write; open-PR / post-review need Pull requests: write; resolve also needs Contents: read & write on a fine-grained token. See [references/auth.md](references/auth.md).
- **Enforce with native automation** — use SHA-pinned reusable workflows/actions, required checks and layered rulesets, protected environments, least-privilege tokens/OIDC, attestations, and provider-native evidence. See [references/automation-and-enforcement.md](references/automation-and-enforcement.md).
- **Transact every setup** — discover, preview the exact native changes and authority, authorize, apply idempotently against observed revisions, re-read and probe both allow and deny, retain rollback, and own the drift. See [references/onboarding.md](references/onboarding.md).
- **Preserve provider conformance** — map workflow tickets, kanban, changes, reviews,
  context, and evidence to their separate native identities and expose unsupported
  guarantees. See [references/provider-conformance.md](references/provider-conformance.md).

## Gotchas

- A fine-grained PAT must have **Contents: write** to push commits/refs (`POST/PATCH .../git/refs`, `PUT .../contents`) — Contents: read is enough only to OPEN a PR and POST a review, never to push.
- `gh pr edit --body` / `--body-file` **replaces** the whole description (never appends) — fetch first (`gh pr view N --json body -q .body`), recombine, then set; metadata is incremental via paired `--add-*` / `--remove-*` flags.
- `gh pr create` is **not** create-or-update — it aborts non-zero if an open PR already exists for the branch; guard with `|| gh pr view --json url -q .url`. Only `--web` proceeds anyway, and `--dry-run` "may still push git changes".
- `gh issue create` is not create-or-update. Reconcile a stable body marker after an
  unknown result; GitHub does not enforce marker uniqueness, so concurrent creates
  still need one writer or an external lock.
- `gh pr comment --edit-last` can overwrite the authenticated identity's unrelated
  latest comment. Context publication is append-only: exact ID/version/digest is a
  no-op, same-version divergence blocks, and changed semantics create a successor.
- Project item identity and Project Status are separate from issue identity and
  open/closed state. `gh issue create --project` adds an item but does not replace
  explicit field discovery and update.
- Project field edits have no general compare-and-swap. Re-read the item, field, and
  old value before apply and report the remaining race window.
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

- Read [references/first-time-setup.md](references/first-time-setup.md) — Load when setting up a fresh machine/account: installing `gh`, running `gh auth login`, picking HTTPS vs SSH, making gh the git credential helper, cloning, and verifying with a read call (GHES included).
- Read [references/auth.md](references/auth.md) — Load when auth fails, choosing classic vs fine-grained PAT, scoping a token per operation (push / open-PR / review / resolve), the GH_TOKEN vs GH_ENTERPRISE_TOKEN split, keyring storage, or wiring tokens into CI / GitHub Actions.
- Read [references/issues.md](references/issues.md) — Load when creating, finding,
  updating, relating, closing, reopening, or commenting on GitHub issues as workflow
  tickets.
- Read [references/projects.md](references/projects.md) — Load when adding tickets or
  pull requests to GitHub Projects, setting Status or typed fields, archiving/removing
  items, or reconciling project automation.
- Read [references/create.md](references/create.md) — Load when opening a PR: `gh pr create` flags, draft, reviewers/labels, issue-linking and auto-close semantics, the additive-body / replace-on-edit trap, idempotency guard, and the raw `POST /pulls` REST equivalent.
- Read [references/context-comments.md](references/context-comments.md) — Load when
  publishing or reconciling versioned decision, rationale, tradeoff, assumption, or
  constraint context as a top-level issue or pull-request comment.
- Read [references/review-post.md](references/review-post.md) — Load when publishing a review: the single `.../reviews` object, the exact path/line/side inline anchor model, the REQUEST_CHANGES blocking mechanism, and deep-linking from `html_url`.
- Read [references/review-resolve.md](references/review-resolve.md) — Load when resolving/replying on threads: listing `reviewThreads`, matching a finding to a thread by id (never line), the GraphQL `resolveReviewThread` mutation, in-thread replies, and the conversation-resolution merge gate.
- Read [references/automation-and-enforcement.md](references/automation-and-enforcement.md) — Load when configuring or onboarding reusable workflows/actions, required checks, rulesets, protected environments, token/OIDC permissions, attestations, evidence, rollback, or governance-only GitHub adoption.
- Read [references/onboarding.md](references/onboarding.md) — Load when setting up, diagnosing, dry-running, verifying, rolling back, uninstalling, or checking drift on repository/organization governance.
- Read [provider conformance](references/provider-conformance.md) — Load when composing
  GitHub tickets, Projects, pull requests, reviews, context, and evidence into a
  provider-neutral workflow handoff or testing adapter behavior.
