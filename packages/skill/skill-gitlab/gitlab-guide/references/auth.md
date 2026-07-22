# GitLab Authentication and Token Scopes

Load **credential-management-guide** for provider-neutral storage, CI boundaries, rotation, and exposure response. This reference owns GitLab token types, exact scopes and roles, host variables, and verification.

## Token types

- A personal access token is account-wide and cannot exceed its owner's access.
- A project or group access token runs as a bot user and contains automation to that resource; prefer it over a personal token.
- Fine-grained PATs provide resource and permission pairs but remain version-dependent; verify support before relying on them.
- `CI_JOB_TOKEN` can read merge requests but cannot create them, post notes, or approve them.
- OAuth tokens serve interactive `glab auth login --web` or `--device` flows.

Use the shortest supported expiry.

## Scopes and roles

| Operation                             | Classic scope                           | Additional authority                |
| ------------------------------------- | --------------------------------------- | ----------------------------------- |
| Read merge requests, diffs, and notes | `read_api`                              | Project read role                   |
| Create merge requests or discussions  | `api`                                   | Developer or eligible author        |
| Approve or resolve discussions        | `api`                                   | Eligible approver, Developer/author |
| Push or read repository files         | `write_repository` or `read_repository` | Corresponding repository access     |

`write_repository` does not authorize API discussions, and `read_api` cannot POST. Scope is necessary but not sufficient: an `api` token still receives `403` when its identity lacks the project role. Some unapprove and reset-approval operations are bot-only.

## Hosts and variables

`GITLAB_TOKEN` is a general API token for any instance; it is not the self-managed switch. Select a self-managed or Dedicated instance with `GITLAB_HOST`, `GL_HOST`, `GITLAB_URI`, or `--hostname`. Environment tokens override stored `glab` credentials and can silently select another identity. Track the installed `glab` version when using the newer `GLAB_` variable names.

Git and API protocols are separate per-host settings. A successful SSH clone does not verify API authentication.

## Storage and verification

`glab` stores tokens in plaintext `~/.config/glab-cli/config.yml` unless login uses `--use-keyring`. Prefer the keyring or runtime retrieval selected by **credential-management-guide**.

```bash
glab auth status
glab api user --hostname <host>
```

Run both against the intended host before a write. In CI, merge-request writes require a protected PAT or project/group token; do not rely on `CI_JOB_TOKEN` for them.
