# description: Write a What / Why / How Description

**Guideline:** Open every PR description with what changed, why it changed, and how you approached it, and match the depth to the size of the change.

**Rationale:** The description is the mental model the reviewer evaluates the code against. Without it even clean code gets misread, and a clear one measurably cuts review time. It is also the durable record someone reads months later to learn why the code is the way it is.

**How to Apply:**

1. **What** - one or two sentences on the net effect of the change.
2. **Why** - the engineering or product goal, link the work item or issue.
3. **How** - the approach, plus any alternative you rejected and the reason.
4. Match length to scope: ~50-100 words for a small fix, ~300-400 for a multi-area change.
5. Link context (work items, related PRs, design docs) and attach screenshots for any UI change.

**Example:**

```markdown
// Bad
Fixed the bug.

// Good

## What

Stop double-charging when an import retries.

## Why

PROJ-1234: a retry re-ran the charge step. Make the step idempotent on the import id.

## How

Guard with an `imported` marker row written in the same tx as the charge. Considered a dedup queue but the marker is simpler and sufficient.
```

**Counter-Example:** A one-line typo or version-bump PR needs only a single sentence - do not pad it into headings.

**Related:** [templates.md](./templates.md), [tradeoffs.md](./tradeoffs.md)
