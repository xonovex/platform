# Sources

## GitHub CLI manual (`gh`)

- **URLs:**
  - https://cli.github.com/manual/gh_auth_login
  - https://cli.github.com/manual/gh_auth_status
  - https://cli.github.com/manual/gh_auth_setup-git
  - https://cli.github.com/manual/gh_help_environment
  - https://cli.github.com/manual/gh_api
  - https://cli.github.com/manual/gh_pr_create
  - https://cli.github.com/manual/gh_pr_edit
  - https://cli.github.com/manual/gh_pr_review
  - https://docs.github.com/en/github-cli/github-cli/quickstart
  - https://github.com/cli/cli/blob/trunk/docs/install_linux.md
- **Last reviewed:** 2026-06-26
- **Used for:** install (onboarding.md), `gh auth login` / `status` / `setup-git` flow, the GH_TOKEN / GH_ENTERPRISE_TOKEN / GH_HOST environment split (auth.md), `gh pr create` flags + `gh pr edit` replace-on-body behavior (create.md), and the `gh pr review` inline-comment limitation (review-post.md).

## GitHub REST API

- **URLs:**
  - https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28
  - https://docs.github.com/en/rest/pulls/reviews
  - https://docs.github.com/en/rest/pulls/comments
  - https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api
- **Last reviewed:** 2026-06-26
- **Used for:** `POST /pulls` (no-push, head/base semantics) in create.md; the `.../pulls/{n}/reviews` object with `comments[]` path/line/side/start_line/start_side/commit_id and the deprecated `position` field in review-post.md; the 404-on-private and `X-Accepted-GitHub-Permissions` header behavior in auth.md.

## GitHub GraphQL API

- **URLs:**
  - https://docs.github.com/en/graphql/reference/mutations
  - https://docs.github.com/en/graphql/reference/input-objects
  - https://docs.github.com/en/graphql/reference/pulls
- **Last reviewed:** 2026-06-26
- **Used for:** `resolveReviewThread` / `unresolveReviewThread` / `addPullRequestReviewThreadReply` mutations and the `pullRequest.reviewThreads` connection in review-resolve.md.

## Tokens, permissions, and Actions

- **URLs:**
  - https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens
  - https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens
  - https://docs.github.com/en/actions/concepts/security/github_token
  - https://github.com/orgs/community/discussions/44650
- **Last reviewed:** 2026-06-26
- **Used for:** the per-operation scope table (Contents: write to push, Pull requests: write to open/review, resolve needing Contents: read & write), classic `repo` scope, and the Actions `permissions:` block in auth.md.

## Issue linking, branch protection, and rulesets

- **URLs:**
  - https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets
  - https://github.blog/changelog/2026-02-17-required-reviewer-rule-is-now-generally-available/
- **Last reviewed:** 2026-06-26
- **Used for:** closing-keyword auto-close-on-default-branch semantics in create.md; the REQUEST_CHANGES blocking + dismissal mechanism (review-post.md) and the conversation-resolution merge gate (review-resolve.md).

## gh CLI limitations (tracking issues)

- **URLs:**
  - https://github.com/cli/cli/issues/12396
  - https://github.com/cli/cli/issues/13358
  - https://github.com/cli/cli/issues/649
  - https://github.com/orgs/community/discussions/161601
  - https://github.com/orgs/community/discussions/10076
- **Last reviewed:** 2026-06-26
- **Used for:** `gh pr review` has no inline support (#12396), `.../pulls/{n}/comments` 422s on line/side (#13358), and related thread/resolve gaps that force the drop to `gh api` REST + GraphQL.

## Refresh Workflow

1. Re-fetch the `gh` manual and the REST/GraphQL reference pages above; scan for new/removed flags and fields (especially `gh pr review` gaining inline support, and any REST resolve-thread endpoint).
2. Re-verify a read call (`gh api user`, `gh api graphql -f query='query{viewer{login}}'`) and one `X-Accepted-GitHub-Permissions` header against a `.../reviews` endpoint.
3. Re-confirm the fine-grained-PAT scope table (Contents: write to push; resolve needing Contents: read & write).
4. Bump **Last reviewed** above.
