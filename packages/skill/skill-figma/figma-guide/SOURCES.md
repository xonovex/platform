# Sources

## Figma REST API

- **URL:** https://www.figma.com/developers/api
- **Last reviewed:** 2026-06-24
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas, Example
  - All files under `references/`
- **Aspects extracted:**
  - Personal-access-token creation flow, `GET /v1/me` verification → `references/first-time-setup.md`
  - Authentication header `X-Figma-Token`, personal-access-token scopes → `references/auth.md`
  - File-key / `node-id` URL structure and the dash→colon id form → `references/url-parsing.md`
  - `GET /v1/files/:key`, `GET /v1/files/:key/nodes`, `depth` param → `references/read-nodes.md`
  - `GET /v1/images/:key` formats, scale, signed-URL expiry → `references/render-images.md`
  - `GET /v1/files/:key/styles`, `/components`, `/variables/local`, `/comments` → `references/variables-styles.md`

## Refresh Workflow

1. Re-fetch the upstream source(s)
2. Diff against the prior pull (or scan for newly added endpoints/params)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
