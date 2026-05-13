# markdown-mirrors: Per-Page `.md` Versions Linked from /llms.txt

Each page that `/llms.txt` references must have a clean markdown version served at a URL derived from the original by appending `.md`.

## URL Convention

| Original URL                       | Markdown mirror URL                         |
| ---------------------------------- | ------------------------------------------- |
| `https://site.com/docs/intro.html` | `https://site.com/docs/intro.html.md`       |
| `https://site.com/docs/intro`      | `https://site.com/docs/intro.md`            |
| `https://site.com/docs/intro/`     | `https://site.com/docs/intro/index.html.md` |
| `https://site.com/` (root)         | `https://site.com/index.html.md`            |

Rule: append `.md` to the original URL. If the URL has no filename (ends in `/`), append `index.html.md` instead.

## Content Rules

The mirror should be a **clean** markdown rendering of the original page:

- Strip site chrome: nav menus, sidebars, footer links, "share" widgets
- Strip ads, popups, cookie banners, analytics markup
- Keep: the article body, code blocks (with language tags), tables, diagrams (link to image), inline links
- Preserve the page's semantic structure (headings, lists, emphasis)
- Resolve relative links to absolute URLs so the mirror is self-contained when ingested

## Generation Approaches

- **Static-site generators** — most modern SSGs (Astro, Hugo, Next.js, Eleventy, Docusaurus) can emit a `.md` alongside each `.html` route via a custom build step or plugin
- **Server middleware** — intercept requests for `*.md` URLs and render the page's source content as markdown on the fly
- **Manual export** — for small sites: export from CMS / docs source, place files at the expected paths
- **Reverse from HTML** — if no source is available: extract main content (e.g. via `readability`-style extraction) and serialize to markdown

Whichever approach: the result must be at the exact URL the `/llms.txt` link expects, returning `text/plain; charset=utf-8` or `text/markdown` with HTTP 200.

## Verification

- Every URL listed in `/llms.txt` must `200 OK`
- Content-Type should be `text/markdown` or `text/plain`
- A short sanity check: fetch the mirror, confirm it's clean markdown (no `<nav>`, no `<header>` chrome, no JavaScript)

## Gotchas

- Mirrors that include nav chrome defeat the whole purpose — they re-introduce the bloat `llms.txt` was meant to bypass
- Relative links inside a mirror are useless once the file is ingested in isolation — resolve to absolute URLs
- A mirror that 404s from a `/llms.txt` listing makes the whole file look stale — verify all links during CI
- Serving `text/html` for a `.md` URL works in browsers but breaks tools that rely on `Content-Type` — set the right MIME type
- Renaming a page changes the mirror URL; remember to update `/llms.txt` whenever URLs shift
- Image-heavy pages don't translate well — link to the page's images by absolute URL rather than embedding base64
