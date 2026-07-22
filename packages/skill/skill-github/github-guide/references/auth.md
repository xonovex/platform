# GitHub Authentication and Token Permissions

Load **credential-management-guide** for provider-neutral storage, CI boundaries, rotation, and exposure response. This reference owns GitHub token families, exact permissions, hosts, variables, and verification.

## Token families

Classic PATs use coarse scopes. `repo` covers private-repository push, pull-request, review, comment, and thread-resolution work; `public_repo` is the public-only alternative. Classic PATs remain necessary for some public non-member, outside-collaborator, and multi-organization cases.

Prefer a fine-grained PAT where its repository and organization constraints fit:

| Operation                             | Fine-grained permission                         |
| ------------------------------------- | ----------------------------------------------- |
| Push commits or update refs           | Contents: write                                 |
| Open a pull request                   | Pull requests: write and Contents: read         |
| Submit a review                       | Pull requests: write                            |
| Post a top-level pull-request comment | Issues: write                                   |
| Resolve a review thread               | Pull requests: write and Contents: read & write |

Metadata: read is automatically required. Organization approval may leave a fine-grained token pending and effectively read-public-only.

## Verify endpoint requirements

```bash
gh api -i repos/{owner}/{repo}/pulls/123/reviews 2>&1 \
  | grep -i x-accepted-github-permissions
```

In `X-Accepted-GitHub-Permissions`, commas mean all listed permissions and semicolons indicate alternatives. Insufficient access to a private repository may return `404` instead of `403`.

## Hosts, variables, and storage

- github.com and `*.ghe.com`: `GH_TOKEN` or `GITHUB_TOKEN`, with `GH_TOKEN` taking precedence.
- GitHub Enterprise Server: `GH_ENTERPRISE_TOKEN` or `GITHUB_ENTERPRISE_TOKEN`.
- `GH_HOST` or `--hostname` selects the target; a working github.com token proves nothing about GHES.

`gh auth login` stores credentials in the OS keyring by default and can fall back to plaintext configuration when no keyring exists. Avoid `--insecure-storage` on shared machines. **credential-management-guide** owns external secret-store retrieval and lifecycle behavior.

Verify the intended host and identity before a write:

```bash
gh auth status
gh api user --hostname <host> -q '.login'
```

## GitHub Actions

The built-in `GITHUB_TOKEN` needs an explicit `permissions:` block. Use `pull-requests: write` for pull-request mutations and `contents: write` for pushes. Events produced with this token generally do not trigger a new workflow run; use a reviewed GitHub App installation token or PAT only when chaining is required.
