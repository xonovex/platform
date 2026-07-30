# writing-descriptions: Writing Skill Descriptions That Trigger Reliably

## Writing Principles

- **Imperative phrasing**: start with "Use when..." / "Use this skill when...", not "This skill does..."
- **Third person**: describe the skill, not the user/agent; avoid "I can..." / "You can..." (imperative "Use when..." is fine, neither first nor second person)
- **User intent over implementation**: match what the user asks for, not internal mechanics
- **Be pushy on triggers**: include "even when the user doesn't say '{keyword}'" for non-obvious matches
- **Concise**: a short paragraph; ≤1024 chars (spec limit)
- **Positive routing only**: describe what the skill is for and let triggers route; do not add "Skip ..." clauses, enumerate out-of-scope work, or point to other skills by name in the description

## Before / After

```yaml
# Before
description: Process CSV files.

# After
description: "Use when analyzing CSV and tabular data through summary statistics, derived columns, charts, or cleaning. Triggers on CSV, TSV, Excel, tables, columns, data exploration, transformation, or visualization, even when the user doesn't say 'CSV' or 'analysis'."
```

## Gotchas

- A keyword match alone doesn't guarantee triggering: agents skip skills for tasks they can handle alone. Aim the description at tasks needing specialized knowledge.
- Descriptions tend to **grow** during iteration; re-check the 1024-char limit each pass.
- Avoid leaking implementation details ("calls the X API, uses Y library"): those don't help the agent decide _when_ to invoke the skill.
- Don't promise behavior the body doesn't deliver: over-promising descriptions cause false-positive triggers.
- Quote the `description` as a double-quoted YAML scalar with single-quoted inner phrases and literal apostrophes (e.g. `"... even when the user doesn't say 'TDD'."`): the skill loader's parser rejects single-quote `''` escaping and stops at an inner double-quote, then reads the tail as an unknown attribute.
