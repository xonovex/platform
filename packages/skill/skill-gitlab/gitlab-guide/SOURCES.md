# Sources

## glab CLI (official GitLab CLI)

- **URLs:** https://docs.gitlab.com/cli/ · https://docs.gitlab.com/cli/auth/ · https://docs.gitlab.com/cli/auth/login/ · https://docs.gitlab.com/cli/auth/status/ · https://docs.gitlab.com/cli/repo/clone/ · https://docs.gitlab.com/cli/mr/ · https://docs.gitlab.com/cli/mr/create/ · https://docs.gitlab.com/cli/mr/note/ · https://docs.gitlab.com/cli/mr/note/create/ · https://docs.gitlab.com/cli/mr/note/resolve/ · https://docs.gitlab.com/cli/mr/approve/ · https://docs.gitlab.com/cli/mr/diff/ · https://docs.gitlab.com/cli/mr/view/
- **Last reviewed:** 2026-06-26
- **Used for:** install / supported versions, `glab auth login` (web / device / stdin / `--hostname`) and `glab auth status`, `git_protocol` vs `api_protocol` config, `glab repo clone`, `glab mr create` flags (`-s`/`-b`, `--yes`, `--push`, `--draft`/`--wip`, `--reviewer`/`--assignee` usernames, `--label`, `--related-issue`), `glab mr note` / `note create` (`--file`/`--line`/`--old-line`/`--reply`/`--unique`) and the EXPERIMENTAL `note list`/`resolve`/`reopen`, `glab mr approve`, env vars (`GITLAB_TOKEN`/`GITLAB_HOST` and the `GLAB_` rename in 2.0.0+), and the plaintext-config / `--use-keyring` storage behavior.

## GitLab REST & GraphQL API

- **URLs:** https://docs.gitlab.com/api/merge_requests/ · https://docs.gitlab.com/api/merge_request_approvals/ · https://docs.gitlab.com/api/discussions/ · https://docs.gitlab.com/api/draft_notes/ · https://docs.gitlab.com/api/graphql/reference/
- **Last reviewed:** 2026-06-26
- **Used for:** `POST /merge_requests` (numeric `reviewer_ids`/`assignee_ids`, comma-separated `labels`, 409 one-open-MR-per-branch) and the replace-only `PUT` body / additive `add_labels`/`remove_labels`, `POST /notes` vs `POST /discussions`, the inline position object (`position_type=text`, mandatory `base_sha`/`start_sha`/`head_sha`, required `old_path`+`new_path`, conditional `old_line`/`new_line`), the `DiffNote`-vs-`Note` downgrade, `PUT .../discussions/:discussion_id?resolved=true` (REST-only resolve; thread-vs-single-comment resolvability), single-note resolve, `/approve` (`&sha`, 409), `detailed_merge_status`, and the GraphQL Global-ID `discussionToggleResolve`.

## GitLab tokens, scopes, and CI

- **URLs:** https://docs.gitlab.com/security/tokens/ · https://docs.gitlab.com/user/profile/personal_access_tokens/ · https://docs.gitlab.com/auth/tokens/fine_grained_access_tokens/ · https://docs.gitlab.com/user/project/settings/project_access_tokens/ · https://docs.gitlab.com/ci/jobs/ci_job_token/
- **Last reviewed:** 2026-06-26
- **Used for:** the `read_api` vs `api` split (no write-only / comment-only scope), `read_repository`/`write_repository` covering only git/repo files, glab needing `api` + `write_repository`, fine-grained PAT (beta, 18.10+) resource/permission least-privilege, project/group access tokens running as bot users with a role, CI_JOB_TOKEN being read-only on MRs, mandatory token expiry (16.0+), and bot-only unapprove/reset_approvals.

## GitLab merge requests, reviews, and discussions (product docs)

- **URLs:** https://docs.gitlab.com/user/project/merge_requests/ · https://docs.gitlab.com/user/project/merge_requests/creating_merge_requests/ · https://docs.gitlab.com/user/project/merge_requests/reviews/ · https://docs.gitlab.com/user/discussions/ · https://docs.gitlab.com/administration/issue_closing_pattern/
- **Last reviewed:** 2026-06-26
- **Used for:** role requirements (>= Developer / MR author / eligible approver), the reviewer "Request changes" state being Premium/Ultimate and not a REST object, `only_allow_merge_if_all_discussions_are_resolved` merge gating (threads block, single comments do not), auto-resolve-on-outdated behavior, and `Closes/Fixes/Resolves #N` issue-closing into the default branch.

## glab tracking issues

- **URLs:** https://gitlab.com/gitlab-org/cli/-/work_items/7646 · https://gitlab.com/gitlab-org/cli/-/issues/7999
- **Last reviewed:** 2026-06-26
- **Used for:** the EXPERIMENTAL status of `glab mr note` review subcommands and the env-var `GLAB_` prefix migration tracking (`gitlab-org/cli` issue 7999).

## Refresh Workflow

1. Re-check the glab CLI docs for renamed flags / env vars (the `GLAB_` prefix migration) and whether the `glab mr note` resolve/list subcommands are still EXPERIMENTAL.
2. Re-verify the Discussions API position fields and the `DiffNote` downgrade behavior, and confirm `read_api` vs `api` scope split and CI_JOB_TOKEN read-only status are unchanged.
3. Re-run a read smoke test (`glab auth status` + `glab mr list`) against a real instance.
4. Bump each **Last reviewed** above.
