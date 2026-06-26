# First-time setup: install and connect glab

Get from a bare machine to a working GitLab CLI: install `glab`, authenticate, choose the protocols, clone, and verify with a read call. Token type and exact scopes are in [auth.md](auth.md); this file is the ordered first-run flow. glab officially supports GitLab 16.0+.

## 1. Install

```bash
brew install glab            # the only officially supported installer
glab version                 # confirm
```

Community packages exist (snap, dnf, pacman, nixpkgs) but may lag the official release. After install, `glab version` should print a 1.x (or 2.x) version.

## 2. Authenticate

```bash
# gitlab.com — run inside a repo so glab auto-detects the instance from the git remote
glab auth login

# self-managed / Dedicated — --hostname is REQUIRED or glab silently targets gitlab.com
glab auth login --hostname gitlab.example.com
```

Pick a flow at the prompt:

- **Web** (browser) — default interactive.
- **`--device`** — OAuth device flow (needs GitLab 17.9+), ideal for headless / SSH sessions.
- **Non-interactive** — `glab auth login --hostname H --stdin < token.txt` (preferred over `--token`, which lands in shell history).

Add `--use-keyring` to store the token in the OS keyring instead of plaintext `~/.config/glab-cli/config.yml`. Token scopes and types are in [auth.md](auth.md) — at minimum the token needs `api` and `write_repository` for full read/write use, or `read_api` for read-only.

## 3. Choose protocols (separate per-host settings)

`git_protocol` (how git clones/pushes) and `api_protocol` (how glab calls the API) are independent — a working ssh clone does NOT imply API calls work.

```bash
glab config set -h gitlab.example.com git_protocol ssh
glab config set -h gitlab.example.com api_protocol https
```

In minimal containers use `git_protocol https` because the ssh binary may be absent.

## 4. Clone

```bash
glab repo clone group/subgroup/project    # by path
glab repo clone 12345678                   # or by numeric project ID
```

## 5. Verify (the read call)

```bash
glab auth status                                   # logged-in user, host, REST /api/v4/ + GraphQL endpoints, protocols
glab api user                                      # GET /api/v4/user — confirms the token identity
glab mr list --assignee=@me -R group/project       # review smoke test
```

Run `glab auth status` immediately before any write to confirm the intended identity on the intended host — a stale env token silently acts as the wrong user (env tokens override stored config, see [auth.md](auth.md)).
