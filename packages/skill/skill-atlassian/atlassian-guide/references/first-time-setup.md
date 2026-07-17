# First-time setup: Install and connect the Atlassian CLI

Get from a bare machine to a working `acli`: install it, authenticate a Jira Cloud site, and verify. `acli` is Atlassian's **official** CLI and the gh-equivalent for Jira / Confluence **Cloud** — there is no self-hosted Server / Data Center support, and the various third-party `jira` / `bb` CLIs are unrelated. Credential storage and the API-token-vs-web-OAuth choice are in [auth.md](auth.md).

## Install acli

`acli` ships from Atlassian's own Homebrew tap (not homebrew-core, and there is no `acli` cask). Add the tap, then install.

```bash
brew tap atlassian/acli
brew install acli
acli --version        # e.g. acli version 1.3.22-stable
```

- **Newer Homebrew blocks untrusted third-party taps.** If `brew install acli` errors `Refusing to load formula atlassian/acli/acli from untrusted tap`, trust the tap once, then re-install:

  ```bash
  brew trust atlassian/acli        # whole-tap trust; or: brew trust --formula atlassian/acli/acli
  brew install acli
  ```

  Whole-tap trust covers all current and future formulae from that tap; per-formula trust is narrower. (`HOMEBREW_NO_REQUIRE_TAP_TRUST=1` disables the check globally — not recommended.)

- **Other platforms.** Atlassian distributes `acli` for Linux (deb / rpm / tarball) and Windows (msi / winget) from `https://developer.atlassian.com/cloud/acli/` — follow the platform installer there; the authenticate / verify steps below are identical once `acli` is on `PATH`.

Verify the top-level surface: `acli --help` lists `auth`, `jira`, `confluence`, `admin`, `guard`, `rovodev`.

## Authenticate

`acli` has two auth surfaces (details in [auth.md](auth.md)):

- `acli auth login` — **global OAuth** across Jira + Confluence sites (browser flow).
- `acli jira auth login` — **per-product**, and the only path that accepts an **API token**.

Prefer the API-token path — the browser OAuth flow is often disabled by org admins.

```bash
# API token (default when web OAuth is admin-restricted)
# create at https://id.atlassian.com/manage-profile/security/api-tokens
echo "$ATLASSIAN_API_TOKEN" | acli jira auth login \
  --site <site>.atlassian.net --email you@example.com --token
# OR read from a file (keeps it out of shell history)
acli jira auth login --site <site>.atlassian.net --email you@example.com --token < token.txt

# Web OAuth (only if the admin allows it) — opens a browser
acli jira auth login --web
```

The `--token` flag reads the value from **stdin** — never pass it as a literal flag argument. `acli` writes the token to the OS keychain and records the non-secret profile (site, email, account, `auth_type`) under `~/.config/acli/` — see [auth.md](auth.md).

## Verify

```bash
acli jira auth status        # ✓ Authenticated / Site / Email / Authentication Type: api_token
acli jira project list       # confirm the token actually reaches Jira
```

`✓ Authenticated` plus a successful `project list` confirms both the credential and site access. A `401` / unauthenticated error on the first command usually means a bad or revoked token, the wrong site, or an email that does not match the token owner.

## Gotchas

- `acli` targets Atlassian **Cloud** only. It cannot authenticate against a self-hosted Server / Data Center host — for a self-hosted Bitbucket, use the REST-over-`curl` route (`bitbucket-guide`).
- Web OAuth (`--web`) needs a browser and is frequently disabled by org policy; when it is, the API-token path is the only option, and some orgs also restrict per-user token creation at `id.atlassian.com`.
- API-token login is under `acli jira auth login` (per-product), **not** `acli auth login` (global OAuth). Using the global one when you meant token auth is a common mix-up.
- The site is the bare host `<site>.atlassian.net`, not a full URL — do not prefix `https://`.
