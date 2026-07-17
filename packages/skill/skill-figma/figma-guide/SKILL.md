---
name: figma-guide
description: "Use when reading Figma designs through the Figma REST API (api.figma.com) - resolving a figma.com/design or /file URL into a file key and node id, fetching file or node JSON, rendering frames to PNG/SVG/PDF, and extracting variables, styles, components, or comments. Triggers on a figma.com link, a node-id like 1234-56, X-Figma-Token, api.figma.com, 'export this frame', 'get the design spec', or pulling colors / typography / spacing from a design, even when the user doesn't say 'API'."
---

# Figma REST API Guidelines

Read designs from Figma over the REST API. This skill owns the API mechanics - auth, URL parsing, reading nodes, rendering images, and extracting design tokens. Implementing the extracted spec in UI code is a separate concern handled by your framework's UI skill.

## Requirements

- A Figma personal access token, stored in the OS keychain (e.g. macOS `security find-generic-password -s figma-token -w`), passed as the `X-Figma-Token` header - see [references/auth.md](references/auth.md)
- `curl` plus a JSON parser (`python3` or `jq`) - there is no Figma CLI to install; first token, keychain, and `GET /v1/me` check are in [references/first-time-setup.md](references/first-time-setup.md)
- Network access to `https://api.figma.com`

## Essentials

- **Authenticate from the keychain** - read the token at runtime, send `X-Figma-Token`, never hardcode or commit it, see [references/auth.md](references/auth.md)
- **Resolve the URL** - file key is the `/design/<key>/` or `/file/<key>/` segment, the node is `node-id=1234-56` and the API needs it dash-to-colon as `1234:56`, see [references/url-parsing.md](references/url-parsing.md)
- **Read targeted nodes** - prefer `GET /v1/files/:key/nodes?ids=` (cheap) over the whole-file `GET /v1/files/:key` (use `depth=` to bound it), see [references/read-nodes.md](references/read-nodes.md)
- **Render frames** - `GET /v1/images/:key?ids=&format=png&scale=2` returns short-lived signed URLs to download promptly, see [references/render-images.md](references/render-images.md)
- **Extract design tokens** - styles, components, and (Enterprise) local variables for colors / type / spacing, see [references/variables-styles.md](references/variables-styles.md)
- **Then implement** - map the extracted names/tokens onto your UI framework's components as a separate implementation step

## Gotchas

- The `node-id` in a share URL uses a dash (`1234-56`); the API's `ids=` param needs a colon (`1234:56`). Pass the dash form and you get an empty `nodes` map back with no error
- A personal access token grants full-account API access - treat it like a password, keep it in the keychain, and rotate it in Figma → Settings → Personal access tokens if it ever leaks
- `GET /v1/files/:key` with no `depth` returns the entire document tree - it can be megabytes and slow; default to `/nodes?ids=` or `depth=1`/`depth=2`
- `/v1/images` returns time-limited signed S3 URLs (they expire) - download immediately, never persist the URL itself
- The REST API is rate-limited (HTTP 429) - batch multiple node ids into one comma-separated `ids=` request instead of looping one call per node
- The `?t=...` query param on a share link is a share token, irrelevant to the API; only the file key and `node-id` matter
- Local variables (`/v1/files/:key/variables/local`) are Enterprise-plan + scoped-token only; the styles endpoint works on all plans
- This is the token-based REST path; an OAuth-based Figma integration authenticates via `Authorization: Bearer` instead - don't conflate the two

## Example

```bash
# token stays in the keychain; read it at call time, never inline it
TOKEN=$(security find-generic-password -s figma-token -w)

# URL: figma.com/design/<file-key>/<slug>?node-id=1234-56
#   file key -> <file-key>   node-id 1234-56 -> ids=1234:56
curl -s -H "X-Figma-Token: $TOKEN" \
  "https://api.figma.com/v1/files/<file-key>/nodes?ids=1234:56" \
| python3 -c 'import sys,json; d=json.load(sys.stdin); \
n=next(iter(d["nodes"].values()))["document"]; \
print(n["name"], n["type"], [c["name"] for c in n.get("children",[])])'
```

## Progressive Disclosure

- Read [references/first-time-setup.md](references/first-time-setup.md) - Load when starting from zero: no CLI exists, creating a first personal access token, storing it in the keychain, and verifying with `GET /v1/me`
- Read [references/auth.md](references/auth.md) - Load when retrieving the token, sending the `X-Figma-Token` header, choosing scopes, storing it locally or in CI/CD, or rotating a leaked token
- Read [references/url-parsing.md](references/url-parsing.md) - Load when turning a figma.com link into a file key and API-ready node id
- Read [references/read-nodes.md](references/read-nodes.md) - Load when fetching file or node JSON and traversing the document tree
- Read [references/render-images.md](references/render-images.md) - Load when exporting frames or layers to PNG / SVG / PDF
- Read [references/variables-styles.md](references/variables-styles.md) - Load when extracting variables, styles, components, or comments as a design spec
