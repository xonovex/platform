# description: Write a What / Why / How Description

Open with what changed, why (link the work item/issue), and how you approached it (including any rejected alternative and the reason). Match length to scope: ~50-100 words for a small fix, ~300-400 for a multi-area change. Attach screenshots for any UI change. A one-line typo or version-bump PR needs a single sentence, not headings.

```markdown
## What

Stop double-charging when an import retries.

## Why

PROJ-1234: a retry re-ran the charge step. Make it idempotent on the import id.

## How

Guard with an `imported` marker row written in the same tx as the charge. Considered a dedup queue but the marker is simpler and sufficient.
```

## Related

[templates.md](./templates.md), [tradeoffs.md](./tradeoffs.md)
