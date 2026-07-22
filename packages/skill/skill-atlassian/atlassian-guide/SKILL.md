---
name: atlassian-guide
description: "Use when operating Atlassian Cloud from the command line with the official `acli`: install it, authenticate Jira or Confluence with web OAuth or an API token, switch accounts, and search, view, create, transition, or comment on Jira work items. Triggers on `acli`, `acli jira`, `acli auth`, `*.atlassian.net`, JQL in a terminal, API-token authentication, or Jira work-item operations, even when the user doesn't say 'acli' and only says 'Jira' or 'Atlassian'."
---

# Atlassian CLI (acli) Guidelines

Operate on Atlassian **Cloud** from the shell with the official `acli`. This skill owns the CLI lifecycle end to end — install, authenticate, verify (see [references/first-time-setup.md](references/first-time-setup.md) and [references/auth.md](references/auth.md)) — plus the first Jira work-item operations. Jira is the first product covered; Confluence / admin operations can be added as references over time.

The one thing to internalize: **`acli` is Cloud-only and stores the credential in the OS keychain itself. When the browser OAuth flow is locked down by an admin, authenticate with an API token piped over stdin — nothing sensitive touches the command line or disk.**

When this skill fires:

1. Confirm the target is Atlassian **Cloud** (`*.atlassian.net`) — `acli` cannot talk to a self-hosted Server / Data Center instance.
2. Confirm auth (`acli jira auth status` shows `✓ Authenticated`) before any real command.
3. Load the `references/*.md` file matching the task, not everything upfront.

## Requirements

- `acli` on `PATH` — install (Homebrew tap-trust flow, other platforms) is in [references/first-time-setup.md](references/first-time-setup.md).
- An authenticated Jira Cloud profile — API token (default when web OAuth is admin-restricted) or web OAuth; credential handling is in [references/auth.md](references/auth.md).

## Essentials

- **Official CLI, Cloud only** - `acli` is Atlassian's first-party CLI for Jira / Confluence **Cloud** (`*.atlassian.net`); it does not support self-hosted Server / Data Center, see [references/first-time-setup.md](references/first-time-setup.md)
- **Install via the Atlassian tap** - `brew tap atlassian/acli` then `brew install acli`; newer Homebrew first needs `brew trust atlassian/acli`, see [references/first-time-setup.md](references/first-time-setup.md)
- **Authenticate with an API token** - Read the token with a hidden prompt or inject it from a CI secret store, then pipe it to `acli jira auth login --token` over stdin; the default when web OAuth (`--web`) is admin-restricted, see [references/auth.md](references/auth.md)
- **acli owns the secret** - the token is written to the OS keychain (macOS service `acli`); the `~/.config/acli/*.yaml` files hold only site / email / account (non-secret) — nothing to hand-store, see [references/auth.md](references/auth.md)
- **Verify** - `acli jira auth status` prints site, email, and auth type; run it before any real command, see [references/auth.md](references/auth.md)
- **Two auth surfaces** - `acli auth *` is global OAuth across sites; `acli jira auth *` is per-product and takes an API token — use the Jira one for token auth, see [references/auth.md](references/auth.md)
- **Jira work items** - `acli jira workitem search|view|create|transition|comment`; `--jql` to search, `--json` for machine output, `@me` to self-assign, see [references/jira.md](references/jira.md)
- **Switch accounts** - `acli jira auth switch --site <site> --email <you>` between profiles, see [references/auth.md](references/auth.md)

## Gotchas

- Newer Homebrew refuses a third-party tap until trusted: `brew install acli` errors `Refusing to load formula ... from untrusted tap` — run `brew trust atlassian/acli` first, then install.
- `acli` is **Cloud only**. It will not authenticate against a self-hosted Atlassian Server / Data Center host (for a self-hosted Bitbucket, that stays the REST-over-`curl` route — `bitbucket-guide`).
- Web OAuth (`acli jira auth login --web`) is frequently **disabled by org admins**. When it is, create an API token at `id.atlassian.com/manage-profile/security/api-tokens` and use the `--token` path — token creation is a per-user setting that some orgs also lock down.
- `--token` reads the value from **stdin only** — use a hidden prompt for interactive setup or CI secret injection for automation. Never type the value into a command, enable shell tracing around it, or create a plaintext scratch file.
- `acli auth` (global, OAuth, multi-site) and `acli jira auth` (per-product, API token or OAuth) are distinct — API-token login lives under `acli jira auth login`, not `acli auth login`.

## Example — install, authenticate with an API token, verify, search

```bash
# 1. install (one-time); trust the tap first on newer Homebrew
brew tap atlassian/acli
brew trust atlassian/acli          # only if `brew install` errors "untrusted tap"
brew install acli
acli --version

# 2. authenticate with an API token (default when web OAuth is admin-restricted)
#    token from https://id.atlassian.com/manage-profile/security/api-tokens
IFS= read -r -s -p "Atlassian API token: " ATLASSIAN_API_TOKEN
printf '\n'
printf '%s' "$ATLASSIAN_API_TOKEN" | acli jira auth login \
  --site <site>.atlassian.net --email you@example.com --token
unset ATLASSIAN_API_TOKEN

# 3. verify (acli stored the token in the OS keychain; config yaml holds only site/email)
acli jira auth status              # ✓ Authenticated / Site / Email / Authentication Type: api_token

# 4. first real call
acli jira workitem search --jql "assignee = currentUser() AND statusCategory != Done" --json
```

## Progressive Disclosure

Each reference is a trigger — read it only when the user's intent matches; do not preload everything.

- Read [references/first-time-setup.md](references/first-time-setup.md) - Load when installing `acli` for the first time (Homebrew tap + `brew trust`, other platforms), choosing between web OAuth and API-token login, or verifying a fresh install connects to a Cloud site.
- Read [references/auth.md](references/auth.md) - Load when authenticating or re-authenticating, creating / scoping / rotating an API token at id.atlassian.com, dealing with admin-restricted web OAuth, understanding where acli keeps the token (OS keychain) vs the config yamls, switching accounts, wiring auth into CI/CD, or handling a leaked token.
- Read [references/jira.md](references/jira.md) - Load when running Jira work-item operations: searching by JQL or filter, viewing, creating (summary / type / project / assignee / labels / parent, or from JSON), transitioning status, or listing / creating comments — including `--json` output.
