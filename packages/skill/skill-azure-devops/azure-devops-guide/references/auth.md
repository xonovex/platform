# CLI Credential Storage and Choice

## Contents

[Credential type and scope](#credential-type-and-scope) · [How the CLI sends it](#how-the-cli-sends-it) · [Verify before any real call](#verify-before-any-real-call) · [Local storage](#where-to-store-it-locally-keychain-first) · [CI/CD](#cicd) · [Cloud / production](#cloud--production) · [Hygiene](#hygiene-checklist)

Connecting the CLI — install, `az login`, the `azure-devops` extension, org/project defaults — is in [first-time-setup.md](first-time-setup.md). This file covers the concrete credential mechanics for the Azure DevOps Services CLI path: how to scope, store, send, verify, and rotate a credential when driving `az repos` / `az boards` or raw REST. The _why_ is owned elsewhere: prefer short-lived, federated credentials for the reasons in [service-connections.md](service-connections.md), and keep every token out of URLs, logs, history, and previews for the reasons in [rest-and-service-hooks.md](rest-and-service-hooks.md). What follows is the operational how.

## Credential type and scope

`az repos` / `az boards` authenticate as the signed-in identity. There are two ways to get one, preferred first.

An `az login` token cache is the preferred path. Interactive `az login` (device code or browser) leaves a short-lived, auto-refreshing token in the CLI cache — nothing to store or rotate. This is the default, and the connect steps are in [first-time-setup.md](first-time-setup.md). Use a non-human identity (managed identity / service principal) for automation, not a personal login.

A personal access token (PAT) is only for a flow that cannot use `az login`. Create it at `https://dev.azure.com/<org>/_usersSettings/tokens` → New Token and treat it like a password. Grant only the scopes the calls need — for PRs that is Code (Read & Write), for work-item links Work Items (Read & Write); never Full access, and scope the token to a single org where possible. Pick the shortest practical expiry (the maximum is one year, not "never") and rotate before it lapses; the value is shown once at creation, so capture it atomically into the keychain. Create, rotate, and revoke all live on that same `_usersSettings/tokens` page. On leak, revoke and reissue first — assume compromise within minutes — then update wherever it is stored.

A PAT can never exceed its owner's own access, and an org may disallow PATs by policy, in which case `az login` is the only path.

## How the CLI sends it

With `az login` active there is nothing to pass — `az repos pr create ...` just works against the cached token. For a PAT, the `azure-devops` extension reads the `AZURE_DEVOPS_EXT_PAT` environment variable; export it from a store at call time, never inline it:

```bash
export AZURE_DEVOPS_EXT_PAT="$(security find-generic-password -a "$USER" -s azure-devops-pat -w)"
az repos pr create --org https://dev.azure.com/<org> --project <project> --repository <repo> \
  --source-branch feat/x --target-branch main --title "feat: x" --description "..."
```

For raw REST (no extension), send the PAT as HTTP Basic auth with an empty username (`:<pat>`, base64-encoded):

```bash
PAT="$(security find-generic-password -a "$USER" -s azure-devops-pat -w)"
curl -s -u ":$PAT" "https://dev.azure.com/<org>/_apis/projects?api-version=7.1"
```

## Verify before any real call

Probe cheaply before the first mutating call so a bad credential fails fast. `200` is good; `203` or a `302` to a sign-in page means an unauthenticated or expired token; `401` is a bad or missing PAT; `403` is over-scoped or lacks write; `404` is a wrong org / project / repo.

```bash
curl -s -o /dev/null -w '%{http_code}\n' -u ":$PAT" \
  "https://dev.azure.com/<org>/_apis/connectionData?api-version=7.1"   # PAT path
az devops project list --org https://dev.azure.com/<org> >/dev/null && echo ok   # az login path
```

Azure DevOps returns `203` and an HTML sign-in redirect rather than a clean `401` for a bad token, so check for a JSON body, not just the status code.

## Where to store it locally (keychain-first)

A PAT is only for flows that cannot use `az login`. When you must keep one, prefer the OS keychain and resolve it at runtime so nothing lands on disk or in shell history. The ladder runs most to least preferred:

```bash
# 1. macOS Keychain (security CLI) — login keychain unlocks at GUI login
security add-generic-password -a "$USER" -s azure-devops-pat -U -w   # prompts on a hidden line; no value in history/ps
export AZURE_DEVOPS_EXT_PAT="$(security find-generic-password -a "$USER" -s azure-devops-pat -w)"
security delete-generic-password -a "$USER" -s azure-devops-pat       # delete / rotate

# 2. Linux Secret Service (secret-tool / libsecret) — gnome-keyring or KWallet
secret-tool store --label="Azure DevOps PAT" service azure-devops username "$USER"   # prompts for the value
export AZURE_DEVOPS_EXT_PAT="$(secret-tool lookup service azure-devops username "$USER")"
secret-tool clear service azure-devops username "$USER"

# 3. Windows — Credential Manager (cmdkey WRITES only; cannot read back)
cmdkey /generic:azure-devops-pat /user:%USERNAME% /pass:%pat%        # set /p pat=PAT: first; never a literal
#   Read AND write -> PowerShell SecretManagement + SecretStore:
#   Set-Secret -Name azure-devops-pat -Secret (Read-Host -AsSecureString)
#   $env:AZURE_DEVOPS_EXT_PAT = Get-Secret -Name azure-devops-pat -AsPlainText
#   WSL: no native keychain -> bridge to Windows Credential Manager via Git Credential Manager, or call powershell.exe Get-Secret
```

For teams and shared automation, fetch on demand from a secret-manager CLI instead of a per-machine keychain:

```bash
export AZURE_DEVOPS_EXT_PAT="$(az keyvault secret show --vault-name <vault> --name azure-devops-pat --query value -o tsv)"
export AZURE_DEVOPS_EXT_PAT="$(op read 'op://<vault>/azure-devops/pat')"      # 1Password
export AZURE_DEVOPS_EXT_PAT="$(vault kv get -mount=secret -field=pat azure-devops)"  # HashiCorp Vault (KV v2)
```

A gitignored `.env` is an acceptable fallback for local dev only: add `.env` / `.envrc` to `.gitignore` before the first commit, commit only a `.env.example` with placeholders, `chmod 600 .env`, and prefer referencing a vault inside it (`export AZURE_DEVOPS_EXT_PAT=$(az keyvault secret show ...)`) over storing the raw value. It is plaintext on disk, so prefer the keychain. Never inline the PAT in code, a committed config, or an MCP JSON, and never `echo` / `pbcopy` / `set -x` the value.

## CI/CD

Federation is the default here — [service-connections.md](service-connections.md) covers why, and how the cloud-side trust is constrained to specific claims (`sub` / `repo` / `ref` / `environment`) so forks and untrusted branches cannot impersonate the identity. Reserve a stored PAT for what genuinely cannot federate, inject it as a masked variable, and export it into `AZURE_DEVOPS_EXT_PAT` at runtime. Emphasis on this ecosystem, Azure first:

| Platform            | Stored-secret store                                                                | OIDC / short-lived                                                                                                 |
| ------------------- | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Azure Pipelines     | Secret variables / variable groups; **Key Vault-linked variable group**; `$(NAME)` | **Workload-identity-federation service connection** (no stored secret) — the preferred path                        |
| GitHub Actions      | Repo / Environment / Org secrets; `${{ secrets.NAME }}`                            | `permissions: id-token: write` → `azure/login` federated credential (issuer `token.actions.githubusercontent.com`) |
| GitLab CI/CD        | CI/CD variables (Masked + Protected), UI not YAML; `$VAR`                          | `id_tokens:` with `aud:` → federate to Azure                                                                       |
| Bitbucket Pipelines | Secured variables (workspace / repo / deploy); `$VAR`                              | `oidc: true` → `$BITBUCKET_STEP_OIDC_TOKEN`                                                                        |
| CircleCI            | Contexts + project env vars; `$VAR`                                                | `$CIRCLE_OIDC_TOKEN` (job must attach a context)                                                                   |
| Bitrise             | Secrets (encrypted, `[REDACTED]`); `$VAR`                                          | "Authenticate with Azure" steps; issuer `token.builds.bitrise.io`                                                  |

- For Azure Pipelines `az` tasks, the `AzureCLI@2` task with a federated service connection authenticates without any PAT in the pipeline.
- Prefer a Key Vault-linked variable group so rotation happens in Key Vault without touching pipeline config.
- Secrets are withheld from forked-repo / PR builds by default — expose them to PRs only with explicit justification, and add human approval gates on production-scoped service connections / environments.

## Cloud / production

Fetch the PAT (or service-principal secret) from a managed secret manager at process startup — never bake it into code, config, or a container image. Emphasis on Azure Key Vault for this ecosystem:

```bash
az keyvault secret show --vault-name <vault> --name azure-devops-pat --query value -o tsv   # needs Key Vault Secrets User (RBAC data role; Owner/Contributor is NOT enough)
aws secretsmanager get-secret-value --secret-id <name> --query SecretString --output text    # needs GetSecretValue (+ kms:Decrypt for CMK)
gcloud secrets versions access <VERSION> --secret=<id>                                        # needs roles/secretmanager.secretAccessor; pin a version, avoid `latest` in prod
vault kv get -mount=secret -field=pat azure-devops                                            # short-lived leases — re-auth in long jobs
```

- Better still, skip the stored PAT: an Azure VM or container with a Managed Identity can `az login --identity` and call Azure DevOps with no stored secret at all — see [service-connections.md](service-connections.md) for the federated-identity rationale.
- Grant the data-plane read permission on the secret itself (Key Vault Secrets User), not a management-plane role; Owner / Contributor does not grant secret reads.
- Pin explicit secret versions for reproducibility and rotate on a cadence. Never put the secret in request parameters — that hygiene principle is owned by [rest-and-service-hooks.md](rest-and-service-hooks.md).

## Hygiene checklist

The underlying principle — keep every token out of URLs, logs, history, fixtures, and error bodies — is owned by [rest-and-service-hooks.md](rest-and-service-hooks.md). The concrete checklist for a CLI PAT:

- Never commit a PAT to source control, not even a private repo; code is cloned, forked, and cached beyond its access controls.
- Add `.env` / `.envrc` to `.gitignore` before the first commit and commit only `.env.example`. `.gitignore` does not scrub history.
- Never pass the PAT as a literal CLI arg or `echo` / pipe it — it lands in shell history and is visible in `ps`. Use the hidden-prompt form (`security add-generic-password ... -w`) or `read -rs … ; unset`. Read it into `AZURE_DEVOPS_EXT_PAT` only at call time via `$(...)`, scoped to one command or `unset` afterward.
- Treat log masking as best-effort — it breaks on transform / encode (base64 / URL) or child-process printing. Do not `set -x` while reading, and clear the clipboard.
- Env vars are a fallback, not a best practice: child processes inherit `AZURE_DEVOPS_EXT_PAT`, and it appears in crash dumps and `/proc/self/environ`.
- Run layered secret scanning: a pre-commit hook (Gitleaks / detect-secrets) plus a CI diff scan plus a scheduled full-history scan (TruffleHog verifies live credentials). Pre-commit hooks are bypassable (`--no-verify`), so back them with CI and server-side push protection.
- Use a distinct PAT per integration so a compromise can be revoked in isolation, and least-privilege who can read the secret — engineers should not all have access, especially in production.
- On any suspected leak, revoke and reissue the PAT first at `_usersSettings/tokens`; only then rewrite history (`git filter-repo` / BFG, the mechanics of which are in git-guide), force-push, have collaborators re-clone, and coordinate fork / PR-cache cleanup.
