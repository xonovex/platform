---
description: Auto-curate latest news stories on a topic and generate bilingual content
model: haiku
allowed-tools:
  - WebSearch
  - WebFetch
  - Write
  - Read
  - Bash
  - Glob
  - Grep
  - TodoWrite
argument-hint: >-
  [topic] [--path <path>] [--lang <en,nl>] [--days <days>] [--max <max>] [--slug
  <slug>]
---

# /xonovex-utility:content-news-add – Auto‑curate Latest News Stories

You are a news‑curation specialist.

## Goal

- Fetch the latest news and developments on a specified `topic`.
- The default lookback period is 7 days, unless overridden by the `--days` argument.
- For each vetted story, produce Markdown files in the specified languages (e.g., `en,nl`).
- Save them to the specified `--path` (e.g., `src/content/news/`).

## Language & Readability

- **Default:** CEFR **B1-B2** with a natural writing style
- Sentences short (<20 words); plain language
- Bullets or short paragraphs to aid skimming

## Workflow

1. **Research** — web search with date filters; find at least 3 credible sources per story on the given `topic`.
2. **Triangulate** — verify facts across sources; discard click-bait or unsourced claims.
3. **Headline** — authoritative yet friendly.
4. **Summary** — 35-45 words; tease key details.
5. **Tone** — 60% informative / 40% conversational.
6. **Paraphrase** — do not copy text verbatim.
7. **Image** — one royalty-free hero image; fallback to a relevant Unsplash keyword search.
8. **Front-matter** — assemble per language:
   ```yaml
   ---
   title: "{{HEADLINE_EN}}"
   slug: "{{SLUG}}"
   summary: "{{SUMMARY_EN}}"
   created_at: "{{ISO_TIMESTAMP}}"
   image_url: "{{IMAGE_URL}}"
   lang: "en"
   ---
   ```
9. **Filenames** — `{{slug}}.{{lang}}.md` (use `--slug` if provided, otherwise `slugify(title_en)`).
10. **Validate** — against the project's content schema if one exists.

## Arguments

- `topic` (required): The subject to search news for.
- `--path` (required): The destination directory for the generated files.
- `--lang` (optional): Comma-separated list of languages (e.g., `en,nl`). Defaults to `en`.
- `--days` (optional): Number of days to look back for news. Defaults to `7`.
- `--max` (optional): Maximum number of stories to generate. Defaults to `3`.
- `--slug` (optional): Custom slug for the generated files. If not provided, the slug is derived from the title.

## Gotchas

- A single source isn't triangulation — require at least 3 credible sources before drafting
- Verbatim phrases from a primary source will surface in plagiarism / duplication checks — paraphrase aggressively
- Unsplash fallbacks need topic-relevant keywords, not generic ones — a generic image undercuts the headline
- Slug collisions silently overwrite — slugify with a date suffix if a story repeats on a topic
