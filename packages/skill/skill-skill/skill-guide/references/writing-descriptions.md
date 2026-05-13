# writing-descriptions: Writing Skill Descriptions That Trigger Reliably

## Writing Principles

- **Imperative phrasing** — start with "Use when…" / "Use this skill when…", not "This skill does…"
- **User intent over implementation** — match what the user asks for, not internal mechanics
- **Be pushy on triggers** — include "even when the user doesn't say '{keyword}'" for non-obvious matches
- **Concise** — a short paragraph; ≤1024 chars (spec limit)
- **Skip clauses** — explicitly list adjacent skills to take instead (disambiguation)

## Anatomy

A strong description usually contains:

1. **What** — a verb-led sentence describing what the skill does
2. **When to trigger** — situations or prompt shapes that should activate it
3. **Skip / handoff clauses** — pointers to adjacent skills for nearby tasks

## Before / After

```yaml
# Before
description: Process CSV files.

# After
description: >
  Analyze CSV and tabular data — summary stats, derived columns, charts,
  cleaning. Use when the user has a CSV / TSV / Excel file and wants to
  explore, transform, or visualize it, even when they don't say "CSV"
  or "analysis."
```

The improved version is more specific about _what_ and broader about _when_ it applies.

## Gotchas

- A keyword match alone doesn't guarantee triggering — agents skip skills for tasks they can handle alone. Aim the description at tasks needing specialized knowledge.
- Descriptions tend to **grow** during iteration; re-check the 1024-char limit each pass.
- Avoid leaking implementation details ("calls the X API, uses Y library") — those don't help the agent decide _when_ to invoke the skill.
- Don't promise behavior the body doesn't deliver — over-promising descriptions cause false-positive triggers.
