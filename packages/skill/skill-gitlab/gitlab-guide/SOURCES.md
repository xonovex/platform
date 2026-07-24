# Sources

## glab CLI (official GitLab CLI)

- **URLs:** https://docs.gitlab.com/cli/ · https://docs.gitlab.com/cli/api/ · https://docs.gitlab.com/cli/auth/ · https://docs.gitlab.com/cli/auth/login/ · https://docs.gitlab.com/cli/auth/status/ · https://docs.gitlab.com/cli/repo/clone/ · https://docs.gitlab.com/cli/issue/ · https://docs.gitlab.com/cli/issue/create/ · https://docs.gitlab.com/cli/issue/update/ · https://docs.gitlab.com/cli/issue/view/ · https://docs.gitlab.com/cli/issue/board/ · https://docs.gitlab.com/cli/work-items/ · https://docs.gitlab.com/cli/work-items/list/ · https://docs.gitlab.com/cli/work-items/update/ · https://docs.gitlab.com/cli/mr/ · https://docs.gitlab.com/cli/mr/create/ · https://docs.gitlab.com/cli/mr/note/ · https://docs.gitlab.com/cli/mr/note/create/ · https://docs.gitlab.com/cli/mr/note/resolve/ · https://docs.gitlab.com/cli/mr/approve/ · https://docs.gitlab.com/cli/mr/diff/ · https://docs.gitlab.com/cli/mr/view/
- **Last reviewed:** 2026-07-24
- **Used for:** install / supported versions; auth, host, and protocol behavior;
  `glab api` scalar/file input and pagination; issue create/update/view, additive
  labels/assignees, board and experimental work-item commands (issues.md, boards.md);
  MR create/review commands; and Notes/discussion behavior.
- **References:** references/auth.md, references/boards.md,
  references/context-notes.md, references/create.md,
  references/first-time-setup.md, references/issues.md,
  references/provider-conformance.md, references/review-post.md,
  references/review-resolve.md

## GitLab REST & GraphQL API

- **URLs:** https://docs.gitlab.com/api/metadata/ · https://docs.gitlab.com/api/issues/ · https://docs.gitlab.com/api/issue_links/ · https://docs.gitlab.com/api/boards/ · https://docs.gitlab.com/api/group_boards/ · https://docs.gitlab.com/api/notes/ · https://docs.gitlab.com/api/merge_requests/ · https://docs.gitlab.com/api/merge_request_approvals/ · https://docs.gitlab.com/api/discussions/ · https://docs.gitlab.com/api/draft_notes/ · https://docs.gitlab.com/api/graphql/ · https://docs.gitlab.com/api/graphql/reference/ · https://docs.gitlab.com/api/graphql/sample_issue_boards/
- **Last reviewed:** 2026-07-24
- **Used for:** instance version/edition discovery; project issue IID/global-ID
  identity; issue types, metadata, state, additive labels, time tracking, marker
  reconciliation and cross-project links (issues.md); project/group board/list REST
  resources and GraphQL Global IDs,
  status/rank schema discovery (boards.md); paginated issue/MR Notes, internal
  visibility, append-only context identity and note URLs (context-notes.md); plus MR,
  discussions, approvals, and review resolution.
- **References:** references/boards.md, references/context-notes.md,
  references/create.md, references/issues.md,
  references/provider-conformance.md, references/review-post.md,
  references/review-resolve.md

## GitLab issues, work items, Status, and boards

- **URLs:** https://docs.gitlab.com/user/project/issues/ · https://docs.gitlab.com/user/project/issues/managing_issues/ · https://docs.gitlab.com/user/project/issues/related_issues/ · https://docs.gitlab.com/user/project/issue_board/ · https://docs.gitlab.com/user/work_items/ · https://docs.gitlab.com/user/work_items/status/ · https://docs.gitlab.com/api/graphql/epic_work_items_api_migration_guide/
- **Last reviewed:** 2026-07-24
- **Used for:** issue/work-item responsibility; board cards as views over labels,
  assignees, milestones, iterations, or Status; cross-list effects; native Status
  version/tier/category behavior; work-item migration and experimental API boundaries.
- **References:** references/boards.md, references/issues.md,
  references/provider-conformance.md

## GitLab tokens, scopes, and CI

- **URLs:** https://docs.gitlab.com/security/tokens/ · https://docs.gitlab.com/user/profile/personal_access_tokens/ · https://docs.gitlab.com/auth/tokens/fine_grained_access_tokens/ · https://docs.gitlab.com/user/project/settings/project_access_tokens/ · https://docs.gitlab.com/ci/jobs/ci_job_token/
- **Last reviewed:** 2026-06-26
- **Used for:** the `read_api` vs `api` split (no write-only / comment-only scope), `read_repository`/`write_repository` covering only git/repo files, glab needing `api` + `write_repository`, fine-grained PAT (beta, 18.10+) resource/permission least-privilege, project/group access tokens running as bot users with a role, CI_JOB_TOKEN being read-only on MRs, mandatory token expiry (16.0+), and bot-only unapprove/reset_approvals.
- **References:** references/auth.md

## GitLab merge requests, reviews, and discussions (product docs)

- **URLs:** https://docs.gitlab.com/user/project/merge_requests/ · https://docs.gitlab.com/user/project/merge_requests/creating_merge_requests/ · https://docs.gitlab.com/user/project/merge_requests/reviews/ · https://docs.gitlab.com/user/discussions/ · https://docs.gitlab.com/administration/issue_closing_pattern/
- **Last reviewed:** 2026-06-26
- **Used for:** role requirements (>= Developer / MR author / eligible approver), the reviewer "Request changes" state being Premium/Ultimate and not a REST object, `only_allow_merge_if_all_discussions_are_resolved` merge gating (threads block, single comments do not), auto-resolve-on-outdated behavior, and `Closes/Fixes/Resolves #N` issue-closing into the default branch.
- **References:** references/create.md, references/review-post.md, references/review-resolve.md

## GitLab CI/CD and external enforcement

- **URLs:** https://docs.gitlab.com/ci/components/ · https://docs.gitlab.com/ci/yaml/inputs/ · https://docs.gitlab.com/user/application_security/policies/pipeline_execution_policies/ · https://docs.gitlab.com/user/compliance/compliance_frameworks/ · https://docs.gitlab.com/ci/environments/protected_environments/ · https://docs.gitlab.com/ci/environments/deployment_approvals/ · https://docs.gitlab.com/ci/jobs/job_artifacts/
- **Last reviewed:** 2026-07-16
- **Used for:** `references/automation-and-enforcement.md`
- **Aspects extracted:** Component layout, typed inputs, configurable job naming, commit/release/partial/latest references, component tests and catalog release/security guidance; pipeline-policy application without project CI, job-conflict suffix behavior, project/project-policy/group-policy merge order, reserved stages, injection strategies and failure modes; compliance-framework scope; protected-environment/deployment approval and provider-native pipeline/job/artifact/deployment evidence. Transaction, exact-revision, failure-policy, governance-only adoption, and no-certification claims are Xonovex adapter constraints.

## glab tracking issues

- **URLs:** https://gitlab.com/gitlab-org/cli/-/work_items/7646 · https://gitlab.com/gitlab-org/cli/-/issues/7999
- **Last reviewed:** 2026-06-26
- **Used for:** the EXPERIMENTAL status of `glab mr note` review subcommands and the env-var `GLAB_` prefix migration tracking (`gitlab-org/cli` issue 7999).
- **References:** references/auth.md, references/review-post.md, references/review-resolve.md

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; these references combine multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/auth.md, references/boards.md,
  references/context-notes.md, references/create.md,
  references/first-time-setup.md, references/issues.md, references/onboarding.md,
  references/provider-conformance.md, references/review-post.md,
  references/review-resolve.md
- **Last reviewed:** 2026-07-16

## Refresh Workflow

1. Re-check `glab issue`, `issue board`, `work-items`, MR, Notes, auth, and
   environment-variable changes, including experimental commands.
2. Re-verify Issues, Issue Links, Boards, Notes, Work Item GraphQL, Discussions, and
   Metadata APIs against supported self-managed versions and tiers.
3. Re-run authenticated read smoke tests for one issue, board/list, work-item schema,
   MR, and complete Notes pagination.
4. Re-test ticket marker reconciliation, label/status board moves, rank changes,
   append-only context publication, and provider-conformance race handling.
5. Re-test enforcement components/policies/environments against supported editions.
6. Bump each **Last reviewed** above.
