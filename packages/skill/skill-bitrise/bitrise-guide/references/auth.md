# Bitrise API auth and app discovery

## Token type and scope

Bitrise API calls authenticate with a **personal access token** (Bitrise -> Account settings -> Security -> Personal access tokens). Treat it like a password.

- **Account-wide, not repo-scoped.** The token can reach every app and org the account can see. There is no separate "read-only" log token — reading a build log and triggering a build use the **same** token, so anyone holding it can start builds. Use a token belonging to a least-privileged automation account rather than a personal admin login where you can.
- **Short-lived.** Set the shortest practical expiry, not "never"; rotate before it lapses. The value is shown **once** at creation — capture it straight into the keychain (below).
- **Create / rotate / revoke** at Bitrise -> Account settings -> Security. On leak: **revoke + reissue first** (assume compromise within minutes), then update wherever it is stored.

## How this skill sends it

Bitrise takes the **raw** token in an `Authorization` header — **no** `Bearer`, **no** `token ` prefix (unlike most OAuth APIs). Read the credential at call time and feed it straight into the request:

```bash
TOKEN=$(security find-generic-password -s bitrise-token -w)
curl -s -H "Authorization: $TOKEN" "https://api.bitrise.io/v0.1/me"
```

Get the scheme exactly right:

```bash
-H "Authorization: $TOKEN"          # correct
-H "Authorization: Bearer $TOKEN"   # WRONG — 401
-H "Authorization: token $TOKEN"    # WRONG — 401
```

If a previously-working token starts 401ing, check that no prefix crept into the header.

Two related quirks of this API, not auth proper:

- The build log's `expiring_raw_log_url` is a short-lived signed (S3) URL — fetch it with **no** Authorization header; the header can make the storage backend reject it (see [builds.md](builds.md)).
- The git host's build-status bridge (commit -> Bitrise build) needs **no** Bitrise token; it authenticates to the git host over HTTP basic auth via `~/.netrc`. That netrc is for the git host only — Bitrise itself never uses netrc (see [builds.md](builds.md)).

## Where to store it (local — keychain-first)

Prefer the OS keychain for local development; resolve at runtime so nothing lands on disk or in shell history. Ladder, most to least preferred:

```bash
# 1. macOS Keychain (security CLI) — login keychain unlocks at GUI login
security add-generic-password -a "$USER" -s bitrise-token -U -w   # prompts on a hidden line; no value in history/ps
export TOKEN="$(security find-generic-password -s bitrise-token -w)"   # read at call time
security delete-generic-password -s bitrise-token                 # delete / rotate

# 2. Linux Secret Service (secret-tool / libsecret) — gnome-keyring or KWallet
secret-tool store --label="Bitrise token" service bitrise username "$USER"   # prompts for the value
export TOKEN="$(secret-tool lookup service bitrise username "$USER")"
secret-tool clear service bitrise username "$USER"

# 3. Windows — Credential Manager (cmdkey WRITES only; cannot read back)
cmdkey /generic:bitrise-token /user:%USERNAME% /pass:%tok%        # set /p tok=Token: first; never a literal
#   Read AND write -> PowerShell SecretManagement + SecretStore:
#   Set-Secret -Name bitrise-token -Secret (Read-Host -AsSecureString)
#   $env:TOKEN = Get-Secret -Name bitrise-token -AsPlainText
```

For **teams / shared automation**, fetch on demand from a secret-manager CLI instead of a per-machine keychain:

```bash
export TOKEN="$(op read "op://<vault>/bitrise/token")"               # 1Password (OP_SERVICE_ACCOUNT_TOKEN in CI)
export TOKEN="$(vault kv get -mount=secret -field=token bitrise)"    # HashiCorp Vault (KV v2; use -mount)
```

**Gitignored file fallback** (documented, explicit) — acceptable for local dev only when no keychain is available:

```bash
umask 077
mkdir -p ~/.config/bitrise
printf '%s' 'YOUR_BITRISE_TOKEN' > ~/.config/bitrise/token
chmod 600 ~/.config/bitrise/token
export TOKEN="$(cat ~/.config/bitrise/token)"
```

Keep it **outside any git repository** so a stray `git add -A` can never commit it. Plaintext on disk — prefer the keychain. Do **not** use `~/.netrc` for the Bitrise token: Bitrise authenticates with an `Authorization` header, not HTTP basic auth, so netrc (which curl applies as basic auth with `-n`) does not help here.

**Never** inline the literal in code, a committed config, or an MCP JSON; never `echo` / `pbcopy` / `set -x` the value. See Hygiene.

## CI/CD — inside a Bitrise pipeline

This skill drives Bitrise from a local shell. When code that needs the token runs **inside** a Bitrise pipeline, do not read a keychain or a `~/.config` file — use **Bitrise Secrets** (encrypted secret env vars, exposed as `$VAR`, redacted to `[REDACTED]` in logs). Expose them only to the steps that need them, and withhold them from PR / forked builds unless explicitly justified; fork exposure, withholding, and the rotation/revocation lifecycle are covered in [secrets-artifacts-and-status.md](secrets-artifacts-and-status.md).

For Bitrise-to-AWS (or other cloud) access, prefer **short-lived OIDC / workload identity federation** over a long-lived stored PAT: Bitrise exposes "Authenticate with AWS/GCP" steps whose tokens carry issuer `token.builds.bitrise.io`, and the cloud-side trust policy must be constrained to specific claims (issuer, audience, and workspace / app / repository / workflow / environment). The full trust-preview and verification flow is in [aws-oidc.md](aws-oidc.md).

## Cloud / production

If a long-running service needs the Bitrise token (rare — this is a CI-querying skill), fetch it from a managed secret manager at process startup; never bake it into code, config, or a container image:

```bash
aws secretsmanager get-secret-value --secret-id bitrise-token --query SecretString --output text  # needs GetSecretValue (+ kms:Decrypt for CMK)
gcloud secrets versions access <VERSION> --secret=bitrise-token                                    # needs roles/secretmanager.secretAccessor; pin a version
az keyvault secret show --vault-name <vault> --name bitrise-token --query value -o tsv             # needs Key Vault Secrets User (data role)
```

Authenticate with non-human, scoped identities (IAM instance/task roles, GCP service accounts, Azure Managed Identity), pin explicit versions, and rotate on a cadence.

## Hygiene

- Never commit the token to source control — not even a private repo; code is cloned, forked, and cached beyond access controls.
- Keep the fallback file outside any repo; if you must use `.env`, add it to `.gitignore` **before** the first commit and commit only a placeholder. `.gitignore` does **not** scrub history.
- Never pass the token as a literal CLI arg or `echo` / pipe it (lands in history, visible in `ps`); read into an env var only at call time via `$(...)`, and `unset` or scope to one command. Don't `set -x` while reading; clear the clipboard.
- Read into env vars is a fallback, not a best practice: child processes inherit them and they appear in crash dumps and `/proc/self/environ`.
- Run layered secret scanning — pre-commit hook (Gitleaks / detect-secrets) + CI diff scan + scheduled full-history scan (TruffleHog verifies live credentials). Pre-commit hooks are bypassable, so back them with CI and server-side push protection.
- On any suspected leak: **revoke/rotate immediately first** (Bitrise -> Account settings -> Security), only then rewrite history (`git filter-repo` / BFG), force-push, and have collaborators re-clone — see git-guide for the history-rewrite and force-push mechanics.
- Use a distinct token per integration so a compromise can be revoked in isolation; revoke when the task is done.

## Verify auth

A `GET /v0.1/me` returning 200 confirms the token works before any real call:

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: $TOKEN" https://api.bitrise.io/v0.1/me
```

- `200` — good.
- `401` — bad or revoked token, or a `Bearer ` / `token ` prefix snuck into the header.

## Discovering an app slug

The app slug identifies the Bitrise app (one per connected repo). Three ways to find it:

1. **From a build URL** — `https://app.bitrise.io/app/<APP_SLUG>/build/<BUILD_SLUG>`. The first slug is the app, the second the build.
2. **From the git host's build-status entry** — the `url` field of a commit's build status is exactly that build URL (see [builds.md](builds.md)).
3. **List your apps** —

```bash
curl -s -H "Authorization: $TOKEN" "https://api.bitrise.io/v0.1/apps?limit=50" \
  | python3 -c "import sys,json;[print(a['slug'],a['title']) for a in json.load(sys.stdin)['data']]"
```

Match the printed `title` to the app whose builds you need and use its `slug`. A project's concrete app slug, git-host, and workflow-name mappings belong in that project's own instructions (e.g. AGENTS.md), not in this skill.
