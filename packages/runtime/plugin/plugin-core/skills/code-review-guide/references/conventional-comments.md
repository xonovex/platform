# Conventional Comments: labels, decorations, format

Based on the Conventional Comments standard.

## Format

```
<label> [decorations]: <subject>

[discussion]
```

In a multi-paragraph markdown body, put `<label> (decoration)` as its own bold first line, then the body.

## Labels

| Label        | Use for                                        | Typically blocking? |
| ------------ | ---------------------------------------------- | ------------------- |
| `praise`     | Something done well (genuine, not padding).    | No                  |
| `nitpick`    | Trivial, preference-level.                     | No                  |
| `suggestion` | A proposed improvement.                        | Sometimes           |
| `issue`      | A problem (bug, regression, gap).              | Often               |
| `todo`       | A small, necessary change before merge.        | Usually             |
| `thought`    | A non-actionable idea or observation.          | No                  |
| `chore`      | A process task (changelog, rebase, generated). | Varies              |

Diverge when it helps. Most-used: `issue`, `suggestion`, `question`.

## Decorations

- `(blocking)`: must be resolved before the change is accepted.
- `(non-blocking)`: may be addressed or not.
- `(if-minor)`: resolve at author's discretion only if the change is small.

`suggestion` without a decoration leaves urgency ambiguous. `question` / `thought` are inherently non-blocking and rarely need one.

## Mapping severity to merge gating

`(blocking)` is a statement about your review, not necessarily a merge gate:

- On most hosts it is advisory: a human decides.
- If the host has a native enforcement primitive (a task, or a required status tied to a merge check), mirror `(blocking)` labels onto it so the gate is enforced. See the host skill.

Keep the two in sync: label-without-enforcement is fine (advisory); enforcement-without-label confuses the author.

## Examples

- `issue (blocking)`: Add a null check for `user`. It NPEs when the session expired.
- `suggestion (non-blocking)`: Consider renaming `x` to `userCount`.

Machine-parseable shape:

```json
{
  "label": "issue",
  "decorations": ["blocking"],
  "subject": "...",
  "discussion": "..."
}
```
