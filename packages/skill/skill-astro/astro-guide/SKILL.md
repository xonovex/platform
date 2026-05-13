---
name: astro-guide
description: "Use when editing or scaffolding Astro sites with islands architecture. Triggers on `.astro` files, `astro:content` imports, `Astro.props`, `client:` hydration directives, content-collection schemas, and on prompts about pages, layouts, MDX, content collections, image optimization, view transitions, or accessibility in an Astro project — even when the user doesn't say 'Astro'. Skip Next.js / Remix work and unrelated static-site generators."
---

# Astro Coding Guidelines

## Essentials

- **Islands architecture** - Default to static HTML, hydrate only where needed, see [references/islands-architecture.md](references/islands-architecture.md)
- **Project structure** - Use `src/pages`, `src/components`, `src/layouts`, `src/content`, see [references/project-structure.md](references/project-structure.md)
- **Content collections** - Use `astro:content` with schema-validated frontmatter, see [references/content-collections.md](references/content-collections.md)
- **Framework components** - Integrate React, Vue, or other frameworks, see [references/components.md](references/components.md)
- **Accessibility** - Use semantic HTML, alt text, ARIA as needed, see [references/accessibility.md](references/accessibility.md)

## Gotchas

- Default rendering is server-side / static — components don't ship JavaScript unless explicitly hydrated with `client:*` directives
- Content Collections enforce a Zod schema at build time; an invalid frontmatter field fails the build, not the page
- `Astro.glob()` is build-time and scans at compile; runtime data needs `getStaticPaths` or endpoints
- Framework components (React/Vue/Svelte) only hydrate on the directive you pick — `client:load`, `client:idle`, `client:visible`, `client:media`, `client:only`

## Progressive disclosure

- Read [references/islands-architecture.md](references/islands-architecture.md) - When deciding which components need client-side JavaScript
- Read [references/project-structure.md](references/project-structure.md) - When organizing files and directories
- Read [references/content-collections.md](references/content-collections.md) - When managing blog posts, docs, or structured content
- Read [references/components.md](references/components.md) - When integrating React, Vue, or other framework components
- Read [references/accessibility.md](references/accessibility.md) - When adding keyboard navigation or screen reader support
