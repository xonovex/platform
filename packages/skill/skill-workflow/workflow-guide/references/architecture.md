# Architecture

## Ownership boundaries

| Owner               | Owns                                                                         | Does not own                                                  |
| ------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------- |
| Workflow contracts  | Result meaning, topology, publication boundaries, exit/completion evaluation | Trigger, executable, host, controls, evidence sinks, maturity |
| Provider adapter    | Native reads, writes, references, revisions, authentication                  | Universal workflow identity or lifecycle meaning              |
| Runtime composition | Selected executor, controls, evidence behavior, capability gate              | Provider representation or lifecycle topology                 |
| Trigger adapter     | Native event normalization and trusted-template binding                      | Executable definitions or implicit controls                   |
| Host adapter        | Process/container isolation, resources, secrets, host policy                 | Workflow maturity or global governance                        |

Workflow contracts and executable composition are peer capabilities. Either can be used
without the other. When they are combined, pass the workflow operation and opaque native
references into a normal runtime invocation; do not add a cross-plane profile.

## Workflow variation

Resolve workflow topology, method, artifact provider, work-item provider, code host, and
workspace provider independently. An unavailable explicit selection fails visibly. A
side-effecting operation never silently falls back to local storage.

Executor, trigger, host, controls, evidence sinks, and maturity remain outside these
workflow selections. See `agent-governance-guide` only when an executable composition is
being built or explained.
