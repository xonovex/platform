# project-structure: Project Structure and Organization

**Guideline:** Organize using conventional directories: pages, components, layouts, content, public.

**Rationale:** Opinionated defaults enable file-based routing, clear separation of concerns, better maintainability, easier navigation.

**Example:**

```
src/
├── pages/
│   ├── index.astro           # → /
│   ├── about.astro           # → /about
│   └── blog/[slug].astro     # → /blog/:slug
├── components/
│   ├── Header.astro
│   ├── Card.tsx
├── layouts/
│   ├── BaseLayout.astro
│   └── BlogLayout.astro
├── content/
│   ├── config.ts
│   └── posts/
└── styles/global.css
```

**Techniques:**

- Pages directory: Place routes in `src/pages/` for automatic file-based routing
- Components folder: Store reusable components in `src/components/`
- Layouts directory: Create reusable layouts in `src/layouts/`
- Content collections: Organize structured content in `src/content/[collection]/`
- Static assets: Place static files in `public/` directory for direct access
