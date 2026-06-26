# GitLab auth — token types and exact per-operation scopes

`glab` authenticates the API with a token; the scope you need is driven entirely by whether the call reads or writes, and a sufficient scope still needs a sufficient role. First-time `glab auth login` flow is in [onboarding.md](onboarding.md); this file covers token type, the exact least-privilege scope per operation, env vars, storage, self-managed targeting, and CI.

## Token types

- **Personal access token (PAT)** — classic, account-wide; the token can never exceed its owner's own access. Scopes are coarse (see below). Create at `https://<host>/-/user_settings/personal_access_tokens`.
- **Project / Group access token** — scoped to one project or group, runs as a bot user with a granted role. Preferred for automation: a compromise is contained to that project. Create under the project's **Settings → Access tokens**.
- **Fine-grained PAT** (beta, GitLab 18.10+) — resource + permission pairs instead of coarse scopes; the only way to get write-without-everything (see least-privilege below). Still beta — track before relying on it in production.
- **CI_JOB_TOKEN** — injected into CI jobs. **Read-only on MRs** (GET list / get / notes only): it cannot create MRs, post notes, or approve. CI that opens or reviews MRs must use a PAT or project/group access token.
- **OAuth token** — for interactive `glab auth login --web` / `--device`; the app requests `openid profile read_user write_repository api`.

Tokens require an expiration (GitLab 16.0+, default 365 days); pick the shortest practical one and rotate before it lapses.

## Exact scopes per operation

GitLab classic PATs have **no write-only or comment-only scope**. The split is binary:

| Operation                                                            | Scope                                  | Also needs                                                                                |
| -------------------------------------------------------------------- | -------------------------------------- | ----------------------------------------------------------------------------------------- |
| `glab auth status`, list / view MR, read diffs / notes / discussions | `read_api`                             | Reporter+ (read access)                                                                   |
| Create MR                                                            | `api`                                  | >= Developer (or be eligible to push the branch)                                          |
| Post summary note / inline discussion                                | `api`                                  | >= Developer (or MR author)                                                               |
| Approve / unapprove                                                  | `api`                                  | eligible approver; unapprove/reset_approvals work **only for bot users** (humans get 401) |
| Resolve / reopen a thread                                            | `api`                                  | >= Developer (or MR author)                                                               |
| git push / read repo files                                           | `write_repository` / `read_repository` | push access                                                                               |

- **`read_api` 403s on any POST** — it is read-only. **`write_repository` covers only Git push and repo files, NOT discussions** — a write_repository-only token cannot comment or resolve.
- **glab itself needs at least `api` AND `write_repository`** — `api` for MR / review API calls, `write_repository` for the git / repo operations it performs.
- **Fine-grained PAT least-privilege (beta):** Merge Request resource = **Create + Approve**, Work Item resource = **Create** (for comments). This is the only classic-free way to avoid granting `api`.
- **Scope is necessary but not sufficient.** An `api`-scoped token still 403s without the role — a 403 with `api` scope is almost always a role or project-config problem, not a scope one.

## Environment variables

- **`GITLAB_TOKEN`** — the **general** authentication token for API requests on **any** instance, **including gitlab.com**. It is NOT a self-managed switch. `GITLAB_ACCESS_TOKEN` and `OAUTH_TOKEN` work too.
- **`GITLAB_HOST`** (aliases `GL_HOST`, `GITLAB_URI`) — targets a self-managed or Dedicated instance by its server URL. This is the self-managed switch, not the token var.
- **Env tokens OVERRIDE stored config credentials** — a leftover `GITLAB_TOKEN` in your shell silently acts as a different identity than `glab auth status` last showed for the stored config. `unset` it when switching.
- **glab 2.0.0+ renames all env vars to a `GLAB_` prefix** (`GLAB_TOKEN`, `GLAB_HOST`); track `gitlab-org/cli` issue 7999 and support both during the transition.

## Self-managed / Dedicated / enterprise targeting

- `glab auth login --hostname gitlab.example.com` (or `GITLAB_HOST`) — REQUIRED, or glab silently targets gitlab.com.
- When the REST endpoint differs from the git remote, set `--api-host` and `--api-protocol` (and `--ssh-hostname` for a distinct SSH host).
- `git_protocol` and `api_protocol` are **separate per-host settings**: `glab config set -h <host> git_protocol ssh` and `glab config set -h <host> api_protocol https`. A working ssh clone does not imply API calls work.

## Storage

- Stored credentials live in **`~/.config/glab-cli/config.yml` in PLAINTEXT by default.** Pass `--use-keyring` on `glab auth login` to put the token in the OS keyring instead.
- Prefer feeding the token from a secret store at call time over `--token` on a shared host (`--token` lands in shell history and `ps`):

```bash
glab auth login --hostname gitlab.example.com --stdin < <(security find-generic-password -a "$USER" -s gitlab-token -w)
# or read into the env var only at call time:
export GITLAB_TOKEN="$(secret-tool lookup service gitlab username "$USER")"   # Linux libsecret
export GITLAB_TOKEN="$(op read 'op://<vault>/gitlab/token')"                   # 1Password, teams
```

- Use a **distinct token per integration** so a compromise is revoked in isolation, and **separate read vs write tokens** (a `read_api` token for listing, an `api` token only where a write happens).

## CI

- **Do not rely on `CI_JOB_TOKEN`** for MR writes — it is read-only on MRs. Inject a PAT or project/group access token as a **masked + protected** CI/CD variable (set in the UI, not committed YAML) and read it into `GITLAB_TOKEN` at runtime.
- Secrets are withheld from forked-repo / detached MR pipelines by default — only expose with explicit justification, and isolate untrusted-branch steps.
- Prefer a project access token scoped to the one project with a short expiry over an account-wide PAT.

## On leak

Revoke + reissue the token first (assume compromise within minutes), then update wherever it is stored, re-run `glab auth login`, and if it was ever committed, rewrite history (`git filter-repo` / BFG, force-push, collaborators re-clone). Never `echo` / `set -x` / `pbcopy` the value.
