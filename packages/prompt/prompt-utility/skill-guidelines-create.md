---
description: Create a guideline skill from a provided document or URL
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - AskUserQuestion
  - TodoWrite
argument-hint: "[source] [--name <name>] [--dry-run]"
---

# /skill-guidelines-create – Create Guideline Skill from Document

Generate a guideline skill with progressive disclosure structure from a document file or URL.

## Goal

- Extract coding guidelines from external sources (docs, blog posts, style guides)
- Generate structured SKILL.md with essentials and progressive disclosure
- Create detail files for complex topics with examples
- Make content project-independent and reusable

## Usage

```bash
# From URL
/skill-guidelines-create "https://example.com/react-best-practices" --name react-guidelines

# From local file
/skill-guidelines-create "./docs/coding-standards.md" --name typescript-guidelines

# Preview without writing
/skill-guidelines-create "https://example.com/go-style-guide" --name go-guidelines --dry-run
```

## Arguments

- `source` (required): URL or file path to the source document
- `--name` (required): Skill name in kebab-case (e.g., `react-guidelines`, `go-guidelines`)
- `--dry-run` (optional): Preview generated structure without writing files

## Core Workflow

1. **Fetch Source**: Use WebFetch for URLs, Read for local files
2. **Extract Guidelines**: Identify requirements, rules, patterns, anti-patterns, and code examples
3. **Categorize Content**: Group by topic (e.g., performance, architecture, testing, security)
4. **Generate SKILL.md**: Create condensed quick reference with essentials (3-7 items)
5. **Create Detail Files**: Extract code examples and explanations to `details/*.md`
6. **Validate Structure**: Ensure all @references point to existing files
7. **Write Files**: Save to `.claude/skills/{name}/` or preview with `--dry-run`

## Output Structure

**SKILL.md:**

````markdown
---
name: {name}
description: "Use when working with {technology}. Apply for {scenarios}."
---

# {Title} Coding Guidelines

## Requirements

- {Technology} ≥ {version}

## Essentials

- {Core guideline 1}
- {Core guideline 2-6}

## Example

```{language}
{Representative code showing best practices}
```
````

## Progressive disclosure

- **details/{topic}.md** - When {scenario}

````

**Detail file (details/{topic}.md):**
```markdown
# {topic}: {Title}

**Guideline:** {Rule statement}
**Rationale:** {Why this matters}

**How to Apply:**
1. {Step-by-step}

**Example:**
```{language}
// Bad
{Anti-pattern}
// Good
{Correct usage}
````

```

## Implementation Details

**Source Parsing:**
- Extract headings as topic groups
- Identify code blocks with language markers
- Detect "do/don't", "good/bad", "prefer/avoid" patterns
- Parse bullet lists as individual guidelines

**Content Condensation:**
- SKILL.md: 3-7 essential bullets, one representative example
- Detail files: Full explanations, multiple examples, rationale
- Bullet format: `- **Rule** - Brief 5-10 word how-to`

**Topic Detection:** Group guidelines by: architecture, performance, testing, security, error-handling, naming, state-management, accessibility

## Error Handling

- URL unreachable: `Error: Could not fetch [url]`
- File not found: `Error: File not found at [path]`
- No guidelines found: `Warning: Could not extract guidelines from source`
- Skill exists: Ask to merge or overwrite
- Name not kebab-case: `Error: Skill name must be kebab-case`

## Safety

- Preview with `--dry-run` before writing
- Check for existing skill directory
- Preserve existing detail files when merging
- Remove source-specific paths and project names
```
