# Sources

## PlantUML command line

- **URL:** https://plantuml.com/command-line
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/rendering.md` → The render command
  - `references/linux.md`, `references/macos.md`, `references/windows.md` → the local and container render commands
  - `assets/templates/lead-diagram.puml`, `assets/templates/option-diagram.puml` → the render comment
- **Aspects extracted:**
  - `-o <directory>` writes the images into a directory → rendering
  - `--svg` and `--png` choose the output format → rendering, templates
  - `-failfast2` runs a checking pass and generates nothing when any diagram has an error → rendering

## PlantUML FAQ

- **URL:** https://plantuml.com/faq
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Gotchas
  - `references/drawing.md` → Layout
  - `references/checking.md` → Check the images
- **Aspects extracted:**
  - The renderer limits image width and height to 4096 pixels and crops a wider image → Gotchas, layout

## PlantUML installation FAQ

- **URL:** https://plantuml.com/faq-install
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/macos.md`, `references/windows.md` → Local renderer
- **Aspects extracted:**
  - Sequence diagrams render without Graphviz, and `-testdot` checks Graphviz → macOS, Windows

## Structurizr export

- **URL:** https://docs.structurizr.com/export
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/rendering.md` → The Structurizr export
- **Aspects extracted:**
  - The `plantuml` format is the same as `plantuml/structurizr` and exports each view as standard PlantUML → rendering

## Structurizr DSL expressions

- **URL:** https://docs.structurizr.com/dsl/expressions
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/drawing.md` → The system model
  - `assets/templates/system-model.dsl` → the settled view
- **Aspects extracted:**
  - A view excludes every element with a tag through an element tag expression, so the settled view leaves the open elements out → drawing, template

## Structurizr binaries

- **URL:** https://docs.structurizr.com/binaries
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/rendering.md` → The Structurizr export
  - `references/macos.md`, `references/linux.md`, `references/windows.md` → Structurizr
- **Aspects extracted:**
  - The Structurizr image runs a command against a directory mounted at `/usr/local/structurizr` → rendering, operating-system references

## podman run

- **URL:** https://docs.podman.io/en/latest/markdown/podman-run.1.html
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/linux.md` → podman
- **Aspects extracted:**
  - `--userns=keep-id` maps the current user and group to the same ids in the container → Linux
  - The `Z` volume suffix relabels the mounted content privately for SELinux → Linux

## Homebrew formula for plantuml

- **URL:** https://formulae.brew.sh/api/formula/plantuml.json
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/macos.md` → Local renderer
- **Aspects extracted:**
  - The formula installs the `plantuml` command with Graphviz and OpenJDK as dependencies → macOS

## Debian plantuml package

- **URL:** https://packages.debian.org/stable/plantuml
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/linux.md` → Local renderer
  - `references/rendering.md` → Choose the route
- **Aspects extracted:**
  - Debian stable ships a 2020 release that only recommends Graphviz, so the packaged renderer can predate the switches → Linux, rendering

## Xonovex diagram conventions

- **Provenance:** Repository-original rules for the diagrams of a change: which decisions get a picture, the lead and option diagrams, the state marks, the file names, participant names, SVG by default with PNG only on request, the render route, the image check, the log entry and the hand-off of link lines to the owner of the plan
- **Last reviewed:** 2026-11-19
- **References:** references/selection.md, references/drawing.md, references/labels.md, references/rendering.md, references/checking.md, references/hand-off.md, references/linux.md, references/macos.md, references/windows.md
- **Aspects extracted:** Drawing only where options change who calls whom, a lead diagram of settled elements only, one option diagram for each open decision, state marks in color and in text, `<topic>-<kind>` names, SVG images unless the user asks for PNG, a local renderer before a container runtime, opening every changed image, and handing the link lines to **writing-plan** and **revising-plan** are conventions of this workflow. The entry step, the log entry and the bundle layout come from **recording-handoff**.
