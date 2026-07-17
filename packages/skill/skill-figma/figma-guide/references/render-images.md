# render-images: Export frames and layers to images

**Guideline:** Render nodes with `GET /v1/images/:key?ids=&format=&scale=`; the response gives signed URLs you must download promptly because they expire.

**Rationale:** The images endpoint rasterizes/vectorizes nodes server-side and returns time-limited signed S3 URLs rather than the bytes themselves. The URLs expire, so they are for immediate download, not storage. Format and scale control fidelity and size.

**How to Apply:**

1. Request: `GET /v1/images/:key?ids=<id1>,<id2>&format=png&scale=2`.
   - `format`: `png` (default), `jpg`, `svg`, `pdf`.
   - `scale`: `0.01`–`4` for raster (`2` ≈ @2x); ignored for vector formats.
   - SVG extras: `svg_outline_text=false`, `svg_include_id=true` when you need editable/identified SVG.
2. The response is `{"images": {"<id>": "<signed-url-or-null>"}}`; a `null` value means that node failed to render.
3. Download each URL right away (`curl -o`), before it expires.
4. Render many nodes in one call by comma-joining ids.

**Example:**

```bash
TOKEN=$(security find-generic-password -s figma-token -w)
URL=$(curl -s -H "X-Figma-Token: $TOKEN" \
  "https://api.figma.com/v1/images/<file-key>?ids=1234:56&format=png&scale=2" \
  | python3 -c 'import sys,json; print(next(iter(json.load(sys.stdin)["images"].values())))')
curl -s "$URL" -o /tmp/frame.png   # download immediately - the signed URL expires
```

**Counter-Example:** Do not store or share the returned S3 URL as a permanent image link - it expires; persist the downloaded file instead.

**Related:** [url-parsing.md](./url-parsing.md), [read-nodes.md](./read-nodes.md)
