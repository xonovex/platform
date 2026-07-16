# content-collections: Content Collections with Type Safety

Define collections in `src/content/config.ts`; frontmatter is validated at build time and typed for queries. Query with `getCollection()` / `getEntry()`; compile a body with `entry.render()`.

- `type: "content"` for Markdown/MDX bodies, `type: "data"` for JSON/YAML entries.
- Store entries under `src/content/[collection]/`.
- Filter inside the query, not after it: `getCollection("posts", ({ data }) => !data.draft)`.
