# Bitbucket Data Center auth — SSH for git, HTTP token for REST

These recipes target a self-hosted Bitbucket Server / Data Center instance and its REST 1.0 API at `/rest/api/1.0`. git and the REST API authenticate through two independent mechanisms: git over SSH is unrelated to REST over HTTPS, and having one working tells you nothing about the other. This file covers the REST side — the HTTP access token, where to keep it, and how `curl` sends it. Bitbucket Cloud (`api.bitbucket.org`, REST 2.0) is a separate product with a different auth model; never infer Cloud parity from a Data Center setup (see [data-center.md](data-center.md) for the versioned-REST boundary).

## No usable CLI — use REST

Bitbucket Server / Data Center has no usable first-party CLI for this. The common third-party Bitbucket tools target Bitbucket Cloud (`api.bitbucket.org`) and cannot talk to a self-hosted `/rest/api/1.0` instance. Drive the REST API directly with `curl`.

## Resolve the real host

The git remote is usually an SSH alias, for example `git@<ssh-alias>:v3/<KEY>/<repo>`. Resolve it:

```bash
ssh -G <ssh-alias> | grep -iE '^(hostname|user|port) '
# hostname <host>
# port 7999            <- SSH port for git
```

REST is on the same hostname over HTTPS (443): `https://<host>/rest/api/1.0/`. Resolve the concrete host / project / repo / alias from your instance's coordinates — do not duplicate them here. First-time SSH-and-token setup is in [first-time-setup.md](first-time-setup.md).

## Credential type and scope

The REST API needs an HTTP **access token**, not your account password and not the SSH key. Treat it like a password.

1. **Where to create / rotate / revoke.** In Bitbucket: avatar → Manage account → **HTTP access tokens** (or a repo-scoped token under the repo's Settings). Self-hosted tokens can be given an expiry — pick the shortest practical one, not "never", and rotate before it lapses.
2. **Least privilege.** Scope **Pull requests: write** (`REPO_WRITE`) to post or edit comments. A read-only token passes auth checks but `403`s on the first write. A token can never exceed its owner's own access; prefer a repo-scoped token over an account-wide one.
3. **On leak: revoke + reissue first** (assume compromise within minutes), then update wherever it is stored. See Hygiene.

## How this skill sends it (keychain is the source of truth)

Store the token once in the OS keychain and resolve it at call time — never inline it, never let it hit shell history. There are two send paths; the keychain feeds both. **Which path is valid depends on the token scope** (see also [first-time-setup.md](first-time-setup.md)):

- A **personal / user** token works with either path — Bearer or Basic (`~/.netrc` login=username + password=token).
- A **project- or repository-scoped** token is **Bearer-only on Bitbucket DC < 9.4** (8.x); Basic / `~/.netrc` `401`s there. DC ≥ 9.4 (10.x) relaxed this so Basic also works for repo tokens. **If your token is repo-scoped and you are unsure of the DC version, use Path A (Bearer) — it is the safe default.**

**Path A — `Authorization: Bearer` header (no `~/.netrc`, nothing on disk) — the default for repo/project-scoped tokens:**

```bash
TOKEN=$(security find-generic-password -a "$USER" -s bitbucket-token -w)
curl -s -H "Authorization: Bearer $TOKEN" "https://<host>/rest/api/1.0/..."
```

The header is `Authorization: Bearer <token>` — no username (Bitbucket Server/DC ignores it for token auth, unlike Bitbucket Cloud which wants Basic `email:token`). To run the `curl -n` recipes in [pr-comments.md](pr-comments.md) without a `~/.netrc`, swap `-n` for `-H "Authorization: Bearer $TOKEN"`.

**Path B — materialize a `0600` `~/.netrc` (Basic auth) so existing `curl -n` examples keep working.** Valid for a user token, or a repo/project token only on DC ≥ 9.4. The `pr-comments.md` recipes use `curl -n`; that flag requires a `~/.netrc`. Generate it from the keychain value at the start of a session rather than persisting the secret by hand:

```bash
umask 077
printf 'machine <host>\nlogin %s\npassword %s\n' \
  "<bitbucket-username>" "$(security find-generic-password -a "$USER" -s bitbucket-token -w)" > ~/.netrc
chmod 600 ~/.netrc            # curl ignores / warns on a group- or world-readable netrc
```

After that, the unchanged `curl -s -n ...` examples in [pr-comments.md](pr-comments.md) work as written.

**Verify before any write** (`200` good, `401` bad/missing token or wrong username, `403` token lacks `REPO_WRITE`, `404` wrong project/repo):

```bash
BASE=https://<host>/rest/api/1.0/projects/<KEY>/repos/<repo>
curl -s -n -o /dev/null -w '%{http_code}\n' "$BASE/pull-requests?limit=1"          # path B
curl -s -H "Authorization: Bearer $TOKEN" -o /dev/null -w '%{http_code}\n' "$BASE/pull-requests?limit=1"   # path A
```

A `200` confirms read access only; write scope is first exercised on the initial `POST` / `PUT`, where a `403` means the token lacks `REPO_WRITE`.

## Where to store it (local — keychain-first)

Prefer the OS keychain; resolve at runtime so nothing lands on disk or in history. Ladder, most to least preferred:

```bash
# 1. macOS Keychain (security CLI) — login keychain unlocks at GUI login
security add-generic-password -a "$USER" -s bitbucket-token -U -w   # prompts on a hidden line; no value in history/ps
export TOKEN="$(security find-generic-password -a "$USER" -s bitbucket-token -w)"   # read at call time
security delete-generic-password -a "$USER" -s bitbucket-token       # delete / rotate

# 2. Linux Secret Service (secret-tool / libsecret) — gnome-keyring or KWallet
secret-tool store --label="Bitbucket token" service bitbucket username "$USER"      # prompts for the value
export TOKEN="$(secret-tool lookup service bitbucket username "$USER")"
secret-tool clear service bitbucket username "$USER"

# 3. Windows — Credential Manager (cmdkey WRITES only; cannot read back)
cmdkey /generic:bitbucket-token /user:%USERNAME% /pass:%tok%         # set /p tok=Token: first; never a literal
#   Read AND write -> PowerShell SecretManagement + SecretStore:
#   Set-Secret -Name bitbucket-token -Secret (Read-Host -AsSecureString)
#   $env:TOKEN = Get-Secret -Name bitbucket-token -AsPlainText
```

For **teams / shared automation**, fetch on demand from a secret-manager CLI instead of a per-machine keychain:

```bash
export TOKEN="$(op read 'op://<vault>/bitbucket/token')"             # 1Password (OP_SERVICE_ACCOUNT_TOKEN in CI)
export TOKEN="$(vault kv get -mount=secret -field=token bitbucket)"  # HashiCorp Vault (KV v2; use -mount)
```

**Gitignored `.env` fallback** — local dev only: add `.env` / `.envrc` to `.gitignore` **before** the first commit, commit only a `.env.example` with placeholders, `chmod 600 .env`, and prefer referencing a vault inside it (`export TOKEN=$(op read ...)`) over storing the raw value. Plaintext on disk is the last resort.

**Never** inline the literal in code, a committed config, or a `~/.netrc` you hand-edit; never `echo` / `pbcopy` / `set -x` the value.

## CI/CD

Where this token feeds automation, the CI platform (not the git host) owns the secret store, so this matrix is downstream-agnostic — a self-hosted Data Center repository can be built by any of these. Default to short-lived OIDC over a long-lived PAT wherever the downstream supports it.

| Platform            | Stored-secret store                                            | OIDC / short-lived                                                            |
| ------------------- | -------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| Bitbucket Pipelines | Secured variables (workspace / repo / deployment); `$VAR`      | `oidc: true` → `$BITBUCKET_STEP_OIDC_TOKEN` (federate to AWS/GCP/Azure/Vault) |
| GitHub Actions      | Repo / Environment / Org secrets; `${{ secrets.NAME }}`        | `permissions: id-token: write` → cloud login action                           |
| GitLab CI/CD        | CI/CD variables (Masked + Protected, UI not YAML); `$VAR`      | `id_tokens:` with `aud:`                                                      |
| Azure Pipelines     | Secret variables / Key Vault-linked variable groups; `$(NAME)` | Workload-identity-federation service connection                               |
| CircleCI            | Contexts + project env vars; `$VAR`                            | `$CIRCLE_OIDC_TOKEN` (attach a context)                                       |
| Bitrise             | Secrets (encrypted, `[REDACTED]`); `$VAR`                      | "Authenticate with AWS/GCP" steps                                             |
| Jenkins             | Credentials + `withCredentials` binding                        | no native OIDC — federate via agent platform or Vault                         |

- In any CI step, mark the token a **secured** / masked variable and read it into the header at runtime; do not write it into committed CI config. Reconstruct the netrc the same way (`printf ... > ~/.netrc; chmod 600`) inside the step if the recipes need `curl -n`.
- Secrets are withheld from forked-repo / PR builds by default — only expose with explicit justification, and isolate PR steps.
- Prefer Vault / Key Vault indirection so rotation happens in the vault without touching pipeline config.

## Cloud / production

If automation runs as a service, fetch the token from a managed secret manager at process startup; never bake it into code, config, or container images.

```bash
aws secretsmanager get-secret-value --secret-id bitbucket-token --query SecretString --output text  # needs GetSecretValue (+ kms:Decrypt for CMK)
gcloud secrets versions access <VERSION> --secret=bitbucket-token                                    # needs roles/secretmanager.secretAccessor; pin a version
az keyvault secret show --vault-name <vault> --name bitbucket-token --query value -o tsv             # needs Key Vault Secrets User RBAC data role
vault kv get -mount=secret -field=token bitbucket                                                    # short-lived leases — re-auth in long jobs
```

- Authenticate with non-human, scoped identities (IAM instance/task role, GCP service account, Azure Managed Identity, Vault AppRole) — not a personal login.
- Pin explicit secret versions for reproducibility; rotate on a cadence and automate it.

## Hygiene

- Never commit the token — not even to a private repo; code is cloned, forked, and cached beyond access controls. Never commit a populated `~/.netrc`.
- Add `.env` / `.envrc` to `.gitignore` **before** the first commit; commit only `.env.example`. `.gitignore` does **not** scrub history.
- Never pass the token as a literal CLI arg or `echo` / pipe it (lands in shell history, visible in `ps`); read into env vars only at call time via `$(...)` and scope to one command or `unset` after. Don't `set -x` while reading; clear the clipboard.
- `chmod 600 ~/.netrc` — curl ignores or warns about a group- / world-readable netrc.
- Log masking is best-effort and breaks on transform/encode or child-process printing; treat it as a backstop, not a guarantee.
- Run layered secret scanning: a pre-commit hook (Gitleaks / detect-secrets) + a CI diff scan + a scheduled full-history scan (TruffleHog verifies live credentials). Pre-commit hooks are bypassable, so back them with CI.
- Use a distinct token per integration so a compromise can be revoked in isolation.
- On any suspected leak: **revoke / reissue the token first** (it is the only thing that stops the bleeding), then rewrite history if it was committed (`git filter-repo` / BFG, `reflog expire --expire=now --all && gc --prune=now`, force-push, collaborators re-clone), and regenerate `~/.netrc` from the new value. For the history-rewrite mechanics see the git-guide skill.
