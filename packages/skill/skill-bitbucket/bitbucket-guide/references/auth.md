# Bitbucket Data Center REST Authentication

Load **credential-management-guide** for provider-neutral storage, automation, rotation, and exposure response. This reference owns Bitbucket Data Center token types, version-specific transport, host resolution, and verification.

## Product and host boundary

These recipes target self-hosted Bitbucket Server or Data Center REST 1.0 at `/rest/api/1.0`. Bitbucket Cloud REST 2.0 has a different authentication model. Git over SSH and REST over HTTPS are independent.

Resolve an SSH alias before constructing the HTTPS REST URL:

```bash
ssh -G <ssh-alias> | grep -iE '^(hostname|user|port) '
```

There is no first-party Data Center CLI for this workflow; call REST with `curl`.

## Token and scope

Create a user token under Manage account → HTTP access tokens or a repository-scoped token under repository settings. Prefer the repository scope, a short expiry, and Pull requests: write (`REPO_WRITE`) for comment mutations. A read-only token can pass a GET probe and still return `403` on POST.

## Transport

- User tokens support Bearer and Basic authentication.
- Project and repository tokens are Bearer-only before Bitbucket Data Center 9.4. Data Center 9.4 and later also support Basic authentication for them.
- When version or scope is uncertain, use `Authorization: Bearer`.

```bash
TOKEN="$(<provider-neutral-secret-store-read>)"
BASE=https://<host>/rest/api/1.0/projects/<KEY>/repos/<repo>
curl -s -H "Authorization: Bearer $TOKEN" \
  -o /dev/null -w '%{http_code}\n' "$BASE/pull-requests?limit=1"
```

Provider recipes that use `curl -n` require Basic auth in a `0600` netrc and therefore are valid only for a user token or a repository token on Data Center 9.4+. A netrc is a materialized secret, not the source of truth; create it for the session from the store selected by **credential-management-guide** and remove it afterward.

`200` verifies read access. `401` indicates bad transport, token, or username; `403` on a write indicates missing `REPO_WRITE`; `404` indicates wrong project or repository coordinates.
