# Workflow diagrams

Follow the full workflow from research through closeout, including review outcomes, failed checks, worktree delivery, and continuous learning. Use the overview or development cycle when only one part of the process is needed.

| Diagram | Readable image | Editable source |
| --- | --- | --- |
| Main phases | [SVG](workflow-lifecycle.svg), [PNG](workflow-lifecycle.png) | [Graphviz](workflow-lifecycle.dot) |
| Complete workflow | [SVG](workflow-diagram.svg), [PNG](workflow-diagram.png) | [Graphviz](workflow-diagram.dot) |
| Implementation loop | [SVG](development-cycle.svg), [PNG](development-cycle.png) | [Graphviz](development-cycle.dot) |

## Read the flow

Read the numbered phases from top to bottom. Dark gray arrows show the normal flow. Dashed blue arrows show retries, and dark boxes name a return to another phase. Blue diamonds mark decisions; a pale green box marks improved guidance. Continuous learning can run during any phase. Core, Agentic, and Integrations labels identify companion plugins; unqualified commands belong to the Workflow Plugin.

The palette, typography, shapes, and export rules follow [Xonovex design](../../../../../DESIGN.md).

The [workflow guide](../README.md) explains how to select commands and reconstruct context when resuming a task.

## Render changes

Edit the `.dot` sources and run the render task from the repository root in the Nix environment. The task uses the supplied Inter and Space Grotesk fonts and renders every source to SVG and PNG; the plugin check includes it.

```bash
nix develop --no-update-lock-file --command npx moon run plugin-workflow:graph-build --force
```

Keep both rendered formats with their sources so documentation readers can view the flow without installing Graphviz.
