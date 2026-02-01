---
description: >-
  Create a comprehensive, multi-language travel guide for a specified topic or
  location
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
  [topic] [subject] [--path <path>] [--lang <en,nl>] [--research-only] [--slug
  <slug>]
---

# /content-travelguide-add – Comprehensive Travel Guide Generator

You are a specialist who creates comprehensive, engaging guides that combine storytelling with practical precision.

## Goal

- Create a detailed travel guide for the specified `topic` and `subject` (e.g., topic="City", subject="Barcelona").
- For each guide, produce Markdown files in the specified languages (e.g., `en,nl`).
- Follow the general structure and format for travel guides in this project.
- Combine engaging narrative with actionable, practical information.
- Save the completed travel guides as markdown files in the directory specified by `--path`.

## Arguments

- `topic` (required): The type of travel guide to create (e.g., "Port", "City", "Museum").
- `subject` (required): The specific subject of the travel guide (e.g., "Barcelona", "The Louvre").
- `--path` (required): The destination directory for the generated files.
- `--lang` (optional): Comma-separated list of languages (e.g., `en,nl`). Defaults to `en`.
- `--research-only` (optional): Conduct research and present findings without creating files.
- `--slug` (optional): Custom slug for the generated files. If not provided, the slug is derived from the subject.

## Language & Readability

- **English:** CEFR **B1–B2** with a natural writing style.
- **Dutch:** CEFR **B1–B2** with a natural writing style.
- Keep sentences short (< 20 words) and use plain language.
- Use bullet points or short paragraphs to aid skimming.

## Workflow

1.  **Research**: Use web search to gather comprehensive, current information about the subject.
2.  **Content Creation**: Write the travel guide for each specified language, following the required structure and focusing on a knowledgeable yet friendly tone.
3.  **Front-matter**: Assemble front-matter for each language. Example for English:
    ```yaml
    ---
    title: "{{TITLE_EN}}"
    slug: "{{SLUG}}"
    summary: "{{SUMMARY_EN}}"
    created_at: "{{ISO_TIMESTAMP}}"
    lang: "en"
    ---
    ```
4.  **File Generation**: Save the travel guides as `{{slug}}.{{lang}}.md` in the target directory (use `--slug` if provided, otherwise `{{slugify(subject)}}-guide`).
5.  **Validation**: Ensure all information is web-verifiable and current.

## Writing Philosophy

- **Narrative Style**: Write as a knowledgeable friend sharing insider wisdom.
- **Information Density**: Pack maximum practical value into flowing prose. Use **bold formatting** for critical details like costs, times, and names.
- **Accessibility Focus**: Include accessibility information where relevant.

---

## Required Structure

### 1. Introduction & Welcome (200-300 words)

- Hook the reader with a vivid description.
- Establish key logistics and practical information upfront.
- Provide time recommendations for a visit.

### 2. Logistics & Planning (400-500 words)

- Cover key specifications (location, hours, contact info).
- Detail transportation options, costs, and routes.
- Provide tips for orientation and getting started.

### 3. Main Attractions & Activities (500-600 words)

- Present the essential 3-5 attractions or activities.
- For each, provide: access details, duration, costs, hours, and an insider tip.

### 4. Itineraries & Strategies (600-800 words)

- Provide distinct itineraries for different traveler types (e.g., budget, family, premium).
- Offer time-based plans (e.g., half-day, full-day).

### 5. Essential Practical Information (400-500 words)

- Cover critical details for a smooth visit: payment methods, language tips, weather, safety, and local customs.
