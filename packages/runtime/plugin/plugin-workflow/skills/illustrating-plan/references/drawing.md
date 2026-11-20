# Drawing the sources

The files, the diagram kinds, the state marks and the layout. Participant names and what stays out of a source are in [labels.md](labels.md).

## Contents

- Files and names
- The lead diagram
- An option diagram
- State marks
- Layout
- The system model

## Files and names

Every file is `diagrams/<topic>-<kind>.<ext>`, with no markdown file next to it.

| Kind | File | Holds |
| --- | --- | --- |
| `lead` | `<topic>-lead.puml` | The settled state and nothing else |
| `option` | `<topic>-option.puml` | All options of one decision, one branch for each |
| `view` | `<topic>-view.puml` | A view exported from the system model |
| `model` | `<topic>-model.dsl` | The Structurizr model that the views come from |

`<topic>` is a lowercase kebab-case word or two naming the flow or the decision, taken from the decision's slug where one decision owns the picture. Never a display number. The image has the same base name: `<topic>-lead.svg`, or `<topic>-lead.png` when the user asked for PNG. A file is not renamed after it exists, because the plan links to it; a decision that changes meaning gets a new topic and the old pair is removed (see [hand-off.md](hand-off.md)).

The first lines of each source carry the render command as a comment, so the next person re-renders without searching. The templates show the form.

## The lead diagram

A sequence of the settled flow, from [../assets/templates/lead-diagram.puml](../assets/templates/lead-diagram.puml). The title states the flow in the future state. Participants are the components that exist or are decided; every call between them is settled. An open decision leaves no trace in it: no placeholder participant, no dotted "maybe" call.

If a settled decision depends on an open one, draw only the part that does not, and let the option diagram carry the rest.

## An option diagram

From [../assets/templates/option-diagram.puml](../assets/templates/option-diagram.puml). One sequence, shared participants first, then an `alt` block with one branch for each option, named as in the decision's `# Options considered` table. The title is the decision's title without a number. A short note under the block says what the options change, in one line.

- Keep the participants the same in every branch, so the difference is the arrows.
- Mark every branch with the decision's state in text: `[open]`, `[paused]` or `[partly settled]`.
- Show the settled surroundings in the settled color, the options in the open color.

## State marks

Color carries the state and a text label repeats it, because a grayscale print or a color-blind reader loses the color.

| State                  | Color     | Text label  |
| ---------------------- | --------- | ----------- |
| settled                | `#C8E6C9` | `[settled]` |
| open                   | `#FFE0B2` | `[open]`    |
| future target          | `#BBDEFB` | `[future]`  |
| current, to be retired | `#E0E0E0` | `[current]` |

Put the label in the participant name or the note, not only in a legend, because a legend can render without its colors.

## Layout

- Lay out top to bottom and keep labels short. The renderer limits width and height to 4096 pixels, and a wider image is cropped. Wrap long messages with `skinparam maxMessageSize`, which wraps a message at the given width.
- Anchor a note on a participant that is inside the diagram, or it clips.
- Prefer at most six participants. More means two decisions share one picture; split it.
- Number the messages only when a table in the plan refers to the steps.

## The system model

Use [../assets/templates/system-model.dsl](../assets/templates/system-model.dsl) when several views share elements. Tag every element and relationship with its state (`Settled`, `Open`, `Future`, `Current`) and give each view only the elements its question needs: `include` what belongs, and `exclude "element.tag==Open"` on a view that shows the settled state. Give each view a key of the form `<topic>-view`. Export with the format `plantuml` (the same as `plantuml/structurizr`), list the output directory, and move an exported file whose name differs from its key to `<topic>-view.puml`. Render as in [rendering.md](rendering.md) and treat the exported `.puml` like any other source: generated, never edited by hand. Edit the model and export again.
