# variables-styles: Extract design tokens, styles, components, and comments

**Guideline:** Pull the design spec from the dedicated endpoints - styles and components on any plan, local variables on Enterprise - rather than scraping values out of raw node fills.

**Rationale:** Figma exposes named design tokens through purpose-built endpoints. Styles and published components carry the semantic names (e.g. "Primary/Blue", "Heading/L") that map onto your design tokens; variables are the modern multi-mode token system. Reading these is more stable than inferring hex codes from individual layer `fills`.

**How to Apply:**

1. **Styles** - `GET /v1/files/:key/styles` lists fill/text/effect/grid styles with `name`, `style_type`, `key`. Resolve the actual values by fetching the styled nodes (see [read-nodes.md](./read-nodes.md)).
2. **Components** - `GET /v1/files/:key/components` (and `/component_sets`) lists published components and variants with names and `node_id`.
3. **Variables** - `GET /v1/files/:key/variables/local` returns variable collections, modes, and resolved values. Enterprise-plan + `file_variables:read` scope only; expect 403 otherwise.
4. **Comments** - `GET /v1/files/:key/comments` returns review threads anchored to nodes - useful for picking up design-handoff notes.
5. Map names → your design tokens, then implement in your UI framework as a separate step.

**Example:**

```bash
TOKEN=$(security find-generic-password -s figma-token -w)
curl -s -H "X-Figma-Token: $TOKEN" \
  "https://api.figma.com/v1/files/<file-key>/styles" \
| python3 -c 'import sys,json; [print(s["style_type"], "-", s["name"]) for s in json.load(sys.stdin)["meta"]["styles"]]'
```

**Counter-Example:** Do not hardcode a hex value read off one layer's `fills` as "the brand color" - look it up via the named style/variable so renames and mode changes stay traceable.

**Related:** [read-nodes.md](./read-nodes.md), [auth.md](./auth.md)
