# create: Curate Latest News Stories

Auto-curate the latest news and developments on a given topic, then produce CEFR B1-B2 bilingual articles as Markdown with structured frontmatter.

## Goal

- Fetch the latest news on a specified `topic`
- Default lookback: 7 days (configurable)
- For each vetted story, produce one Markdown file per requested language (e.g. `en,nl`)
- Save to the specified target directory (e.g. `src/content/news/`)

Apply **editorial-writing-guide** for the requested CEFR level, natural register, and multilingual consistency.

## Checklist

- [ ] **Research**: Search with date filters and find at least three credible sources per story on the given `topic`.
- [ ] **Triangulate**: Verify the central facts across sources and discard clickbait or unsupported claims.
- [ ] **Headline**: Write an authoritative and approachable headline.
- [ ] **Summary**: Write 35 to 45 words that state the news value without giving every detail.
- [ ] **Tone**: Keep the article informative and use conversational phrasing only when the publication permits it.
- [ ] **Paraphrase**: Do not copy source text verbatim.
- [ ] **Image**: Select one relevant image with verified reuse rights. Use a relevant royalty-free image search only as a fallback.
- [ ] **Frontmatter**: Assemble one record per language:
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
- [ ] **Filenames**: Use `{{slug}}.{{lang}}.md`. Use the supplied slug or derive it from the source-language title.
- [ ] **Validate**: Check for an existing slug, then validate against the project's content schema when one exists.

## Gotchas

- A single source isn't triangulation: require ≥3 credible sources before drafting
- Verbatim phrases from a primary source will surface in plagiarism / duplication checks: paraphrase aggressively
- An image search result does not prove reuse rights. Verify the license before publication.
- Slug collisions silently overwrite: slugify with a date suffix if a story repeats on a topic
