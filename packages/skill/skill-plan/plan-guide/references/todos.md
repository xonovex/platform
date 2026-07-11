# todos: Scan and Group TODO Comments

Scan for TODO comments, group by intent, and map each group to a next action. Read-only: produces a research report.

## Technique

- Scan recursively for `TODO:`, `FIXME:`, `NOTE:`; extract and normalize unique messages
- Group by similarity (identical text, conceptual intent, file patterns); count occurrences and affected files
- Infer applicable skills from file extensions / framework indicators
- Categorize by priority (blocking · technical debt · nice-to-have) and map to an action, e.g. high-priority → `plan-create` task, systemic cluster → `plan-research` code-quality audit
