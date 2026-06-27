# review-resolve: list, match, resolve, and reply on review threads

## Guideline

Resolution is **GraphQL-only** — list `pullRequest.reviewThreads`, match a finding to a thread by its node id (`PRRT_…`), never by line, then `resolveReviewThread`. Reply in-thread with `addPullRequestReviewThreadReply`.

## Rationale

REST has no resolve field and does not model review threads at all; only GraphQL exposes thread resolution and the full reply chains.

## List threads (the only source of full reply chains)

```bash
gh api graphql -f query='
query($o:String!,$r:String!,$n:Int!){
  repository(owner:$o,name:$r){
    pullRequest(number:$n){
      reviewThreads(first:100){
        nodes{
          id isResolved isOutdated path line viewerCanResolve
          resolvedBy{login}
          comments(first:50){nodes{databaseId body author{login}}}
        }
      }
    }
  }
}' -F o=OWNER -F r=REPO -F n=123
```

`reviewThreads` is the ONLY connection returning full reply chains (PullRequestReview / timeline drops replies). Paginate past the first 100 if needed.

## Match by id, never by line

Match a finding to a thread via:

- the thread node **`id`** (`PRRT_…`), or
- a contained comment's **`databaseId`** (the REST review-comment id you posted), or
- `path` + comment `body`.

Never match by line number — lines shift as the diff evolves.

## Resolve

```bash
gh api graphql -f query='mutation($t:ID!){resolveReviewThread(input:{threadId:$t}){thread{id isResolved}}}' \
  -f t=PRRT_kwDOxxxxx
# reverse with unresolveReviewThread(input:{threadId:$t})
```

### Permission gotcha

on a fine-grained PAT / App token, `resolveReviewThread` needs **Pull requests: write AND Contents: read & write**, or it fails "Resource not accessible by integration" (community #44650). A classic `repo` token suffices.

## Reply in-thread (preferred, unambiguous)

```bash
gh api graphql -f query='mutation($t:ID!,$b:String!){addPullRequestReviewThreadReply(input:{pullRequestReviewThreadId:$t,body:$b}){comment{id}}}' \
  -f t=PRRT_kwDOxxxxx -f b='Verified the fix, resolving.'
```

This replaced the deprecated `addPullRequestReviewComment` (2023-10-01). The REST `POST .../pulls/{n}/comments/{comment_id}/replies` works only for a top-level comment id and 404s for comments inside a submitted review.

## Merge-gating effect

Resolving a thread blocks merge **only** when "Require conversation resolution before merging" is enabled — a classic branch-protection checkbox; in rulesets it nests under "Require a pull request before merging" and cannot be required independently. Check with:

```bash
gh pr view 123 --json reviewDecision,mergeStateStatus,mergeable   # mergeStateStatus=BLOCKED means a gate is unmet
```

### Orphaned/outdated threads

after a force-push / rebase / squash, threads stay unresolved (`isOutdated:true`) and can block merge with nothing visibly open — use `isOutdated` to detect them. An outdated thread can often be resolved only by the original comment author, so a different automation identity may be unable to clear it.

### Related

[review-post.md](./review-post.md), [auth.md](./auth.md)
