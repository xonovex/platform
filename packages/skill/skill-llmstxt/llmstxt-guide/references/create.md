# create: Author a /llms.txt from Scratch

Produce a spec-compliant `/llms.txt` for a project or site.

## Prerequisites

- A list of pages worth giving LLMs (API docs, reference, tutorials, recipes)
- Markdown mirror plan (see [markdown-mirrors.md](markdown-mirrors.md)) — every linked page must have a `.md` version

## Workflow

1. **Inventory** — list candidate pages; cut anything inferable from a quick title scan
2. **Group** — cluster pages into themes (e.g. Docs / API / Examples / Recipes); each becomes an H2 section
3. **Triage** — mark secondary clusters as `## Optional` so processors can produce a short context too
4. **Write descriptions** — one-line notes that say _what the page gives_, not "see X for more"
5. **Compose the file** in strict order:
   - H1 (project name)
   - Blockquote summary (1-3 sentences; what the project is, why an LLM cares)
   - Optional prose (no headings — paragraphs and lists only)
   - H2 sections with link lists
   - `## Optional` at the end if used
6. **Validate ordering** — H1 first, then blockquote, then prose, then H2 lists; nothing reordered
7. **Test** — expand to a context bundle and ask 2-3 LLMs questions a user would ask; refine link descriptions where responses are vague
8. **Serve** at `/llms.txt` (root) or a documented subpath

## Template

```markdown
# {Project Name}

> {1-3 sentence summary: what the project is and what an LLM should know going in.}

{Optional prose paragraphs — no headings allowed here. Lists are fine.}

## {Primary section, e.g. Docs}

- [Page title](https://example.com/docs/intro.html.md): What this page covers in one line
- [Page title](https://example.com/docs/api.md): One-line note
- [Page title](https://example.com/docs/guide.md)

## {Another section, e.g. Examples}

- [Example title](https://example.com/examples/x.md): What the example demonstrates

## Optional

- [Page title](https://example.com/changelog.html.md): Release history
- [Page title](https://example.com/old-design.html.md): Pre-v2 design notes
```

## Description-Writing Rules

- Lead with the **outcome** the page enables ("Set up auth with JWT", not "Auth documentation")
- Avoid vague verbs (`covers`, `discusses`, `explores`) — use concrete ones (`configures`, `lists`, `returns`)
- ≤80 chars per description; one line per list item
- Don't restate the link title
- Skip the description entirely if the title already says it ("Changelog" + nothing is fine)

## Gotchas

- Linking the HTML URL instead of the `.md` mirror defeats the format — LLMs receive raw HTML with nav chrome
- A page without filename needs `.../page/index.html.md`, not `.../page.md`
- Renaming `## Optional` to anything else breaks processor short-context output
- Descriptive prose between blockquote and first H2 must contain no headings — sub-headings break the spec
- "Too large to fit in an LLM context" is the whole motivation — if a project has only 5 pages total, `llms.txt` adds little value over a sitemap
- Single-LLM testing hides framing issues; test against ≥2 models
