# Shared Launch Go

Use this shared Go library to plan one confined agent run: choose the isolation and the provisioning from a registry, check the requested guarantees, resolve the credentials, and build the portable run spec that an executor carries out.

The library holds the ports and the planning core only. The isolators, provisioners, credential stores and the egress proxy that touch a machine are adapters of each consumer, which registers them in its composition root.

## Packages

| Package | Description |
| --- | --- |
| `pkg/isolation` | The isolation ports: `Confinement` declares a sandbox and its guarantees, `Launcher` starts it on one machine, `Isolator` is both; `RunConfig`, `Process` and the `EgressRoute` of a run |
| `pkg/provision` | The provisioning port and its `Input` |
| `pkg/sandbox` | The `Registry` of lazy isolator and provisioner factories, `Evaluate` and `Select` of an isolation, and `EnforcePolicy` over the declared capabilities |
| `pkg/credential` | `Secret`, `Binding`, the `Store` port and `Resolve` of the credentials of a run |
| `pkg/launch` | The portable `Spec` of a run and the planner that turns a `Request` into it |
| `pkg/record` | The run record, its JSON Schema and its store over a file-system port |

No package reads the disk, the environment, the clock or the network, and no package names a concrete isolation, provisioning or harness. `internal/arch` enforces this, keeps the dependency direction inside the module, keeps the factories lazy, keeps a credential value inside `Secret`, and keeps the agent command line out of the isolation port.
