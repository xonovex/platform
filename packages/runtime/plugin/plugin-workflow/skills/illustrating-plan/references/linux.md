# Linux

Install the renderer from the distribution, or use docker or podman. The general rules are in [rendering.md](rendering.md).

## Local renderer

```bash
sudo apt install plantuml graphviz   # Debian and Ubuntu
sudo dnf install plantuml            # Fedora
plantuml --version
```

The Debian package recommends Graphviz without requiring it, so name it in the install line when you draw anything but sequence diagrams. Debian stable ships a 2020 release of PlantUML, which can predate the switches in [rendering.md](rendering.md): read `plantuml --version` and the renderer's help, and use the container route when the switches are unknown.

Render from the `diagrams/` directory:

```bash
plantuml --svg -failfast2 -o "$PWD" reminders-lead.puml
```

## docker

Map your user, or the images come out owned by root:

```bash
docker info
docker run --rm -u "$(id -u):$(id -g)" -v "$PWD:/data" plantuml/plantuml --svg -failfast2 -o /data /data/reminders-lead.puml
```

When `docker info` reports a permission error, your user is not in the docker group or the daemon is stopped. Use podman rather than changing the group.

## podman

`--userns=keep-id` maps your user and group to the same ids inside the container, so the images are yours:

```bash
podman run --rm --userns=keep-id -v "$PWD:/data" plantuml/plantuml --svg -failfast2 -o /data /data/reminders-lead.puml
```

On a host with SELinux, add the `:Z` suffix to the volume, `-v "$PWD:/data:Z"`. It relabels the mounted content with a private label for this container, so mount the `diagrams/` directory and nothing wider.

## Structurizr

```bash
docker run --rm -u "$(id -u):$(id -g)" -v "$PWD:/usr/local/structurizr" structurizr/structurizr export -workspace /usr/local/structurizr/reminders-model.dsl -format plantuml -output /usr/local/structurizr
```

With podman use the same arguments with `--userns=keep-id` instead of `-u`.
