# Sources

## GitHub CLI manual (`gh`)

- **URLs:**
  - https://cli.github.com/manual/gh_auth_login
  - https://cli.github.com/manual/gh_auth_status
  - https://cli.github.com/manual/gh_auth_setup-git
  - https://cli.github.com/manual/gh_help_environment
  - https://cli.github.com/manual/gh_api
  - https://cli.github.com/manual/gh_issue
  - https://cli.github.com/manual/gh_issue_create
  - https://cli.github.com/manual/gh_issue_edit
  - https://cli.github.com/manual/gh_issue_close
  - https://cli.github.com/manual/gh_project
  - https://cli.github.com/manual/gh_project_view
  - https://cli.github.com/manual/gh_project_field-list
  - https://cli.github.com/manual/gh_project_item-list
  - https://cli.github.com/manual/gh_project_item-add
  - https://cli.github.com/manual/gh_project_item-edit
  - https://cli.github.com/manual/gh_project_item-archive
  - https://cli.github.com/manual/gh_project_item-delete
  - https://cli.github.com/manual/gh_pr_create
  - https://cli.github.com/manual/gh_pr_comment
  - https://cli.github.com/manual/gh_pr_edit
  - https://cli.github.com/manual/gh_pr_review
  - https://docs.github.com/en/github-cli/github-cli/quickstart
  - https://github.com/cli/cli/blob/trunk/docs/install_linux.md
- **Last reviewed:** 2026-07-24
- **Used for:** install, `gh auth login` / `status` / `setup-git`
  flow, the GH_TOKEN / GH_ENTERPRISE_TOKEN / GH_HOST environment split (auth.md),
  issue create/edit/close/reopen, metadata and relationship flags (issues.md);
  project scope, discovery, item add/list/edit/archive/delete and typed field flags
  (projects.md); `gh pr create` flags + `gh pr edit` replace-on-body behavior
  (create.md); `gh pr comment` body and last-comment behavior; and the `gh pr review` inline-comment limitation
  (review-post.md).
- **References:** references/auth.md, references/create.md, references/issues.md,
  references/projects.md, references/review-post.md

## GitHub REST API

- **URLs:**
  - https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28
  - https://docs.github.com/en/rest/pulls/reviews
  - https://docs.github.com/en/rest/pulls/comments
  - https://docs.github.com/en/rest/issues/issues
  - https://docs.github.com/en/rest/issues
  - https://docs.github.com/en/rest/issues/sub-issues
  - https://docs.github.com/en/rest/issues/comments
  - https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api
  - https://docs.github.com/en/rest/using-the-rest-api/troubleshooting-the-rest-api
- **Last reviewed:** 2026-07-24
- **Used for:** repository issue identity, pagination, metadata, create/update/state,
  sub-issue relationships, and comment endpoints (issues.md); `POST /pulls` (no-push, head/base semantics) in
  create.md; list and create issue comments on issues and pull requests plus unsafe
  conditional-request limitations; the
  `.../pulls/{n}/reviews` object with `comments[]`
  path/line/side/start_line/start_side/commit_id and the deprecated `position` field
  in review-post.md; the 404-on-private and `X-Accepted-GitHub-Permissions` header
  behavior in auth.md.
- **References:** references/auth.md, references/create.md, references/issues.md, references/review-post.md

## GitHub GraphQL API

- **URLs:**
  - https://docs.github.com/en/graphql/reference/mutations
  - https://docs.github.com/en/graphql/reference/input-objects
  - https://docs.github.com/en/graphql/reference/pulls
  - https://docs.github.com/en/graphql/reference/objects#projectv2
  - https://docs.github.com/en/graphql/reference/mutations#addprojectv2itembyid
  - https://docs.github.com/en/graphql/reference/mutations#updateprojectv2itemfieldvalue
- **Last reviewed:** 2026-06-26
- **Used for:** ProjectV2 project/item/field identities, add-item and typed field
  mutations (projects.md); `resolveReviewThread` / `unresolveReviewThread` /
  `addPullRequestReviewThreadReply` mutations and the `pullRequest.reviewThreads`
  connection in review-resolve.md.
- **References:** references/projects.md, references/review-resolve.md

## GitHub Issues and Projects product guidance

- **URLs:**
  - https://docs.github.com/en/issues/tracking-your-work-with-issues
  - https://docs.github.com/en/issues/planning-and-tracking-with-projects/learning-about-projects/best-practices-for-projects
  - https://docs.github.com/en/issues/planning-and-tracking-with-projects/managing-items-in-your-project/adding-items-to-your-project
  - https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-built-in-automations
- **Last reviewed:** 2026-07-24
- **Used for:** issue-versus-project responsibility; project Status, typed fields,
  item lifecycle, built-in automatic add/status/archive/close effects, and
  re-verification requirements.
- **References:** references/issues.md, references/projects.md

## Tokens, permissions, and Actions

- **URLs:**
  - https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens
  - https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens
  - https://docs.github.com/en/actions/concepts/security/github_token
  - https://github.com/orgs/community/discussions/44650
- **Last reviewed:** 2026-06-26
- **Used for:** the per-operation scope table (Contents: write to push, Pull requests: write to open/review, resolve needing Contents: read & write), classic `repo` scope, and the Actions `permissions:` block in auth.md.
- **References:** references/auth.md

## Issue linking, branch protection, and rulesets

- **URLs:**
  - https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/managing-a-branch-protection-rule
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets
  - https://github.blog/changelog/2026-02-17-required-reviewer-rule-is-now-generally-available/
- **Last reviewed:** 2026-06-26
- **Used for:** closing-keyword auto-close-on-default-branch semantics in create.md; the REQUEST_CHANGES blocking + dismissal mechanism (review-post.md) and the conversation-resolution merge gate (review-resolve.md).
- **References:** references/create.md, references/review-post.md, references/review-resolve.md

## GitHub Actions and external enforcement

- **URLs:**
  - https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
  - https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action
  - https://docs.github.com/en/actions/reference/security/secure-use
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets
  - https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets
  - https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments
  - https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations
- **Last reviewed:** 2026-07-16
- **Used for:** ``
- **Aspects extracted:** `workflow_call` location/input/secret behavior; same-commit and full-SHA references; nested permission attenuation; composite-action scope; full-SHA immutable action pins; least-privilege tokens and untrusted-input handling; aggregated ruleset layering and bypass actors; exact required checks; environment reviewers, self-review prevention, target restrictions and secret release; OIDC/attestation permissions and native evidence. Transaction, failure-policy, exact-revision, and governance-only adoption rules are Xonovex adapter constraints.

## gh CLI limitations (tracking issues)

- **URLs:**
  - https://github.com/cli/cli/issues/12396
  - https://github.com/cli/cli/issues/13358
  - https://github.com/cli/cli/issues/649
  - https://github.com/orgs/community/discussions/161601
  - https://github.com/orgs/community/discussions/10076
- **Last reviewed:** 2026-06-26
- **Used for:** `gh pr review` has no inline support (#12396), `.../pulls/{n}/comments` 422s on line/side (#13358), and related thread/resolve gaps that force the drop to `gh api` REST + GraphQL.
- **References:** references/review-post.md, references/review-resolve.md

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; these references combine multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/auth.md, references/create.md, references/issues.md,
  references/onboarding.md, references/projects.md,
  references/review-post.md,
  references/review-resolve.md
- **Last reviewed:** 2026-07-16

## Refresh Workflow

1. Re-fetch the `gh` manual and REST/GraphQL references; scan for issue relationship,
   Projects field/item, `gh pr review` inline, and review-thread changes.
2. Re-verify a read call (`gh api user`, `gh api graphql -f query='query{viewer{login}}'`) and one `X-Accepted-GitHub-Permissions` header against a `.../reviews` endpoint.
3. Re-confirm issue, Projects, pull-request, and fine-grained-PAT/App permissions.
4. Re-test reusable workflow pin/secret/permission behavior, ruleset layering and required-check source binding, environment approval semantics, and artifact-attestation permissions.
5. Re-test ticket marker reconciliation, Projects add/edit plus automation races, and
   append-only issue/PR context publication.
6. Bump **Last reviewed** above.
