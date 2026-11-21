# Sandbox diagrams

Read how each isolation of the agent sandbox confines an agent, and how the operator runs one as a pod.

The [overview](00-overview.png) compares every isolation in one picture: the layer that confines the agent, the kernel it reaches, and what the isolation does with the network modes `none` and `managed`. A numbered diagram below shows one isolation in detail. Each claim in a diagram comes from the isolator package of `shared-launch-go` (`pkg/isolation`), from the provisioner (`pkg/provision/nix`), or from the Kubernetes isolation of the operator.

| Diagram | Image | Source | What it shows |
| --- | --- | --- | --- |
| Overview | [PNG](00-overview.png) | [Graphviz](00-overview.dot) | The isolation axis, the network modes per isolation, the four guarantees and the threat model |
| `none` | [PNG](01-none.png) | [Graphviz](01-none.dot) | Direct execution on the host |
| `bwrap` | [PNG](02-bubblewrap.png) | [Graphviz](02-bubblewrap.dot) | Linux namespaces, deny-default mounts and an environment allowlist |
| `docker` | [PNG](03-docker.png) | [Graphviz](03-docker.dot) | A container with dropped capabilities, a read-only root and resource limits |
| `nix` provision | [PNG](04-nix.png) | [Graphviz](04-nix.dot) | The pinned closure that a sandbox mounts read-only |
| gVisor | [PNG](05-gvisor.png) | [Graphviz](05-gvisor.dot) | The user-space kernel, through `docker --runtime runsc` or a pod runtime class |
| Kata Containers | [PNG](06-kata.png) | [Graphviz](06-kata.dot) | A lightweight virtual machine, through a pod runtime class |
| Confidential Containers | [PNG](07-coco.png) | [Graphviz](07-coco.dot) | A trusted execution environment, through a pod runtime class |
| Operator architecture | [PNG](08-operator-architecture.png) | [Graphviz](08-operator-architecture.dot) | Resources, webhooks, reconcilers and phase lifecycles |
| `seatbelt` | [PNG](09-seatbelt.png) | [Graphviz](09-seatbelt.dot) | A macOS profile on a host process |
| `docker-sandboxes` | [PNG](10-docker-sandboxes.png) | [Graphviz](10-docker-sandboxes.dot) | An `sbx` sandbox with its own agent image |
| `openshell` | [PNG](11-openshell.png) | [Graphviz](11-openshell.dot) | A gateway sandbox under a policy |

## Read the diagrams

The sandbox has three independent axes: the isolation, the provision (`none`, `nix`, `command`) and the network (`host`, `none`, `managed`). The isolation diagrams show the first axis, and the nix diagram shows one provision. The [agent CLI guide](../../tooling/cli/agent-cli-go/README.md#isolations) lists what each isolation declares, and the [operator guide](../../sandbox/operator/agent-operator-go/README.md) describes the pod runtime classes.

## Render changes

Edit the `.dot` source, run the render task from the repository root in the Nix environment, and commit the source with its PNG. A source names a role and sets no colour and no font; `diagram-role-check` fails on either and on an unknown role.

```bash
nix develop --no-update-lock-file --command npx moon run asset-diagrams:graph-build
```
