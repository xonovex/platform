# simplify: Condense a Skill's Text

Reduce a verbose skill to its essentials — condense the `SKILL.md` body and trim bloated reference files — while preserving the quick reference and keeping content project-independent.

Spec limits (name / description / body / optional-field rules) live in the parent SKILL.md — re-check after edits.

## Contents

- [Condensing SKILL.md](#condensing-skillmd) — 60-80% reduction, bullet format, extract examples to references
- [Condensing reference files](#condensing-reference-files) — merge overlapping sections, 20-40% reduction
- [Error Handling](#error-handling)
- [Safety](#safety)

## Condensing SKILL.md

### Goal

Reduce SKILL.md by 60-80% (default 70%) while preserving quick reference. Extract examples to reference files. Land under the spec ceiling (<500 lines / ~5000 tokens body).

### Core Workflow

1. Track steps in a task list
2. Read SKILL.md and measure baseline
3. Extract code examples to reference files (group by topic)
4. Simplify SKILL.md to a bullet list
5. Remove project-specific references
6. Preview or write changes
7. Report metrics

### Simplification Rules

#### Remove from SKILL.md

code blocks, prose paragraphs, long explanations, project paths/names, example sections, duplicate content, anything the agent already knows by general training

#### Keep in SKILL.md

frontmatter, one-line description, section headings, bullet points with rule + brief how-to + link, small inline code, a **Gotchas** section for non-obvious env-specific facts

#### Bullet format

`- **Rule** - Brief 5-10 word how-to (references/{topic}.md)`

#### Reference triggers

each reference link states a load-when condition (e.g., "Load when API returns non-200"), not a generic "see X"

#### Create reference files when

code examples exist, detailed explanation needed, multiple examples, or counter-examples to show

#### Defaults over menus

never leave 3+ equal options when condensing — pick one default, mention alternatives briefly

#### Project-independence

replace specific project names/paths/domains with generic equivalents ("your app", "project root")

### Success Metrics

Report: lines removed ([X]% reduction), reference files created/updated, project references removed, final size vs target.

## Condensing reference files

### Guideline

Reduce `references/*.md` files by merging overlapping sections, trimming redundant examples, and removing prose that restates headings.

### Rationale

Reference files are the detail layer of progressive disclosure. When they grow verbose with overlapping sections or redundant examples, they defeat their purpose as scannable technical references.

### Techniques

- Read the full reference file and identify overlapping sections
- Merge sections that cover the same concept from different angles
- Consolidate multiple code examples into fewer that demonstrate more
- Remove prose that restates what the heading already says
- Remove redundant code examples already shown in the top-level example block
- Preserve all distinct technical content — **merge, don't delete**
- Target 20-40% reduction (reference files are already the detail layer)
- Do not move content to sub-files — reference is the bottom of the disclosure hierarchy

### Example

```markdown
# BEFORE (task-inheritance.md, 139 lines, 3 overlapping merge sections)

## Deep Merging

- Configs merge sequentially: global → extends → local
- command arrays merge (use script: to fully override)

## Preventing Merge Issues

- Use script: instead of command: to fully replace inherited commands

## Merge Strategy Overrides

- Merging applies to args, deps, env, inputs, outputs, toolchains...

# AFTER (task-inheritance.md, 110 lines, single Merging section)

## Merging

Configs merge sequentially: global → extends → local.

- command arrays merge (use script: to fully replace)
- args, deps, env, inputs, outputs, toolchains merge via configurable strategies

# Reduction: 3 sections → 1, redundant examples removed, all content preserved
```

## Error Handling

- **File not found:** `Error: SKILL.md not found at [path]`
- **Already minimal:** `Skipping [file]: Already minimal at [N] lines`
- **Invalid target:** `Target reduction must be between 50-90%`
- **Broken references:** reference points to a non-existent file

## Safety

- Recommend a git commit before running
- Never modify the skill `name:` in frontmatter
- Preserve `description:` triggering quality (imperative, ≤1024 chars, keeps trigger contexts)
- Skip skills <30 lines
- Warn if reference files would be overwritten
