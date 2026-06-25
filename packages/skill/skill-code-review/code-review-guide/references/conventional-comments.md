# Conventional Comments — labels, decorations, format

A consistent format for review feedback. It makes intent explicit for the author and keeps comments machine-parseable. Based on the Conventional Comments standard.

## Format

```
<label> [decorations]: <subject>

[discussion]
```

- **label** — one word naming the kind of comment (required).
- **decorations** — optional, in parentheses, comma-separated, signalling urgency or scope.
- **subject** — the actual message.
- **discussion** — optional extra context, reasoning, or a suggested fix (code block).

In a markdown comment body that spans paragraphs, put the label on its own bold first line, then the body:

```markdown
**issue (blocking)**

When the sheet closes, focus does not return to the button that opened it...
```

## Labels

| Label        | Use for                                                          | Typically blocking? |
| ------------ | ---------------------------------------------------------------- | ------------------- |
| `praise`     | Highlighting something done well. Use genuinely, not as padding. | No                  |
| `nitpick`    | Trivial, preference-level. Almost always `(non-blocking)`.       | No                  |
| `suggestion` | A proposed improvement. State clearly if it blocks.              | Sometimes           |
| `issue`      | A problem with the change (bug, regression, gap).                | Often               |
| `todo`       | A small, necessary change before merge.                          | Usually             |
| `question`   | You need information to assess the code.                         | No                  |
| `thought`    | A non-actionable idea or observation.                            | No                  |
| `chore`      | A process task (changelog, rebase, generated file).              | Varies              |

The list is not rigid — diverge when it helps. The three most-used are `issue`, `suggestion`, `question`.

## Decorations

- `(blocking)` — must be resolved before the change is accepted.
- `(non-blocking)` — should not prevent merge; the author may address it or not.
- `(if-minor)` — resolve at the author's discretion only if the change is small.

The decoration removes ambiguity. `suggestion` without one leaves the author unsure whether it is required; `suggestion (non-blocking)` makes it clearly optional. `question` / `thought` rarely need a decoration because they are inherently non-blocking.

## Mapping severity to merge gating

`(blocking)` is a statement about _your review_. Whether it actually stops the merge depends on the host:

- On most hosts a `(blocking)` label is advisory — a human decides.
- Some hosts have a native enforcement primitive (a task, or a required status, tied to a merge check). When available, mirror your `(blocking)` labels onto that primitive so the gate is enforced, not just stated. See the matching host skill for how.

Keep the two in sync: a label that says blocking but no enforcement is fine (advisory); enforcement with no label is confusing (the author sees a gate with no category).

## Examples

- `issue (blocking)`: Add a null check for `user` before this line — it NPEs when the session expired.
- `suggestion (non-blocking)`: Consider renaming `x` to `userCount` for clarity.
- `nitpick (non-blocking)`: Stray trailing whitespace here.
- `question`: Does the order matter once a thread has won here?
- `praise`: Nice — the snapshot coverage for the new states makes this easy to verify.

Machine-parseable shape (useful when generating or analysing reviews programmatically):

```json
{
  "label": "issue",
  "decorations": ["blocking"],
  "subject": "...",
  "discussion": "..."
}
```
