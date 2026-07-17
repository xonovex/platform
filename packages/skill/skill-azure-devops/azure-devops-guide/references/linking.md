# Cross-link PRs and Work Items

Reference other pull requests with `!<PRID>` and work items with `#<WorkItemId>` in titles, descriptions, and comments. A coordinated change spanning repos is easier to review when the PRs point at each other, and linking a work item ties the change to its tracked task and updates the board. Both mentions auto-link only inside the same Azure DevOps project.

## Writing the links

In a description or comment, write `!12345` to link PR 12345 and `#6789` to link work item 6789. For a set of coordinated PRs, add a "Related PRs" section listing the siblings by `!<PRID>`, and link the shared work item from each PR so they roll up to it. To attach work items at create time, pass `--work-items <id> [<id> ...]` on `az repos pr create`.

```markdown
## Related PRs

Coordinated set: !12345, !12346, !12347. Work item: #6789.
```

A full PR URL also works but does not render as a first-class link and breaks if the repo is renamed, so prefer `!<PRID>` within the project. The mention only records the association; what the work-item relationship _means_ — parent/child, related, commit, or pull request — and how to verify both sides of it is owned by [boards.md](boards.md).
