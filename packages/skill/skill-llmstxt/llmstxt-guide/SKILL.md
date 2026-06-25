---
name: llmstxt-guide
description: "Use when authoring, reviewing, or maintaining an `/llms.txt` file or per-page markdown mirrors per the llmstxt.org specification. Triggers on edits to `llms.txt` / `llms-full.txt`, on prompts about LLM-friendly site content, providing markdown versions of pages (`.md` suffix), generating expanded context files (`llms-ctx.txt`), or making documentation accessible to LLMs — even when the user doesn't say 'llms.txt'."
---

# llms.txt Guidelines

Author and maintain a project's `/llms.txt` file (and the per-page markdown mirrors it points to), so LLMs can ingest the site at inference time within a bounded context window.

## What llms.txt Is

A markdown file served at `/llms.txt` (or a subpath) that gives an LLM a curated, concise map of a site's most useful content. Complements — does not replace — robots.txt (which is about crawler permissions) and sitemap.xml (which is exhaustive and HTML-oriented).

## File Format (strict ordering)

The file must contain these elements in this order:

1. **H1 heading** _(required)_ — project or site name
2. **Blockquote** _(optional)_ — short summary with key understanding info
3. **Descriptive content** _(optional)_ — zero or more markdown sections, **excluding headings**
4. **H2-delimited file lists** _(optional, repeatable)_ — curated link sections

Only the H1 is required. The other sections are optional but follow this ordering when present.

### Link list grammar (inside an H2 section)

```markdown
## Section name

- [Link title](https://example.com/page.md): Optional notes
- [Another link](https://example.com/other.md)
```

Each list item is:

- A markdown hyperlink `[name](url)` _(required)_
- Optionally followed by `:` and brief notes

### The "Optional" section

An H2 section literally titled `## Optional` designates secondary resources that can be skipped when a shorter context is needed. Processors use this distinction to produce two variants of expanded context.

## Markdown Mirror Convention

For every HTML page worth giving to LLMs, serve a clean markdown version at the **same URL with `.md` appended**:

- `https://example.com/docs/intro.html` → `https://example.com/docs/intro.html.md`
- `https://example.com/docs/intro` → `https://example.com/docs/intro.md`
- `https://example.com/docs/intro/` → `https://example.com/docs/intro/index.html.md`
- `https://example.com/` → `https://example.com/index.html.md`

The `llms.txt` file links to these `.md` mirrors, not the HTML originals.

## Expanded Context Variants

When a site provides `llms.txt` + linked `.md` mirrors, a processing tool (e.g. `llms_txt2ctx`) can produce expanded single-file context bundles:

- **`llms-ctx.txt`** — expansion **without** the `## Optional` URLs (shorter context)
- **`llms-ctx-full.txt`** — expansion **including** the `## Optional` URLs (full context)

Some projects also publish `llms-full.txt` as a single-file dump of everything (e.g. `https://hono.dev/llms-full.txt`) — this is a common convention even though the spec primarily defines `llms.txt`.

## Authoring Workflow

1. **Inventory** the site's content; identify pages worth ingesting (API docs, reference, tutorials, recipes)
2. **Mirror to `.md`** each selected page (clean markdown, no nav chrome)
3. **Group** related links into H2 sections (e.g. `## Docs`, `## API`, `## Examples`)
4. **Mark secondary content** under `## Optional`
5. **Draft `/llms.txt`** following the strict ordering: H1 → blockquote → optional prose → H2 lists
6. **Test** by expanding to a context file and running it through ≥2 LLMs; refine link descriptions if responses are vague
7. **Serve** at `/llms.txt` (root preferred); commit alongside docs so it stays current

## Core Principles

- **Concise, clear language** — every link description earns its tokens
- **Brief informative notes** — say what the link gives, not "click here for more"
- **Avoid ambiguous terms / unexplained jargon** — LLMs can't ask follow-up questions
- **Mirror at parity** — every linked `.md` must actually exist at the URL
- **Curate, don't dump** — `llms.txt` is a guide, not a sitemap; aggregate-too-large content is the whole point of the format
- **Test with multiple LLMs** — single-model testing hides framing issues

## Gotchas

- The H1 is the **only required** element; everything else is optional but must follow strict ordering when present (blockquote → prose → H2 lists, never reordered)
- Descriptive prose between the blockquote and the first H2 must **not** contain headings — sub-section structure is reserved for the H2 file lists
- Linking the HTML URL instead of the `.md` mirror is the #1 authoring mistake — LLMs receive raw HTML with nav chrome, breaking the whole purpose
- A page without filename needs `index.html.md`, not just `.md`
- `llms-full.txt` (single dump) is a popular sibling convention but **not part of the spec** — don't confuse it with the spec's `llms-ctx-full.txt` (which is a processing output)
- `## Optional` is a literal section title with semantic meaning to processors — don't rename it ("Extra", "Bonus", "Additional") if you want the short-context variant to work

## Operations

- **Create** — author a new `/llms.txt` from scratch — see [references/create.md](references/create.md)
- **Markdown mirrors** — produce per-page `.md` versions linked from `/llms.txt` — see [references/markdown-mirrors.md](references/markdown-mirrors.md)
- **Examples** — reference implementations to model — see [references/examples.md](references/examples.md)
- **Processing tools** — generate `llms-ctx.txt` / `llms-ctx-full.txt` from `llms.txt` + mirrors — see [references/processing-tools.md](references/processing-tools.md)

## Progressive Disclosure

- Read [references/create.md](references/create.md) - Load when authoring a new `/llms.txt` from scratch for a project or site
- Read [references/markdown-mirrors.md](references/markdown-mirrors.md) - Load when producing the per-page `.md` mirrors that `/llms.txt` links to (URL conventions, content rules)
- Read [references/examples.md](references/examples.md) - Load when looking for reference `/llms.txt` implementations to model your own on
- Read [references/processing-tools.md](references/processing-tools.md) - Load when expanding `/llms.txt` + mirrors into a single-file context bundle (`llms-ctx.txt`, `llms-ctx-full.txt`)
