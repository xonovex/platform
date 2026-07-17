# First-time connection to a self-hosted Bitbucket Server / Data Center

One-time setup for a new machine / account against a self-hosted Bitbucket Server / Data Center instance: generate an SSH key for git, register it, add a `~/.ssh/config` host alias, clone over SSH, then mint an HTTP access token for the REST 1.0 API at `/rest/api/1.0`. git over SSH and REST over HTTPS are independent auths — set up both. Resolve the concrete host / project / alias / port from your instance's coordinates; where to STORE the token lives in [auth.md](auth.md). This is Data Center-specific — Bitbucket Cloud uses app passwords / API tokens over `api.bitbucket.org` instead, so do not assume this setup carries over.

## Prerequisites

```bash
git --version        # any recent git
ssh -V               # OpenSSH client (ships ssh-keygen; bundled with Git / Git Bash)
ls -al ~/.ssh        # reuse an existing key if id_ed25519(.pub) is already there
```

"No such file or directory" from the `ls` means no key yet — generate one. An admin must have enabled SSH access to repositories on the instance first.

## Generate an SSH key

```bash
ssh-keygen -t ed25519 -C "you@example.com"   # Enter for default ~/.ssh/id_ed25519, then set a passphrase
```

The `-C` comment is just a label to identify the key later. Ed25519 is preferred; the server also accepts RSA/ECDSA and hardware keys (`ed25519-sk` / `ecdsa-sk`), and an admin may mandate a key type or minimum length. The private key (`~/.ssh/id_ed25519`) never leaves the machine — only ever upload the `.pub`.

## Register the public key

Copy the **public** key:

```bash
pbcopy < ~/.ssh/id_ed25519.pub             # macOS
xclip -sel clip < ~/.ssh/id_ed25519.pub    # Linux
clip < ~/.ssh/id_ed25519.pub               # Windows / Git Bash
```

In Bitbucket: avatar → **Manage account** → **SSH keys** → **Add key**. Paste, optionally set a label and an expiry, then **Add key**. Bitbucket shows an uneditable SHA-256 fingerprint. Key expiry is set once and cannot be edited afterwards.

## Add a `~/.ssh/config` host alias

A host alias pins the right key, host, and SSH port so git and `ssh` don't guess. Server / Data Center listens on port **7999** by default, not 22:

```
Host <alias>
    HostName <host>
    User git
    Port 7999
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
```

`IdentitiesOnly yes` forces SSH to offer only this key — without it ssh-agent may present the wrong key first and the clone fails with "repository does not exist or you do not have access" even though it exists. Omit the `Port` line only if the admin forwards 22 → 7999.

## Verify SSH and clone

```bash
ssh -T <alias>                                          # or: ssh -T git@<host> -p 7999
git clone ssh://git@<host>:7999/<PROJECT>/<repo>.git    # explicit form
git clone <alias>:<PROJECT>/<repo>.git                  # scp-style, via the alias
```

Bitbucket SSH gives no interactive shell, so a successful `ssh -T` returns immediately with a usage banner rather than a prompt. Prefer the exact SSH URL from the repo's **Clone** button — it embeds the admin-configured SSH base URL and port and is authoritative over a hand-built one. Forgetting `:7999` (or the `Port` line) is the most common clone failure.

## Create an HTTP access token

git's SSH key does **not** authenticate the REST API; mint a separate HTTP access token. Avatar → **Manage account** → **HTTP access tokens** → **Create token**. Set a name, scope **Permissions** to the minimum (e.g. Repository read/write for the project/repos you'll hit), and set an expiry — expiry cannot be edited after creation, so pick the shortest practical one. Use a distinct token per integration so one can be revoked in isolation.

Where to STORE that token (keychain-first, `~/.netrc`, CI) is in [auth.md](auth.md) — do not keep it in plaintext or hand-edit it into a config.

## Verify the REST API

A personal/user token works with either Basic auth (username + token) or Bearer. A project- or repository-scoped token uses Bearer auth with **no** username on Bitbucket DC < 9.4 (Basic `-u user:token` `401`s there); DC ≥ 9.4 relaxed this so Basic also works for repo tokens. When unsure, prefer Bearer for a repo/project token. See [auth.md](auth.md) for the send paths.

```bash
curl -u <username>:<TOKEN> https://<host>/rest/api/1.0/projects                       # personal token
curl -H "Authorization: Bearer <TOKEN>" https://<host>/rest/api/1.0/projects          # project/repo token
```

A `200` confirms read; a `403` on the first write means the token lacks the write scope. An HTTP access token cannot log into the web UI, cannot edit its owner's account/tokens, and cannot merge a PR.

## Per-repo identity (optional)

On a machine with multiple accounts, set identity inside the clone instead of globally (see the git-guide skill for identity and commit conventions):

```bash
git config user.name "..."
git config user.email "..."
```
