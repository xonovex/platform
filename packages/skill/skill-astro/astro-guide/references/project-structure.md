# project-structure: Project Structure and Organization

Use Astro's conventional directories. `src/pages/` drives file-based routing; `[slug].astro` becomes a dynamic segment; `public/` files are served as-is at the root.

```
src/
├── pages/
│   ├── index.astro           # → /
│   ├── about.astro           # → /about
│   └── blog/[slug].astro     # → /blog/:slug
├── components/               # reusable .astro / framework components
├── layouts/                  # shared page shells
├── content/
│   ├── config.ts             # collection schemas
│   └── posts/
└── styles/global.css
public/                       # served verbatim at site root
```
