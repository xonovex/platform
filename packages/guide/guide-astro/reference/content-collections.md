# content-collections: Content Collections with Type Safety

**Guideline:** Use Astro content collections API with Zod schemas for type-safe content management.

**Rationale:** Content collections provide type safety, schema validation, IntelliSense, standardized organization, and build-time frontmatter validation.

**Example:**

```typescript
// src/content/config.ts
import {defineCollection, z} from "astro:content";

export const collections = {
  posts: defineCollection({
    type: "content",
    schema: z.object({
      title: z.string(),
      published: z.date(),
      author: z.string(),
      tags: z.array(z.string()),
      draft: z.boolean().default(false),
    }),
  }),
};
```

**Techniques:**

- Collection schemas: Define Zod schemas in `src/content/config.ts` for validation
- Directory structure: Organize content in `src/content/[collection-name]/` directories
- Query functions: Use `getCollection()` and `getEntry()` for type-safe queries
- Filtering: Apply runtime filters in `getCollection()` callback
- Rendering: Use `.render()` method to compile markdown to HTML components
