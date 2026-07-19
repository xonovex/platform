# auth: Authenticate to the Figma REST API

## Contents

[Credential type and scope](#credential-type-and-scope) · [How this skill sends it](#how-this-skill-sends-it) · [Local storage](#where-to-store-it-local--keychain-first) · [CI/CD](#cicd) · [Cloud / production](#cloud--production) · [Hygiene](#hygiene)

## Credential type and scope

Figma REST API calls authenticate with a **personal access token** (create at Figma → Settings → Security → Personal access tokens). Treat it like a password — it grants API access across the whole account.

- **Least privilege.** Scope the token to only what the endpoints you call need. Most read flows need `file_content:read` (or the legacy `file_read`); the variables endpoint (`GET /v1/files/:key/variables/local`) needs `file_variables:read`. Skip write scopes for read-only design-to-code work.
- **Short-lived.** Pick the shortest practical expiry (not "never"); rotate before it lapses. Figma shows the token value **once** at creation — capture it atomically into the keychain.
- **Where to create / rotate / revoke.** Figma → Settings → Security → Personal access tokens. On leak: **revoke + reissue first** (assume compromise within minutes), then update wherever it is stored.

> `file_variables:*` scopes (and the local-variables endpoint) are Enterprise-plan only; the styles and components endpoints work on all plans.

## How this skill sends it

Read the token at call time and feed it straight into the request — never inline it, never let it hit shell history:

```bash
TOKEN=$(security find-generic-password -s figma-token -w)
curl -s -H "X-Figma-Token: $TOKEN" "https://api.figma.com/v1/files/<file-key>"
```

Get the scheme exactly right — Figma uses two different ones:

| Scheme                             | Header                                                       |
| ---------------------------------- | ------------------------------------------------------------ |
| Personal access token (this skill) | `X-Figma-Token: <token>` (raw token, **no** `Bearer` prefix) |
| OAuth 2.0 access token             | `Authorization: Bearer <access-token>`                       |

OAuth-based Figma integrations send `Authorization: Bearer` for apps acting on behalf of other users; personal scripts stay on the personal access token + `X-Figma-Token`. Do not conflate the two — sending a PAT as `Bearer` (or vice versa) yields `401`.

**Verify before any real call** (`200` good, `401` bad/missing/wrong-prefix token, `403` over-scoped/wrong plan, `404` wrong file key):

```bash
curl -s -o /dev/null -w '%{http_code}\n' -H "X-Figma-Token: $TOKEN" "https://api.figma.com/v1/me"
```

For MCP/server configs that only accept a static value, launch via a wrapper that `export`s the token from one of the stores below — do not paste the literal.

## Where to store it (local — keychain-first)

This skill runs locally (design-to-code scripts on a developer machine) and occasionally in CI. Prefer the OS keychain for local development; resolve at runtime so nothing lands on disk or in history. Ladder, most to least preferred:

```bash
# 1. macOS Keychain (security CLI) — login keychain unlocks at GUI login
security add-generic-password -a "$USER" -s figma-token -U -w   # prompts on a hidden line; no value in history/ps
export TOKEN="$(security find-generic-password -a "$USER" -s figma-token -w)"   # read at call time
security delete-generic-password -a "$USER" -s figma-token       # delete / rotate

# 2. Linux Secret Service (secret-tool / libsecret) — gnome-keyring or KWallet
secret-tool store --label="Figma token" service figma username "$USER"    # prompts for the value
export TOKEN="$(secret-tool lookup service figma username "$USER")"
secret-tool clear service figma username "$USER"

# 3. Windows — Credential Manager (cmdkey WRITES only; cannot read back)
cmdkey /generic:figma-token /user:%USERNAME% /pass:%tok%         # set /p tok=Token: first; never a literal
#   Read AND write -> PowerShell SecretManagement + SecretStore:
#   Set-Secret -Name figma-token -Secret (Read-Host -AsSecureString)
#   $env:TOKEN = Get-Secret -Name figma-token -AsPlainText
#   WSL: no native keychain -> use the Linux Secret Service path above, or bridge to Windows Credential Manager
```

For **teams / shared automation**, fetch on demand from a secret-manager CLI instead of a per-machine keychain:

```bash
export TOKEN="$(op read "op://<vault>/Figma/token")"            # 1Password (OP_SERVICE_ACCOUNT_TOKEN in CI)
export TOKEN="$(vault kv get -mount=secret -field=token figma)" # HashiCorp Vault (KV v2; use -mount)
export TOKEN="$(doppler secrets get FIGMA_TOKEN --plain)"       # Doppler (project/config scoped)
```

**Gitignored `.env` fallback** — acceptable for local dev only:

- Add `.env` / `.envrc` to `.gitignore` **before** the first commit; commit only a `.env.example` with placeholders.
- `chmod 600 .env`; better, reference a vault inside it (`export FIGMA_TOKEN=$(op read ...)`) rather than storing the raw value.
- Plaintext on disk — prefer the OS keychain. With `direnv`, run `direnv allow` (re-approve after every edit) and parse via the `dotenv` stdlib, not `source`.

All paths resolve to the same `X-Figma-Token: $TOKEN` header — only where the value comes from differs. **Never** inline the literal in code, a committed config, or an MCP JSON; never `echo`/`pbcopy`/`set -x` the value. See Hygiene.

## CI/CD

Design-to-code scripts run in CI occasionally. A Figma PAT cannot be federated, so it is a genuine stored secret — inject it as a masked env var and pass it into the `X-Figma-Token` header at runtime. For _other_ secrets in the same pipeline (cloud logins, registries), prefer short-lived OIDC / workload identity federation over long-lived keys, and constrain the cloud-side trust policy to specific `repo`/`ref`/`environment` claims.

| Platform            | Stored-secret store                                                          | OIDC / short-lived (for federatable secrets)                                                       |
| ------------------- | ---------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| GitHub Actions      | Repo / Environment / Org secrets; `${{ secrets.FIGMA_TOKEN }}`               | `permissions: id-token: write` → cloud login action (issuer `token.actions.githubusercontent.com`) |
| GitLab CI/CD        | CI/CD variables (Masked + Protected + hidden), UI not YAML; `$FIGMA_TOKEN`   | `id_tokens:` with `aud:` → federate to AWS/GCP/Azure/Vault                                         |
| Azure Pipelines     | Secret variables / variable groups; Key Vault-linked group; `$(FIGMA_TOKEN)` | Workload-identity-federation service connection                                                    |
| Bitbucket Pipelines | Secured variables (workspace/repo/deploy); `$FIGMA_TOKEN`                    | `oidc: true` → `$BITBUCKET_STEP_OIDC_TOKEN`                                                        |
| CircleCI            | Contexts + project env vars; `$FIGMA_TOKEN`                                  | `$CIRCLE_OIDC_TOKEN` (job MUST attach a context)                                                   |
| Bitrise             | Secrets (encrypted, `[REDACTED]`); `$FIGMA_TOKEN`                            | "Authenticate with AWS/GCP" steps; issuer `token.builds.bitrise.io`                                |

- Secrets are withheld from forked-repo / PR builds by default — only expose to PRs with explicit justification, and isolate PR steps.
- Prefer Key Vault / Vault-backed indirection so PAT rotation happens in the vault without touching pipeline config.
- Add human approval gates on production-scoped environments.

## Cloud / production

If a service reads Figma in production, fetch the PAT from a managed secret manager at process startup (env injection or explicit read); never bake it into code, config, or container images. This keeps one source of truth, picks up rotated values automatically, and limits blast radius.

```bash
aws secretsmanager get-secret-value --secret-id figma-token --query SecretString --output text   # needs GetSecretValue (+ kms:Decrypt for CMK)
aws ssm get-parameter --name /figma/token --with-decryption --query Parameter.Value --output text # SecureString; lighter than Secrets Manager
gcloud secrets versions access <VERSION> --secret=figma-token                                      # needs roles/secretmanager.secretAccessor; pin a version, avoid `latest` in prod
az keyvault secret show --vault-name <vault> --name figma-token --query value -o tsv               # needs Key Vault Secrets User (RBAC data role; Owner/Contributor is NOT enough)
vault kv get -mount=secret -field=token figma                                                      # short-lived lease — re-auth in long jobs
```

- **IAM least privilege, data-plane specific.** Grant the read permission on the secret itself, not management-plane roles.
- **Authenticate with non-human, scoped identities** — IAM instance/task roles, GCP service accounts, Azure Managed Identity, Vault AppRole — not personal logins.
- **Pin explicit versions** for reproducibility; use version stages (`AWSCURRENT`/`AWSPREVIOUS`) for safe rotation/rollback.
- **Rotate on a cadence** and automate it. Every read may be audit-logged (e.g. CloudTrail) — never put the token in request parameters.

## Hygiene

- Never commit the token to source control — not even private repos; code is cloned, forked, and cached beyond access controls.
- Add `.env` / `.envrc` to `.gitignore` (a global `~/.gitignore` for `.envrc`) **before** the first commit; commit only `.env.example`. `.gitignore` does **not** scrub history.
- Never pass the token as a literal CLI arg or `echo`/pipe it (lands in history, visible in `ps`); use interactive prompts or `read -rs … ; unset`. Read into env vars only at call time via `$(...)`; scope to one command or `unset` after.
- Treat log masking as best-effort everywhere — it breaks on transform/encode (base64/URL) or child-process printing. Don't `set -x` while reading; clear the clipboard.
- Env vars are a fallback, not a best practice: child processes inherit them, they appear in crash dumps/logs, and `/proc/self/environ` dumps them.
- Run layered secret scanning: pre-commit hook (Gitleaks / detect-secrets) + CI diff scan + scheduled full-history scan (TruffleHog verifies live credentials). Pre-commit hooks are bypassable (`--no-verify` / `SKIP=`) — back them with CI and server-side push protection.
- Use a distinct token per integration so a compromise can be revoked in isolation.
- On any suspected leak: **revoke/rotate immediately first** (Figma → Settings → Security → Personal access tokens), then update every store; only then rewrite history (`git filter-repo` / BFG), `reflog expire --expire=now --all && gc --prune=now`, force-push, and have collaborators re-clone.

**Related:** [read-nodes.md](./read-nodes.md), [render-images.md](./render-images.md), [variables-styles.md](./variables-styles.md)
