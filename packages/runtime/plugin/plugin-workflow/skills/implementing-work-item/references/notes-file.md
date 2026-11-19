# The notes file

One file per work item records what the run did and why the code looks as it does. It feeds **preparing-delivery**, which writes the commit message and the pull request description, so a decision recorded here is cited there.

## Contents

- Path and keys
- Decision notes
- Checks

## Path and keys

`implementation/<work-id>/notes.md`, where `<work-id>` is the work file's name without `.md`. The type is `Workflow Implementation Notes` and the schema is `workflow/implementation-notes/1`. Use [../assets/templates/implementation-notes.md](../assets/templates/implementation-notes.md). The keys beyond the common header:

| Key | Holds |
| --- | --- |
| `work` | Concept ID of the work file, `work/<slug>-<suffix>`. |
| `worktree` | Path of the worktree. |
| `branch` | The branch the changes sit on. |
| `tasks` | One `{ id, concern, summary }` for each task, in the order they ran. |
| `decision_notes` | One `{ decision, location, note }` for each place a decision shaped code. May be empty. |
| `checks` | One `{ command, result }` for each check you ran. |

## Decision notes

Record a note only where a decision shaped code that the code cannot explain by itself: a constraint the decision imposed, a choice among options the decision settled, a limit the decision set. Do not record a note for code that merely follows from the criteria.

- `decision`: the concept ID of the decision, which must exist in the bundle.
- `location`: `path:line` relative to the root of the work repository, one place for each note. Check each location against the final files, because lines move while tasks run. Do not use a range or a wrapped path.
- `note`: one sentence stating how the decision shaped this code, not what the code does.

**preparing-delivery** may turn a note into an inline comment. Write the note so it still reads true next to the code, with no reference to the session, a task or a plan step.

## Checks

List the commands the criteria name first, then the others. The `result` is what you observed, in one line. Do not copy output beyond that. A secret that appears in output stays out of the file; name the place, never the value.

[closing-the-run.md](closing-the-run.md) says how the file is written and how a rerun edits it.
