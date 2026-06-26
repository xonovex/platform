# GitHub auth — token families and per-operation least-privilege scopes

`gh` and `gh api` authenticate with a token. There are two families with different scope models; pick the narrowest scope for the operation, scope it per endpoint, and store it in the OS keyring. Onboarding (the interactive `gh auth login` flow) is in [onboarding.md](onboarding.md); this file is the token + scope + storage reference.

## Two token families

### Classic PAT — coarse, all-or-nothing scopes

- The single **`repo`** scope covers the whole lifecycle: push, open a PR, post/submit a review, and resolve a thread. Use **`public_repo`** for public-only repos.
- `gh auth login --with-token` (and the browser flow) wants a classic PAT with at least **`repo`, `read:org`, `gist`**. SSH key upload also needs **`admin:public_key`** (the browser flow requests it; a hand-rolled `--with-token` PAT will not have it).
- Classic PATs work everywhere fine-grained PATs do not: public repos where you are a non-member, outside-collaborator access, and tokens that must span multiple orgs.

### Fine-grained PAT — per-repo, per-permission (preferred where it fits)

Grant only the permissions the operation needs. Per the GitHub fine-grained-PAT permissions reference:

| Operation                                                                           | Required fine-grained permission                        |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------- |
| Push commits / create / update refs (`POST/PATCH .../git/refs`, `PUT .../contents`) | **Contents: write**                                     |
| Open a PR (`POST .../pulls`)                                                        | **Pull requests: write** + Contents: read               |
| Post / submit a review (`POST .../pulls/{n}/reviews`)                               | **Pull requests: write**                                |
| Top-level PR conversation comment (a PR is an issue)                                | **Issues: write**                                       |
| Resolve a thread (`resolveReviewThread`, GraphQL)                                   | **Pull requests: write** AND **Contents: read & write** |

- **Contents: read is NOT enough to push** — pushing and ref creation are Contents: write operations. Contents: read suffices only to open a PR and post a review.
- `Pull requests: write` subsumes read; `Metadata: read` is auto-required on every fine-grained token.
- `resolveReviewThread` needs the non-obvious **Contents: read & write** on top of Pull requests: write, or it fails "Resource not accessible by integration" (community #44650). Classic `repo` has no such gap.
- **Fine-grained gaps (2025–2026):** cannot contribute to public repos where you are a non-member, cannot be used by outside collaborators, cannot span multiple orgs — use a classic PAT for those. Org-approval-required tokens sit "pending" (read public only, 403/404 on writes) until an admin approves.

### Verify the exact permission an endpoint wants

Don't guess — read the response header:

```bash
gh api -i repos/{owner}/{repo}/pulls/123/reviews 2>&1 | grep -i x-accepted-github-permissions
# X-Accepted-GitHub-Permissions: pull_requests=write,contents=read
```

Comma = AND, semicolon = alternatives. On a private repo, insufficient permission returns **404 (not 403)** — GitHub hides existence; this header still tells you what was required.

## Environment variables (host matters)

- **github.com and `*.ghe.com`:** `GH_TOKEN` or `GITHUB_TOKEN` (GH_TOKEN wins).
- **GitHub Enterprise Server (self-managed):** `GH_ENTERPRISE_TOKEN` or `GITHUB_ENTERPRISE_TOKEN`. Mixing these with GH_TOKEN is the classic "works on github.com, 401 on GHES".
- **`GH_HOST`** sets the default host; per-command `--hostname ghe.example.com` overrides it.

A token in any of these env vars takes precedence over the keyring and disables interactive `gh auth login` for that host.

## Storage (keyring-first)

- `gh` stores the token in the **OS keyring** by default, falling back to plaintext config only when no keyring is present. Never pass `--insecure-storage` on a shared machine.
- Resolve secrets at call time from a secret manager rather than inlining them:

```bash
export GH_TOKEN="$(op read 'op://<vault>/github/token')"          # 1Password
export GH_TOKEN="$(vault kv get -mount=secret -field=token github)" # HashiCorp Vault
```

- Never commit a token (not even to a private repo), never pass it as a literal CLI arg or `echo` it (shell history / `ps`), and use a distinct token per integration so a leak is revoked in isolation. On suspected leak, **revoke/reissue first**, then update every store.

## Self-managed / Enterprise targeting

```bash
gh auth login --hostname ghe.example.com      # interactive
export GH_HOST=ghe.example.com
export GH_ENTERPRISE_TOKEN=ghp_xxx            # NOT GH_TOKEN for GHES
gh api user --hostname ghe.example.com -q '.login'   # verify against the enterprise host
```

## CI

- Store the token as a **masked secret** and expose it as `GH_TOKEN` (or `GH_ENTERPRISE_TOKEN` for GHES); never commit it.
- **GitHub Actions:** the built-in `GITHUB_TOKEN` needs an explicit `permissions:` block — e.g. `pull-requests: write` to post reviews, `contents: write` to push. Events it creates do NOT trigger downstream workflows; to chain CI use a GitHub App installation token or a PAT secret.

```yaml
permissions:
  contents: write # push commits / refs
  pull-requests: write # open PR, post review
env:
  GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
