---
name: astro-guidelines
description: Trigger on `.astro` files. Use when building Astro static sites with islands architecture. Apply for content collections, component hydration, or accessibility implementation. Keywords: Astro, islands, static HTML, content collections, `astro:content`, schema validation.
---

# Astro Coding Guidelines

## Essentials

- **Islands architecture** - Default to static HTML, hydrate only where needed, see [reference/islands-architecture.md](reference/islands-architecture.md)
- **Project structure** - Use `src/pages`, `src/components`, `src/layouts`, `src/content`, see [reference/project-structure.md](reference/project-structure.md)
- **Content collections** - Use `astro:content` with schema-validated frontmatter, see [reference/content-collections.md](reference/content-collections.md)
- **Framework components** - Integrate React, Vue, or other frameworks, see [reference/components.md](reference/components.md)
- **Accessibility** - Use semantic HTML, alt text, ARIA as needed, see [reference/accessibility.md](reference/accessibility.md)

## Progressive disclosure

- Read [reference/islands-architecture.md](reference/islands-architecture.md) - When deciding which components need client-side JavaScript
- Read [reference/project-structure.md](reference/project-structure.md) - When organizing files and directories
- Read [reference/content-collections.md](reference/content-collections.md) - When managing blog posts, docs, or structured content
- Read [reference/components.md](reference/components.md) - When integrating React, Vue, or other framework components
- Read [reference/accessibility.md](reference/accessibility.md) - When adding keyboard navigation or screen reader support
