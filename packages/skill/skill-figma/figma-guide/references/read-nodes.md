# read-nodes: Read file and node JSON

**Guideline:** Fetch specific subtrees with `GET /v1/files/:key/nodes?ids=` and bound any whole-file read with `depth`. Reserve the full `GET /v1/files/:key` for when you genuinely need the entire document.

**Rationale:** A Figma document is a deep tree (canvas → frames → groups → layers). Fetching the whole thing is megabytes and slow, and rate limits punish repeated large pulls. The `nodes` endpoint returns just the requested subtrees, and `depth` truncates how many child levels come back, so you fetch only what you traverse.

**How to Apply:**

1. **Targeted** - `GET /v1/files/:key/nodes?ids=<id1>,<id2>`; each id maps to `nodes.<id>.document`, the subtree root.
2. **Bounded whole-file** - `GET /v1/files/:key?depth=1` (pages/canvases only) or `depth=2` (top-level frames); omit `depth` only when you must walk everything.
3. **Geometry** - add `&geometry=paths` for vector path data; skip it otherwise to keep payloads small.
4. **Traverse** - recurse `document.children`; key fields are `name`, `type` (`FRAME`/`INSTANCE`/`TEXT`/`COMPONENT`/...), `absoluteBoundingBox`, `fills`, `characters` (text).
5. Parse with a real JSON tool (`python3 -c` / `jq`), never regex.

**Example:**

```bash
TOKEN=$(security find-generic-password -s figma-token -w)
curl -s -H "X-Figma-Token: $TOKEN" \
  "https://api.figma.com/v1/files/<file-key>/nodes?ids=1234:56" \
| python3 -c '
import sys, json
doc = next(iter(json.load(sys.stdin)["nodes"].values()))["document"]
def walk(n, d=0):
    print("  "*d + f'\''[{n["type"]}] {n["name"]}'\'')
    for c in n.get("children", []): walk(c, d+1)
walk(doc)'
```

**Counter-Example:** Do not loop one `nodes?ids=<single>` request per layer - comma-join the ids into one call; many small calls hit the 429 rate limit fast.

**Related:** [url-parsing.md](./url-parsing.md), [render-images.md](./render-images.md), [variables-styles.md](./variables-styles.md)
