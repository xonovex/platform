# merge: Port Elements from One Skill into Another

Extracts elements from a source skill and integrates them into a target skill while strictly preserving the target's structure, style, and voice.

## Core Workflow

1. Track steps in a task list
2. Read target/source skills (including reference files)
3. Analyze target's DNA (structure, style, voice, formatting, conventions)
4. Extract source elements (rules, gotchas, examples, patterns, terminology)
5. Filter by aspects/percentage
6. Ask clarifying questions if interactive mode was requested
7. Rewrite source in target's voice, match formatting exactly
8. Update/create reference files
9. Preview or apply
10. Report summary

## Integration Rules

### Preserve (CRITICAL)

Frontmatter, section order, bullet format, voice/tone, code style, reference file structure, spacing, terminology

### Spec ceilings (CRITICAL)

Target `description` ≤1024 chars after merge; target body <500 lines / ~5000 tokens — push overflow to `references/` with explicit load-when triggers

### Extract from source

New rules, gotchas, enhanced explanations, code examples, reference topics, complementary patterns

### Style matching

Match bullet patterns, bold/italic/code usage, sentence structure, vocabulary, whitespace, heading caps

### Approach

Rewrite in target's voice → insert in existing sections → match format exactly → adapt code style → avoid duplicates → omit material the agent already knows

### Percentage scale

10-30% critical only, 30-50% important (default), 50-70% comprehensive, 70-100% extensive

### Aspect filtering

Extract only specified aspects (e.g., "validation" = validation content only)

## Implementation

### Discovery

Accept SKILL.md paths or names (e.g., `typescript-guide` → `<skills-dir>/typescript-guide/SKILL.md`)

### Analysis

Parse target structure → analyze bullet/formatting patterns → detect voice → extract style rules → build template

### Extraction

Parse source → extract examples/patterns → read reference files → filter by aspects

### Integration

Rewrite in target's voice → insert in existing sections → merge reference files → validate consistency

## Error Handling

- File not found, invalid percentage (10-100), no new content, aspect not found, structure conflict, style detection failed

## Safety

Recommend git commit, never modify frontmatter `name:` and preserve target `description:` triggering quality (imperative "Use when…", trigger contexts, ≤1024 chars), preserve all target content (add only), preview before writing, warn if >40% added or body would exceed 500 lines, abort if style confidence <80%
