# Cut proposal template

Copy the block into the reply that proposes the cut, replace every angle-bracket placeholder and delete any comment lines. The reply ends after the closing line, and nothing is written until the operator approves.

```markdown
## Proposed cut for <change>: <N> work items for <team>

<One sentence: how many items, which package or none, and how many rows stay as outline.>

### Work package

| Package | Action                                    | Team   | Objective            |
| ------- | ----------------------------------------- | ------ | -------------------- |
| <title> | <new, or existing `work/<slug>-<suffix>`> | <team> | <the shared outcome> |

<!-- Keep the table when the plan calls for a package or the invoker named one. Otherwise delete it and keep the next line; delete the next line when the table stays. -->

No work package, and why: <the reason the items need no shared parent>.

### Work items

| Item    | Action                                                     | Repository   | Depends on      | Decisions        |
| ------- | ---------------------------------------------------------- | ------------ | --------------- | ---------------- |
| <title> | <define, split from `<file>`, new, or merge into `<file>`> | <repository> | <files or none> | <decision files> |

### Criteria

**<Item title>**

| Id  | Condition              | Verification        |
| --- | ---------------------- | ------------------- |
| c1  | <observable condition> | <how it is checked> |

### Rows left as outline

| Row     | Team   | Waits on         |
| ------- | ------ | ---------------- |
| <title> | <team> | <open decisions> |

<!-- Include "Rows left as outline" when any row of the team waits. -->

### Not touched

<Rows of other teams, one line each with the team. Include rows whose decisions have since settled and need revising-plan.>

### Questions

<Only what the operator must answer before approving: a team, a repository, a split that follows repositories. Delete the section when there is none.>

Reply "approve the cut" to write these files, or name the changes you want first.
```

## Fill in

- The title line names the change and the team whose rows are defined.
- One row for each item, in dependency order. Every criterion has its verification; a proposal with a criterion that lacks one is not ready to send.
- "Action" tells the operator what happens to the file: `define` completes an outline row, `split` and `merge` follow [the outline row rules](../../references/outline-rows.md), `new` is a file that does not exist yet.
- Say in one line why a split follows repositories or teams when it does.
- When the cut has no package, keep the "No work package, and why" line and delete the table. A cut with no package is a valid result; the line says why.
- Never ask for the confirmation of the definitions in the proposal. That question comes after the files are written and shown.
- Do not include file contents, frontmatter or the next step in the proposal.
