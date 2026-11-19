# Windows

Install the renderer with Chocolatey, or use a container from PowerShell. The general rules are in [rendering.md](rendering.md).

## Local renderer

```powershell
choco install plantuml
plantuml --version
```

The package needs a Java runtime and installs it as a dependency. Sequence diagrams do not need Graphviz. Install it before drawing anything else, and check it with `plantuml -testdot`.

Render from the `diagrams` folder:

```powershell
plantuml --svg -failfast2 -o "$PWD" reminders-lead.puml
```

## Container route

Use Docker Desktop with Linux containers. In PowerShell, `${PWD}` is the current folder:

```powershell
docker info
docker run --rm -v "${PWD}:/data" plantuml/plantuml --svg -failfast2 -o /data /data/reminders-lead.puml
```

The paths after the colon are paths inside the Linux container and use forward slashes. The first run pulls the image from a registry. Check the result with `dir`.

## Structurizr

```powershell
docker run --rm -v "${PWD}:/usr/local/structurizr" structurizr/structurizr export -workspace /usr/local/structurizr/reminders-model.dsl -format plantuml -output /usr/local/structurizr
```
