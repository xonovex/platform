# url-parsing: Resolve a Figma URL into a file key and node id

**Guideline:** From a `figma.com` link, take the file key from the path and the node id from the `node-id` query param, then convert the node id's dash to a colon before sending it to the API.

**Rationale:** Share URLs and API identifiers disagree on the node-id separator. The browser shows `node-id=1234-56` (dash), but `GET /v1/files/:key/nodes?ids=` expects `1234:56` (colon). Passing the dash form returns an empty `nodes` map with HTTP 200 - a silent miss, not an error - so the conversion is mandatory.

**How to Apply:**

1. **File key** - the segment after `/design/` or `/file/`:
   `figma.com/design/<KEY>/<slug>?...` → `<KEY>`.
2. **Node id** - the `node-id` query value; replace the first `-` with `:`:
   `node-id=1234-56` → `1234:56`.
3. Ignore `?t=<share-token>`, `&m=`, and the human-readable slug - none affect the API call.
4. Request multiple nodes by comma-joining ids: `ids=1234:56,1234:78`.

**Example:**

```bash
URL='https://www.figma.com/design/<file-key>/<slug>?node-id=1234-56&t=<share-token>'

KEY=$(printf '%s' "$URL" | sed -E 's#.*/(design|file)/([^/]+)/.*#\2#')
NODE=$(printf '%s' "$URL" | sed -E 's#.*[?&]node-id=([0-9]+)-([0-9]+).*#\1:\2#')

echo "$KEY"   # <file-key>
echo "$NODE"  # 1234:56
```

**Counter-Example:** Do not send `ids=1234-56` - the dash form yields `{"nodes":{}}` with no error, which looks like "node not found" but is really a malformed id.

**Related:** [read-nodes.md](./read-nodes.md), [render-images.md](./render-images.md)
