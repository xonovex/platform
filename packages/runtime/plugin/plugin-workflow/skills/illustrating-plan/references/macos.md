# macOS

Install the renderer with Homebrew, or use a container. The general rules are in [rendering.md](rendering.md).

## Local renderer

```bash
brew install plantuml
plantuml --version
```

The formula installs the `plantuml` command and, as dependencies, Graphviz and OpenJDK. Sequence diagrams do not need Graphviz; the views exported from a Structurizr model can. Check Graphviz with `plantuml -testdot`.

Render from the `diagrams/` directory:

```bash
plantuml --svg -failfast2 -o "$PWD" reminders-lead.puml
```

## Container route

Use Docker Desktop or any docker-compatible runtime. Check the daemon, then render:

```bash
docker info
docker run --rm -v "$PWD:/data" plantuml/plantuml --svg -failfast2 -o /data /data/reminders-lead.puml
```

The first run pulls the image from a registry. Check the ownership of the images with `ls -l`. When they are owned by root, add `-u "$(id -u):$(id -g)"` after `--rm`.

## Structurizr

The maintained route is the Structurizr image:

```bash
docker run --rm -v "$PWD:/usr/local/structurizr" structurizr/structurizr export -workspace /usr/local/structurizr/reminders-model.dsl -format plantuml -output /usr/local/structurizr
```

A community Homebrew formula, `structurizr-cli`, installs the older command line tool as `structurizr-cli` with the same `export` and `validate` commands. The Structurizr project does not maintain it, so prefer the image.
