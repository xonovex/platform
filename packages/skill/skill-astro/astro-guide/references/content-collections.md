# content-collections: Content Collections with Type Safety

Define collections in `src/content/config.ts` with a Zod `schema`; frontmatter is validated at build time and typed for queries. Query with `getCollection()` / `getEntry()`; compile a body with `entry.render()`.

```typescript
// src/content/config.ts
import {defineCollection, z} from "astro:content";

export const collections = {
  posts: defineCollection({
    type: "content", // or "data" for JSON/YAML entries
    schema: z.object({
      title: z.string(),
      published: z.date(),
      tags: z.array(z.string()),
      draft: z.boolean().default(false),
    }),
  }),
};
```

- Store entries under `src/content/[collection]/`.
- Filter in the `getCollection("posts", ({data}) => !data.draft)` callback.
