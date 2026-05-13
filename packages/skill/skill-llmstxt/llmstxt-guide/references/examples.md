# examples: Reference /llms.txt Implementations

Concrete `/llms.txt` examples to model authoring decisions on.

## Minimal Example (from the spec)

```markdown
# llms.txt

> A proposal that those interested in providing LLM-friendly content add a
> /llms.txt file to their site. This is a markdown file that provides brief
> background information and guidance, along with links to markdown files
> providing more detailed information.

## Docs

- [llms.txt proposal](https://llmstxt.org/index.md): The proposal for llms.txt
- [Python library docs](https://llmstxt.org/intro.html.md): Docs for `llms-txt` python lib
- [ed demo](https://llmstxt.org/ed-commonmark.md): Tongue-in-cheek example of how llms.txt could be used in the classic `ed` editor, used to show how editors could incorporate llms.txt in general.
```

Notes:

- Single H1, single blockquote, single H2 section — minimum viable structure
- Link descriptions vary in length; longer is fine when it actually adds value (the `ed` demo line clarifies the purpose)
- Each URL ends in `.md` — no HTML URLs

## Docs-Site Example

```markdown
# Acme

> Acme is a TypeScript HTTP framework focused on edge runtimes. Use this index
> when you need API surface, recipes, or middleware patterns.

## Docs

- [Getting started](https://acme.dev/docs/getting-started.md): Install + first server
- [API reference](https://acme.dev/docs/api.md): All exported types and functions
- [Middleware](https://acme.dev/docs/middleware.md): Built-in middleware list
- [Routing](https://acme.dev/docs/routing.md): Route definition + path matching

## Recipes

- [Auth with JWT](https://acme.dev/recipes/jwt-auth.md): JWT validation middleware end-to-end
- [Rate limiting](https://acme.dev/recipes/rate-limit.md): Token-bucket pattern
- [File uploads](https://acme.dev/recipes/upload.md): Streaming multipart parsing

## Optional

- [Changelog](https://acme.dev/changelog.md): Release notes by version
- [Migration v1 → v2](https://acme.dev/migration-v1-v2.md): Breaking changes between major versions
- [Internal design notes](https://acme.dev/design.md): Rationale behind framework decisions
```

Pattern this demonstrates:

- Two themed sections (`## Docs`, `## Recipes`) for primary content
- `## Optional` for context-when-needed material (changelog, migration, design)
- Descriptions lead with the outcome ("JWT validation middleware end-to-end", not "JWT documentation")

## Sibling Convention: `llms-full.txt`

Some projects publish a single-file dump of all their docs as `/llms-full.txt` (e.g. `https://hono.dev/llms-full.txt`). This is **not the spec** — the spec defines `llms.txt` (the curated index) and `llms-ctx*.txt` (processor outputs). The `llms-full.txt` convention is a community pattern that works well for small-to-medium docs sites where the whole corpus fits in a single file.

When a project provides both:

- `/llms.txt` — curated index with `.md` links (for tooling that follows the spec)
- `/llms-full.txt` — single-file dump (for one-shot ingestion)

## Real-World Examples to Study

- **llmstxt.org itself** — minimal example, just the spec links
- **FastHTML docs** — referenced in the spec as a worked example
- **Hono** — `https://hono.dev/llms-full.txt` (full-dump sibling convention)
- **Anthropic, Stripe, Vercel** docs sites — many large doc sets now publish `llms.txt` files

## Gotchas

- Copying a competitor's structure verbatim without revisiting groupings produces awkward fits — group by what _your_ users would ask, not what someone else grouped
- A bloated `## Docs` section with 50 links is unhelpful — split into themed sub-sections (`## API`, `## Guides`, `## Reference`) when one section grows past ~15 entries
- "Examples" sections that just list page titles without describing the example's outcome waste space — say what each example demonstrates
- Mixing HTML and `.md` URLs inside the same file is a fast way to make it look unmaintained — pick one (`.md` is the spec) and stick with it
