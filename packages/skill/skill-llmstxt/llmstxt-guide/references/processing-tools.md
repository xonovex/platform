# processing-tools: Expanding /llms.txt into Single-File Context Bundles

`/llms.txt` is a curated index of links to `.md` mirrors. Processing tools follow those links and inline the content into a single file suitable for one-shot LLM ingestion.

## Output Variants

Per the spec, two variants are produced:

- **`llms-ctx.txt`** — expansion of `/llms.txt` **without** the `## Optional` URLs (shorter context)
- **`llms-ctx-full.txt`** — expansion **with** the `## Optional` URLs (full context)

Both are single Markdown files containing:

1. The original `/llms.txt` header (H1, blockquote, prose)
2. For each linked `.md` page: a section header (the link title) followed by the page's markdown content inlined

## Reference Tool: `llms_txt2ctx`

The official Python tool (from the `llms-txt` library) produces both variants from a given `/llms.txt`:

```bash
# Install
pip install llms-txt

# Generate the short variant (no Optional)
llms_txt2ctx https://example.com/llms.txt > llms-ctx.txt

# Generate the full variant (includes Optional)
llms_txt2ctx --optional https://example.com/llms.txt > llms-ctx-full.txt
```

Exact CLI flags depend on the tool version — `--help` is authoritative.

## What the Tool Does

1. Fetches `/llms.txt`
2. Parses the H1, blockquote, prose, and H2 sections
3. For each H2 section (except `## Optional` when producing the short variant):
   - Fetches each linked `.md` URL
   - Emits a section header for the link title
   - Inlines the fetched markdown content
4. Writes the assembled file

## Hosting the Outputs

Some projects serve the expanded variants directly so callers don't have to run the tool:

- `/llms.txt` — curated index (links)
- `/llms-ctx.txt` — pre-expanded short variant
- `/llms-ctx-full.txt` — pre-expanded full variant
- `/llms-full.txt` — informal single-file dump (community convention, not spec)

Pre-hosting is convenient when the docs change infrequently — generate on build, deploy alongside.

## Distinction: `llms-ctx-full.txt` vs `llms-full.txt`

- **`llms-ctx-full.txt`** — spec output of expanding `/llms.txt` + its linked mirrors (including Optional)
- **`llms-full.txt`** — informal single-file dump of all docs; **not** part of the spec

Both end in similar files in practice but the provenance differs: one is generated from a curated index, the other is a raw dump. Hosting both is fine; using the names interchangeably is confusing.

## Integration With Build Pipelines

Typical setup:

1. CI/CD regenerates per-page `.md` mirrors when source content changes
2. CI/CD runs `llms_txt2ctx` against the live `/llms.txt`
3. Both `llms-ctx.txt` and `llms-ctx-full.txt` are deployed alongside the site

When `/llms.txt` itself rarely changes (mostly just link list and descriptions), regenerating the expanded bundles on each docs build keeps them in sync without manual intervention.

## Gotchas

- A broken `.md` mirror surfaces as a missing section in `llms-ctx.txt` — verify all linked URLs `200 OK` before generating
- Re-fetching every link on every build is slow for large doc sets — cache by `ETag` / `Last-Modified` and invalidate per file
- The short variant (`llms-ctx.txt`) only differs from the full one by the `## Optional` section — if you didn't mark anything Optional, the two files are identical
- Hosting only the expanded variants without the source `/llms.txt` loses the curated structure tooling expects to find — always serve `/llms.txt` itself
- Re-naming `## Optional` to `## Extra` or similar means processors won't recognize it — the short variant will include those URLs anyway, breaking the short-context promise
