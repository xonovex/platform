# Rendering

## Contents

- Detect the renderer
- Choose the route
- The render command
- The Structurizr export

## Detect the renderer

Run these before drawing, from the `diagrams/` directory, and report what they show:

```bash
command -v plantuml docker podman
plantuml --version
docker info
ls
```

`command -v` lists which of the three programs are on the path; the others are absent. `plantuml --version` prints the PlantUML and Java versions. `docker info` fails when the daemon is stopped or the user may not reach it. Use the equivalent of each command on Windows from the Windows reference.

## Choose the route

1. A local `plantuml` on the path: use it. It needs no network and no image pull.
2. Otherwise a container runtime that you can reach: docker, or podman where docker is not accessible. The first run pulls the image from a registry, so tell the invoker before it runs.
3. Neither: write the sources, skip the render, and give the invoker the install steps for the operating system and the render command. Say plainly that no image exists yet. Do not state a picture as done.

Read the renderer's own help before the first render and compare the switches below with it. A packaged renderer can be older than the switches here: the Debian stable package, for example, is a 2020 release. When a switch is unknown, use the format and output switches the installed help lists.

## The render command

Run from the `diagrams/` directory, one source at a time or all of them. The switches, as the PlantUML command-line page documents them:

- `--svg` or `--png` chooses the image format. Render SVG, because it stays sharp and diffs as text. Render PNG only when the user asks for it, and then render every image of the change as PNG, so one format holds in `diagrams/`.
- `-o <directory>` writes the images into the directory. Pass the `diagrams/` directory itself so the image lands next to its source.
- `-failfast2` runs a checking pass first and generates nothing when any diagram has an error.

A local render:

```bash
plantuml --svg -failfast2 -o "$PWD" reminders-lead.puml reminders-option.puml
```

A container render mounts the directory at `/data`, so every path inside the container is under `/data`. The operating-system references give the full command, including the flag that keeps the files owned by you:

```bash
docker run --rm -v "$PWD:/data" plantuml/plantuml --svg -failfast2 -o /data /data/reminders-lead.puml
```

The image takes the renderer's switches after its name. Run it once with `--help` and compare before relying on the command.

The commands here and in the operating-system references render SVG. When the user asked for PNG, put `--png` in place of `--svg` in each of them.

Put the command you used in a comment on the first lines of each source, so the next run reuses it.

## The Structurizr export

When a model feeds the views, validate it, export the views to PlantUML, then render the exported sources as above. The Structurizr image mounts a directory at `/usr/local/structurizr`:

```bash
docker run --rm -v "$PWD:/usr/local/structurizr" structurizr/structurizr validate -workspace /usr/local/structurizr/reminders-model.dsl
docker run --rm -v "$PWD:/usr/local/structurizr" structurizr/structurizr export -workspace /usr/local/structurizr/reminders-model.dsl -format plantuml -output /usr/local/structurizr
```

`-format plantuml` is the same as `plantuml/structurizr` and exports each view as standard PlantUML. Use it for every view.
