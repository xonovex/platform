---
name: illustrating-plan
description: "Use when the diagrams of a change are to be drawn or redrawn: a lead diagram of what is settled, one diagram for each set of options whose choice changes who calls whom, and an optional system model, saved as PlantUML sources with rendered SVG images, or PNG on request, in the bundle's diagrams directory. Chooses which decisions deserve a picture, renders with a local renderer or else a container runtime, and checks every image against the plan. Triggers on 'draw the diagram', 'a picture of the options', 'render the diagrams', or a diagram that went stale after a decision changed, even when the user doesn't say 'PlantUML'. Not for deciding what the plan says, charts of measured data, or linking the pictures into the plan."
compatibility: "PlantUML renders the diagrams, locally or through Docker or Podman with the plantuml/plantuml image; Graphviz is needed for diagrams other than sequences; the structurizr/structurizr image exports the views of an optional system model. A container route pulls its image from a registry on first use."
---

# Illustrating the Plan

Purpose: give the settled state and each set of options a picture a reader can take in within a minute, and keep every picture true to the decisions. The sources are text and the images are generated; neither is edited by hand after rendering.

**Load first:** **recording-handoff**, and run its entry step before anything below.

## Essentials

- Draw only where the options change the topology, who calls whom. A payload shape or a gate is shown better by a table: no diagram for it.
- The lead diagram shows what is settled and nothing else. Each open decision gets its own diagram with one branch for each option, and an element is marked with its state in color and in text.
- Sources are PlantUML. A Structurizr model is added only when one system model feeds several views.
- Render with a local renderer, else a container runtime. If neither is available, stop at the sources and report the command to run.
- Images are SVG. Render PNG only when the user asks for it, and then keep PNG for every image of the change.
- The outputs are non-markdown files in the `diagrams/` directory of the change: `<topic>-<kind>.puml` with its image next to it. No markdown file is created there.
- This skill links nothing into the plan: it hands the link lines to the step that owns the plan.
- Keep sensitive findings out of every label, note and comment. Never commit or stage. Never start the next step.

## Procedure

1. **Run the entry step** of **recording-handoff**. The diagrams are written into the repository of the change directory, so confirm that branch. The precondition is `plan.md` and at least one decision file. If either is missing, say which step normally writes it and stop.
2. **Read** the plan, the decisions in every state through the decisions index, the research they rest on for the current topology, and a listing of `diagrams/`. Detect the renderer now, with the commands in [references/rendering.md](references/rendering.md), so a missing tool is known before any drawing.
3. **Choose and announce what gets a picture.** Go through the decisions and write one line each in the reply: the decision, and either its picture or the reason for none. Apply [references/selection.md](references/selection.md). Draw the lead diagram only when something is settled, and a set of options only when the options differ in who calls whom.
4. **Write the sources** from [assets/templates/lead-diagram.puml](assets/templates/lead-diagram.puml) and [assets/templates/option-diagram.puml](assets/templates/option-diagram.puml); add [assets/templates/system-model.dsl](assets/templates/system-model.dsl) only when a model helps. Name each file `<topic>-<kind>.puml`, put the render command in a comment at the top, and announce each write in one line. See [references/drawing.md](references/drawing.md) and [references/labels.md](references/labels.md).
5. **Render** to SVG from the `diagrams/` directory with the command for the context: [references/macos.md](references/macos.md), [references/linux.md](references/linux.md) or [references/windows.md](references/windows.md). Use `--png` instead only when the user asked for PNG. A container image is pulled from a registry on first use: say so before the first run. See [references/rendering.md](references/rendering.md).
6. **Open every changed image** and read it against the plan: labels, state marks, width, clipped notes. A picture the change did not touch must come out unchanged. See [references/checking.md](references/checking.md).
7. **Close the run.** Write one `Workflow Log Entry` with `skill: illustrating-plan` and the paths in the one-sentence body, regenerate the log, validate the bundle and report any finding. In the reply, give the link lines for the plan and name the step that owns it. See [references/hand-off.md](references/hand-off.md).
8. **Stop.** Offer writing the links with **writing-plan** or **revising-plan**. Never start either.

## When not to use

- Deciding what the plan says, or which options exist: use **writing-plan** and **settling-questions**.
- Charts of measured data: this skill draws flows and structures, never data.
- Linking a picture into the plan: **writing-plan** and **revising-plan** own the plan.
- A sketch for the user alone: draw it in the reply.

## Gotchas

- A diagram for a payload shape, a field list or a gate. A table shows it more clearly; say so in the choice line and draw nothing.
- An open option in the lead diagram. The lead shows settled elements only; the open ones belong to their own option diagram.
- A picture that states what is open as settled. Mark the state in color and in text, because a printed or color-blind reader loses the color.
- A diagram that renders exactly 4096 pixels wide is cropped, because the renderer limits width and height to 4096. Lay it out top to bottom and shorten the labels. Move a note whose anchor clips it.
- A PNG nobody asked for, or a change that mixes SVG and PNG. Render SVG unless the user asks for PNG, and then render every image of the change as PNG.
- A container run on a Linux host without the user mapping leaves root-owned files. The Linux reference carries the flag.
- A stale label after a decision changed. Search the sources for the old wording before reporting.
- A decision number in a title or label. Display numbers change with the discussion order; use the decision's title.
- An `!include` of a remote address. It needs the network at render time and leaks the machine's address; keep sources self-contained.
- Editing a render by hand, or rendering with a flag the installed renderer does not know. Read the renderer's own help first.
- Creating a markdown file in `diagrams/`, or editing `plan.md` to add the link. Hand the link lines on.

## Done when

- Every `.puml` has an image of the same base name next to it, and every image has a source.
- Every image is SVG, or every image is PNG because the user asked for PNG.
- The lead diagram holds only settled elements; each option diagram shows every option of one decision, and each option appears in its decision file.
- No diagram exists for a payload shape or a gate, and no label carries a hostname, a token, an address or a person's name.
- Each changed image was opened and read against the plan, and an unchanged source rendered an unchanged image.
- The log entry exists, the bundle validates and the reply lists the link lines and their owner.

## Progressive Disclosure

- Read [references/selection.md](references/selection.md) - Load when deciding which decisions get a picture, or when asked to draw a payload, a gate or a field list.
- Read [references/drawing.md](references/drawing.md) - Load when writing or editing a source: file names, the lead and option diagrams, state marks, layout and the system model.
- Read [references/labels.md](references/labels.md) - Load when naming participants or deciding what stays out of a label, a note or a comment.
- Read [references/rendering.md](references/rendering.md) - Load when detecting the renderer, choosing SVG or PNG, rendering or exporting the views of a system model.
- Read [references/checking.md](references/checking.md) - Load when opening a rendered image to check it, or re-rendering after a decision changed.
- Read [references/hand-off.md](references/hand-off.md) - Load when closing the run: the log entry, the link lines and renaming or removing pictures when decisions change.
- Read [references/macos.md](references/macos.md) - Load when working on macOS: installing the renderer, the container route and the Structurizr route.
- Read [references/linux.md](references/linux.md) - Load when working on Linux: the distribution packages, docker and podman with the user mapping.
- Read [references/windows.md](references/windows.md) - Load when working on Windows: the Chocolatey package and the container route from PowerShell.
- Read [assets/templates/lead-diagram.puml](assets/templates/lead-diagram.puml) - Load when drawing the lead diagram.
- Read [assets/templates/option-diagram.puml](assets/templates/option-diagram.puml) - Load when drawing the diagram of one set of options.
- Read [assets/templates/system-model.dsl](assets/templates/system-model.dsl) - Load when one system model feeds several views.
