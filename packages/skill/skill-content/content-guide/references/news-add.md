# news-add: Curate Latest News Stories

Auto-curate the latest news and developments on a given topic, then produce CEFR B1-B2 bilingual articles as Markdown with structured frontmatter.

## Goal

- Fetch the latest news on a specified `topic`
- Default lookback: 7 days (configurable)
- For each vetted story, produce one Markdown file per requested language (e.g. `en,nl`)
- Save to the specified target directory (e.g. `src/content/news/`)

## Language & Readability

- **Default:** CEFR **B1-B2** with a natural writing style
- Sentences short (<20 words); plain language
- Bullets or short paragraphs to aid skimming

## Workflow

1. **Research** — web search with date filters; find ≥3 credible sources per story on the given `topic`
2. **Triangulate** — verify facts across sources; discard click-bait or unsourced claims
3. **Headline** — authoritative yet friendly
4. **Summary** — 35-45 words; tease key details
5. **Tone** — 60% informative / 40% conversational
6. **Paraphrase** — do not copy text verbatim
7. **Image** — one royalty-free hero image; fallback to a relevant Unsplash keyword search
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
9. **Filenames** — `{{slug}}.{{lang}}.md` (specified slug if provided, otherwise `slugify(title_en)`)
10. **Validate** — against the project's content schema if one exists

## Gotchas

- A single source isn't triangulation — require ≥3 credible sources before drafting
- Verbatim phrases from a primary source will surface in plagiarism / duplication checks — paraphrase aggressively
- Unsplash fallbacks need topic-relevant keywords, not generic ones — a generic image undercuts the headline
- Slug collisions silently overwrite — slugify with a date suffix if a story repeats on a topic
