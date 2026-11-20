# Shared Launch Go

Use this shared Go library to plan and start one confined agent run: choose the isolation and the provisioning from a registry, check the requested guarantees, resolve the credentials, build the portable run spec, and confine the agent process on one machine.

The library holds a core of ports and planning, and the leaves that implement the ports. A consumer builds its own registry from the leaves it wants and hands each leaf the operating-system adapter in its composition root. A consumer that runs the agent elsewhere, such as a Kubernetes Job, imports only the core.

## Core packages

| Package | Description |
| --- | --- |
| `pkg/isolation` | The isolation ports: `Confinement` declares a sandbox, its guarantees and whether it starts from a container image, which the platform never supplies (`ErrNoImage`), `Launcher` starts it on one machine, `Isolator` is both; `RunConfig`, `Process`, the `EgressRoute` of a run and the bridge and proxy environment of that route |
| `pkg/provision` | The provisioning port and its `Input` |
| `pkg/credential` | `Secret`, `Binding`, the `Store` port and `Resolve` of the credentials of a run |
| `pkg/sandbox` | The `Registry` of lazy isolator and provisioner factories, `Evaluate` and `Select` of an isolation, and `EnforcePolicy` over the declared capabilities |
| `pkg/launch` | The portable `Spec` of a run and the planner that turns a `Request` into it |
| `pkg/record` | The run record, its JSON Schema and its store over a file-system port |
| `pkg/hostlaunch` | What the isolation leaves share to start the agent on one machine: the `Host` port, the resolution of the host paths a sandbox binds, and the environment and init commands of the agent process |

No core package reads the disk, the environment, the clock or the network, and no core package names a concrete isolation, provisioning or harness.

## Leaves

| Package | Description |
| --- | --- |
| `pkg/isolation/{none,bwrap,docker,seatbelt,dockersandboxes,openshell}` | One isolation each, behind `isolation.Isolator` |
| `pkg/provision/{none,command,nix}` | One provisioning each, behind `provision.Provisioner` |
| `pkg/credential/{env,file,keychain,pass,secretservice}` | One credential store each, behind `credential.Store` |
| `pkg/egress/managed` | The allowlist proxy of a managed run and the bridge that relays the loopback port of a sandbox to it |
| `pkg/oshost` | The adapter over the operating system: files, environment, processes, signals, network, clock and user identity |
| `pkg/oshost/oshosttest` | `Fake`, the in-memory double of `oshost` for unit tests |

Each leaf reaches the machine through a port that `oshost.System` satisfies, so its unit tests run against `oshosttest.Fake`. Only `pkg/oshost` imports `os`, `os/exec`, `os/signal`, `syscall` or `runtime`, and the integration tier (`npx moon run shared-launch-go:ci-integration`) covers it.

`internal/arch` enforces the split: the dependency direction inside the module, a core that uses only the standard library and the shared libraries and names no network, no leaf that imports another leaf, lazy factories, a credential value that stays inside `Secret`, and an agent command line that stays out of the isolation port and its leaves.
