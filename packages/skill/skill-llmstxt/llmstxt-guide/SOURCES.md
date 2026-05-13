# Sources

## llms.txt Specification

- **URL:** https://llmstxt.org/
- **URL:** https://llmstxt.org/index.md
- **Last reviewed:** 2026-05-13
- **Used for:**
  - `SKILL.md` → all sections
  - `references/create.md` → Workflow, Template, Description-Writing Rules, Gotchas
  - `references/markdown-mirrors.md` → URL Convention, Content Rules
  - `references/processing-tools.md` → Output Variants, distinction between `llms-ctx*.txt` and `llms-full.txt`
- **Aspects extracted:**
  - Motivation: LLM context-window limits → SKILL.md "What llms.txt Is"
  - Strict file format ordering (H1 required; blockquote optional; descriptive content optional with no headings; H2 link lists optional) → SKILL.md "File Format"
  - Link list grammar (`[name](url)` required, optional `:` notes) → SKILL.md + `create.md` Template
  - Special `## Optional` section semantics for short-context processor output → SKILL.md + `processing-tools.md`
  - `.md` suffix convention for per-page markdown mirrors, including `index.html.md` for filename-less URLs → SKILL.md + `markdown-mirrors.md` URL Convention table
  - Distinction from robots.txt (crawler permissions) and sitemap.xml (exhaustive HTML listing) → SKILL.md "What llms.txt Is"
  - `llms-ctx.txt` vs `llms-ctx-full.txt` processor output variants → SKILL.md + `processing-tools.md`
  - Authoring best practices: concise clear language, brief informative descriptions, avoid ambiguous jargon, test with multiple LLMs → SKILL.md Core Principles + `create.md` Workflow

## llms.txt Example File

- **URL:** https://llmstxt.org/llms.txt
- **Last reviewed:** 2026-05-13
- **Used for:**
  - `references/examples.md` → "Minimal Example (from the spec)"
- **Aspects extracted:**
  - Verbatim reproduction of the official minimal example demonstrating H1 + blockquote + single H2 section + 3 links

## Refresh Workflow

1. Re-fetch `https://llmstxt.org/` and `https://llmstxt.org/index.md`
2. Diff against the prior pull — look for new sections, changed grammar, new processor tool conventions
3. Update SKILL.md and the relevant references
4. Re-fetch `https://llmstxt.org/llms.txt` to keep the canonical example in `examples.md` current
5. Bump **Last reviewed** dates above

## Related (not direct sources, mentioned in this skill)

- `https://hono.dev/llms-full.txt` — example of the informal `llms-full.txt` community convention (single-file dump), not the spec. Mentioned in `examples.md` and `processing-tools.md`.
