# Figma REST API Authentication

Load **credential-management-guide** for provider-neutral storage, CI injection, rotation, and exposure response. This reference owns Figma token scopes, request headers, plan constraints, and verification.

## Token and scope

Create a personal access token under Figma → Settings → Security → Personal access tokens. Most read flows need `file_content:read` or the legacy `file_read`; local variables need `file_variables:read`. Avoid write scopes for design-to-code reads. The variables scopes and local-variables endpoint require an Enterprise plan.

## Transport

A personal access token uses `X-Figma-Token`; an OAuth access token uses `Authorization: Bearer`. They are not interchangeable.

```bash
TOKEN="$(<provider-neutral-secret-store-read>)"
curl -s -H "X-Figma-Token: $TOKEN" \
  "https://api.figma.com/v1/files/<file-key>"
```

Replace the placeholder with a retrieval command selected by **credential-management-guide**. For a static-value-only process configuration, use a wrapper that retrieves and exports the value rather than pasting it into configuration.

## Verification

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "X-Figma-Token: $TOKEN" https://api.figma.com/v1/me
```

`200` confirms the PAT. `401` usually indicates a missing token or wrong header scheme; `403` may indicate scope or plan limitations; `404` on a file endpoint indicates the wrong key or inaccessible file.

## Automation boundary

A Figma PAT cannot be replaced with a cloud-provider OIDC token. Store the PAT as a protected CI secret when automation needs it and expose it only to the read step. OIDC remains preferable for separate federatable cloud credentials in the same pipeline.
