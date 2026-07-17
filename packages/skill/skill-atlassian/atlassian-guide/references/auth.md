# auth: Authenticate acli to Atlassian Cloud

Installing and connecting `acli` (tap, `brew trust`, first login, verify) is in [first-time-setup.md](first-time-setup.md). This file covers the **credential**: the two auth surfaces, the API-token path when web OAuth is admin-restricted, where `acli` keeps the secret, switching accounts, CI/CD, and hygiene.

## Two auth surfaces

`acli` exposes two independent login paths:

- **`acli auth login` — global OAuth.** Browser flow, authenticates Jira **and** Confluence sites at once, tokens auto-refresh. No secret for you to store. Requires a browser and org policy that permits it.
- **`acli jira auth login` — per-product, and the only path that accepts an API token.** Use this for headless / restricted environments and whenever web OAuth is disabled.

`acli jira auth status` shows which is active (`Authentication Type: api_token` or `oauth`). `acli auth status` shows the global OAuth accounts.

## Credential type and scope

The API token is a personal Atlassian credential — treat it like a password.

1. **Create / rotate / revoke** at `https://id.atlassian.com/manage-profile/security/api-tokens` → **Create API token**. This is a per-user setting (no admin needed) — but some orgs disable it too; if both web OAuth and token creation are locked, there is no `acli` path and you escalate to the admin.
2. **Least privilege.** A basic API token inherits the owner's own Jira / Confluence permissions — it cannot exceed them. `acli` also supports scoped tokens (the `--token` login has a scopes variant); prefer the narrowest that covers the calls you make.
3. **Short-lived.** Atlassian API tokens can be given an expiry — pick the shortest practical one and rotate before it lapses. The value is shown **once** at creation; capture it straight into a store, not a scratch file.

## How this skill sends it

The token is read from **stdin** by the `--token` flag — never a literal CLI argument (which lands in shell history and `ps`):

```bash
# pipe from an env var resolved from a store (see "Where to store it")
echo "$ATLASSIAN_API_TOKEN" | acli jira auth login \
  --site <site>.atlassian.net --email you@example.com --token

# or redirect a file, then remove it
acli jira auth login --site <site>.atlassian.net --email you@example.com --token < token.txt
```

After a successful login there is nothing to pass on subsequent commands — `acli jira workitem ...` uses the stored credential automatically.

**Verify before any real call:**

```bash
acli jira auth status                 # ✓ Authenticated + site/email/type
acli jira project list >/dev/null && echo ok    # exercises the token against Jira
```

`✓ Authenticated` confirms a stored credential; the `project list` confirms it actually reaches the site. An unauthenticated / `401` error means a bad, expired, or revoked token, the wrong `--site`, or an `--email` that does not own the token.

## Where acli keeps the token (OS keychain — native)

Unlike a raw-`curl` workflow, **you do not store the token yourself** — `acli` writes it to the OS secret store on login:

- **macOS:** the login Keychain, service `acli`, account `jira:<cloud_id>:<account_id>`. Inspect (not the secret) with `security find-generic-password -s acli`.
- **Linux:** the Secret Service (libsecret / gnome-keyring or KWallet) where available.
- **Windows:** Credential Manager.

The `~/.config/acli/*.yaml` files (`jira_config.yaml`, `global_auth_config.yaml`, …) hold only **non-secret** profile data — site, `cloud_id`, `account_id`, email, `auth_type`, current profile. They are `0600` and safe to read for debugging, but **do not commit** them (they identify accounts/sites). The secret is never written there.

`acli jira auth logout` clears the Jira credential; `acli auth logout` clears all global OAuth accounts.

If you must hand-hold the raw token before piping it in (e.g. a shared machine), keep it in the OS keychain and resolve it at call time rather than in a plaintext file:

```bash
# macOS — store once (hidden prompt, nothing in history/ps), read at login time
security add-generic-password -a "$USER" -s atlassian-api-token -U -w
echo "$(security find-generic-password -a "$USER" -s atlassian-api-token -w)" \
  | acli jira auth login --site <site>.atlassian.net --email you@example.com --token
# Linux Secret Service
secret-tool store --label="Atlassian API token" service atlassian username "$USER"
secret-tool lookup service atlassian username "$USER" \
  | acli jira auth login --site <site>.atlassian.net --email you@example.com --token
```

## Switch accounts

```bash
acli jira auth switch                                    # interactive picker
acli jira auth switch --site <site>.atlassian.net        # by site
acli jira auth switch --email you@example.com            # by email
acli jira auth switch --site <site>.atlassian.net --email you@example.com
```

Each authenticated site/email is a profile in `jira_config.yaml`; `switch` flips the `current_profile`.

## CI/CD

Web / device OAuth needs a browser, so CI authenticates with an **API token** (ideally a service-account token, not a personal one). Inject it as a masked secret and pipe it into `acli jira auth login --token` at the start of the job; never write it into committed pipeline YAML. Default to short-lived / scoped tokens and rotate in the store, not in config.

| Platform            | Stored-secret store                                       | Notes                                                   |
| ------------------- | --------------------------------------------------------- | ------------------------------------------------------- |
| Bitbucket Pipelines | Secured variables (workspace / repo / deploy); `$VAR`     | pipe `$VAR` into `acli jira auth login --token`         |
| GitHub Actions      | Repo / Environment / Org secrets; `${{ secrets.NAME }}`   | withheld from fork PRs by default                       |
| GitLab CI/CD        | CI/CD variables (Masked + Protected, UI not YAML); `$VAR` | mark Protected so only protected branches see it        |
| Azure Pipelines     | Secret variables / Key Vault-linked variable group        | prefer a Key Vault-linked group so rotation is external |
| CircleCI / Bitrise  | Contexts / Secrets (`[REDACTED]`); `$VAR`                 | scope contexts to the jobs that need Jira access        |

```bash
# CI step — token comes from a masked secret variable
printf '%s' "$ATLASSIAN_API_TOKEN" | acli jira auth login \
  --site "$JIRA_SITE" --email "$JIRA_EMAIL" --token
acli jira auth status
```

- Use a dedicated service-account token per pipeline so a compromise is revoked in isolation.
- Secrets are withheld from forked-repo / PR builds by default — only expose with explicit justification, and isolate PR steps.

## Cloud / production

If automation runs as a service, fetch the token from a managed secret manager at process startup and pipe it into login — never bake it into code, config, or a container image:

```bash
aws secretsmanager get-secret-value --secret-id atlassian-api-token --query SecretString --output text \
  | acli jira auth login --site "$JIRA_SITE" --email "$JIRA_EMAIL" --token   # needs GetSecretValue (+ kms:Decrypt for CMK)
az keyvault secret show --vault-name <vault> --name atlassian-api-token --query value -o tsv \
  | acli jira auth login --site "$JIRA_SITE" --email "$JIRA_EMAIL" --token   # needs Key Vault Secrets User (RBAC data role)
vault kv get -mount=secret -field=token atlassian \
  | acli jira auth login --site "$JIRA_SITE" --email "$JIRA_EMAIL" --token   # short-lived leases — re-auth in long jobs
```

- Authenticate with a non-human service account, scoped to the minimum Jira / Confluence permissions.
- Pin explicit secret versions for reproducibility; rotate on a cadence and automate it.

## Hygiene

- Never pass the token as a literal CLI arg or `echo` it to the terminal — it lands in shell history and is visible in `ps`. Pipe from a store or a redirected file, and delete the file after.
- Never commit `~/.config/acli/*.yaml` — non-secret but account/site-identifying — and never commit a token file; add `token.txt` / `.env` / `.envrc` to `.gitignore` **before** the first commit.
- The token lives in the OS keychain after login — do not also copy it into a dotfile "for convenience".
- Log masking is best-effort; don't `set -x` while piping the token, and clear the clipboard if you pasted it.
- On any suspected leak: **revoke / reissue the token first** at `id.atlassian.com/manage-profile/security/api-tokens` (it is the only thing that stops the bleeding), then `acli jira auth login` again with the new value and update every store that held the old one.
