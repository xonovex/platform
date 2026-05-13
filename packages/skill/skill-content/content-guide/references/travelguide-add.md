# travelguide-add: Generate Comprehensive Travel Guides

Create detailed, multilingual travel guides combining engaging narrative with practical, web-verifiable information.

## Goal

- Create a detailed travel guide for the specified `topic` and `subject` (e.g. topic=`City`, subject=`Barcelona`)
- Produce one Markdown file per requested language (e.g. `en,nl`)
- Follow the project's travel-guide structure (see Required Structure)
- Combine engaging narrative with actionable practical information
- Save to the specified target directory

## Language & Readability

- **Default:** CEFR **B1-B2** with a natural writing style
- Sentences short (<20 words); plain language
- Bullets or short paragraphs to aid skimming

## Workflow

1. **Research** — web search to gather comprehensive, current information about the subject
2. **Write** — produce content per language; knowledgeable yet friendly tone
3. **Front-matter** — assemble per language:
   ```yaml
   ---
   title: "{{TITLE_EN}}"
   slug: "{{SLUG}}"
   summary: "{{SUMMARY_EN}}"
   created_at: "{{ISO_TIMESTAMP}}"
   lang: "en"
   ---
   ```
4. **Filenames** — `{{slug}}.{{lang}}.md` (specified slug if provided, otherwise `slugify(subject)-guide`)
5. **Validate** — ensure all information is web-verifiable and current

## Writing Philosophy

- **Narrative style** — knowledgeable friend sharing insider wisdom
- **Information density** — pack maximum practical value into flowing prose; **bold** critical details (costs, times, names)
- **Accessibility** — include accessibility information where relevant

## Required Structure

1. **Introduction & Welcome** (200-300 words) — vivid hook, key logistics upfront, time recommendations
2. **Logistics & Planning** (400-500 words) — location, hours, contact, transportation, costs, routes, orientation tips
3. **Main Attractions & Activities** (500-600 words) — 3-5 essentials; per entry: access, duration, costs, hours, insider tip
4. **Itineraries & Strategies** (600-800 words) — distinct itineraries per traveler type (budget, family, premium); time-based plans (half-day, full-day)
5. **Essential Practical Information** (400-500 words) — payment methods, language tips, weather, safety, local customs

## Gotchas

- Hours / prices / contact info go stale fast — note the research date in the summary or footer so a future refresh knows what to re-verify
- Multi-language versions must agree on facts — produce English first, then translate; don't re-research per language
- "Insider tip" that's actually a Wikipedia paraphrase isn't insider — name a source, an experience, or cut it
- Accessibility info that's not verified is worse than omitting it — never invent ramp / elevator / restroom claims
