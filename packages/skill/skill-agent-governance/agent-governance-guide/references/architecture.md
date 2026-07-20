# Architecture

The workflow kernel has three input dimensions and three optional cross-cutting dimensions:

| Dimension | Selection                                                         | Kernel assumption                                     |
| --------- | ----------------------------------------------------------------- | ----------------------------------------------------- |
| Trigger   | Open `kind`, native reference, optional actor and data            | None about source or trust mechanism                  |
| Executor  | One registered executor plugin                                    | Returns a neutral outcome and references              |
| Host      | Process, agent harness, CI runner, Kubernetes, or another adapter | Outside the kernel                                    |
| Controls  | Zero or more registered control plugins                           | Selection chooses `observe` or `enforce`              |
| Evidence  | Zero or more registered sink plugins                              | Selection chooses `ignore` or `fail` on sink error    |
| Maturity  | Optional caller-owned capability model                            | Assessed after composition, never an execution switch |

The composition root resolves plugin names once and passes concrete executor, control, and evidence ports into execution. The kernel does not import command, LLM, agent, hook, CI, or Kubernetes implementations.

Control phases wrap execution once: before controls run in selection order, the executor runs at most once, then after controls run in selection order. An enforcing denial stops the remaining phase. Observing denials are retained as observations.

Capabilities are inert strings declared by plugins. A composition can name required capabilities; resolution fails before controls or execution when any requirement is absent. External tools can use the same declarations for maturity or inventory without changing runtime semantics.
