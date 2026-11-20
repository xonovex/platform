# Workflow diagrams

Follow a change from intent through planning, the execution of each work item and the validation of the work package.

| Diagram | Readable image | Editable source |
| --- | --- | --- |
| Change flow | [SVG](change-flow.svg), [PNG](change-flow.png) | [Graphviz](change-flow.dot) |

## Read the flow

Read the planning panel from left to right, then the execution panel, which repeats for each work item. Dark gray arrows show the normal flow, and the dashed blue arrow returns to the next work item. A box names the step and the skill that owns it; a `Core:` prefix names a skill of the Core Plugin. The dark box marks the validation of the work package.

The palette, typography, shapes, and export rules follow the Xonovex design system in the `drodan-platform` repository (`packages/xonovex/xonovex-design-system`), as mirrored in `packages/asset/asset-brand/diagrams/style.gvpr`.

The [workflow guide](../README.md) explains each step and the files it reads and writes.

## Render changes

Edit the `.dot` sources and run the render task from the repository root in the Nix environment. A source names a role, such as `role=decision`, and sets no colour and no font; `diagram-role-check` fails on either and on an unknown role. The render task applies the Graphviz style that `@xonovex/asset-brand` mirrors from the Xonovex design system, uses the supplied Inter and Space Grotesk fonts, and renders every source to SVG and PNG. It replaces an export only when the render changed, so a run without a change leaves the working tree clean. The plugin check includes both tasks.

```bash
nix develop --no-update-lock-file --command npx moon run plugin-workflow:graph-build
```

Keep both rendered formats with their sources so documentation readers can view the flow without installing Graphviz.
