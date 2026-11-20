# Closing the run

## Contents

- The log entry
- The link lines
- When decisions change
- The reply

## The log entry

Write one `Workflow Log Entry` for the run, from the log entry template of `recording-handoff`, as `log/<date>-illustrating-plan-<suffix>.md`.

- `skill: illustrating-plan` and `action: Update`, because a diagram is not a concept document. `generated.by` is `illustrating-plan/<version>`.
- No `touched` key: no concept ID names a diagram, and the list is empty.
- The one-sentence body names the files: for example, "Drew the lead diagram and the option diagram for the transport decision as `diagrams/reminders-lead.puml` and `diagrams/reminders-option.puml`, rendered to SVG."
- Then regenerate `log.md` with the `recording-handoff` script, validate the bundle and report any finding. `diagrams/` has no generated index.

## The link lines

This skill does not edit `plan.md` or a decision. It hands over the lines, in the reply, with their owner:

- For the plan: an image embed line whose alt text names the flow, such as "The settled reminder flow", and whose path is `diagrams/reminders-lead.svg`, under the objective or the section it illustrates, with the `.png` path when the user asked for PNG. **writing-plan** and **revising-plan** own the plan.
- For any other document the user names: the same embed line with the path relative to that document, because an embed resolves from the file that holds it.

## When decisions change

Pictures follow the decisions; the decisions never follow the pictures.

- A decision settles: delete its option diagram and image, and add the settled elements to the lead diagram. Re-render the lead.
- An option is added or dropped: change the branches of the option diagram to match the decision's options table and re-render. A branch for a dropped option goes.
- A decision changes meaning, not only its state: draw a new topic, remove the old pair and list the links to move.
- A decision is superseded: draw from the superseding decision only.
- A title changes: change the title in the source and re-render. The file name stays.

List in the reply every link that now points to a removed file, with the document that owns it.

## The reply

Lead with the outcome: which pictures were drawn, redrawn or removed, and which decisions got none and why. Then the render result, the checks done on the images, the link lines with their owner and anything not done (no renderer found, an image not opened). Keep sensitive findings in the reply only when the invoker needs them, never in a file.
