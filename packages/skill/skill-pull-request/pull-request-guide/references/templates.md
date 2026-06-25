# templates: Use a Lean PR Template

**Guideline:** Drive descriptions with a short repo PR template of 4-6 prompts, and cross-link related and stacked PRs.

**Rationale:** A template turns good description habits into the default - nobody forgets to say why or link the work item. Lean is the point: a fifteen-checkbox form gets rubber-stamped, four to six focused prompts get filled in. Cross-links let a reviewer follow a change split across PRs.

**How to Apply:**

1. Add a `PULL_REQUEST_TEMPLATE.md` (host-specific location) so the description field is pre-filled on every PR.
2. Keep it to What / Why / Changes / Testing / Tradeoffs / Related - prompts, not exhaustive checklists.
3. In a coordinated change set, link the sibling PRs to each other and to the shared work item.
4. Match the depth of each section to the size of the PR.

**Example:**

```markdown
## What

## Why

## Changes

## Testing

## Tradeoffs / risks

## Related PRs
```

**Counter-Example:** A long template with many mandatory checkboxes - it raises friction and gets ignored or rubber-stamped.

**Related:** [description.md](./description.md)
