# First-time GitHub CLI setup

Get from a bare machine to a working `gh`: install, authenticate, pick a git protocol, clone, and verify with a read call. Token families and per-operation scopes are in [auth.md](auth.md); this file is the ordered first-run flow.

## 1. Install (prefer a package manager for auto-updates)

```bash
brew install gh                                  # macOS
sudo dnf install gh                              # Fedora / RHEL
winget install --id GitHub.cli                   # Windows
```

Debian / Ubuntu (official apt repo, so `gh` updates with the system):

```bash
sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
  | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null
sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
  | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
sudo apt update && sudo apt install gh

gh --version
```

## 2. Authenticate

```bash
gh auth login                                    # interactive (default)
```

The default is the **browser / device-code** flow, not a token paste. Answer **GitHub.com vs Other** (for GHES pass `--hostname ghe.example.com`), pick **HTTPS or SSH** for the git protocol, and answer **yes** to authenticate git so `gh` becomes the git credential helper. The token lands in the OS keyring.

Headless interactive setup — read a classic PAT without echoing or storing it in a scratch file:

```bash
IFS= read -r -s -p "GitHub classic PAT: " GITHUB_CLASSIC_PAT
printf '\n'
printf '%s' "$GITHUB_CLASSIC_PAT" | gh auth login --with-token
unset GITHUB_CLASSIC_PAT
```

Do NOT feed a fine-grained PAT to `--with-token` (its per-resource scoping confuses that flow) — inject it as `GH_TOKEN` from a secret store. In CI, configure `GH_TOKEN` in the runner's secret mechanism and validate that it exists without printing it: `: "${GH_TOKEN:?inject GH_TOKEN through the CI secret store}"`.

## 3. Git protocol

- HTTPS: if you authenticated by token/env rather than the browser flow, git-over-HTTPS will not use `gh` until you run `gh auth setup-git`.
- SSH: `gh auth login --git-protocol ssh` lets `gh` generate and upload a key (needs `admin:public_key` / `write:public_key`, which the browser flow requests but a hand-rolled `--with-token` PAT will not have — then use `--skip-ssh-key` and upload manually).

```bash
gh auth setup-git                                # make gh the git credential helper (HTTPS, token login)
```

## 4. Clone

```bash
gh repo clone OWNER/REPO
```

## 5. Verify with a read call

```bash
gh auth status                                   # exits 1 on an auth problem — a good CI gate
gh api user -q '.login'                          # real REST read
gh api graphql -f query='query{viewer{login}}'   # real GraphQL read
```

`gh auth status` exits non-zero on auth problems, but with `--json` it exits 0 even when broken — back it with a real read call.

## Enterprise (GHES)

```bash
gh auth login --hostname ghe.example.com
export GH_HOST=ghe.example.com
: "${GH_ENTERPRISE_TOKEN:?inject GH_ENTERPRISE_TOKEN through the secret store}"
gh api user --hostname ghe.example.com -q '.login'
```
