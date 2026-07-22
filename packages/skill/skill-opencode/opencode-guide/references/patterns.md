## Governance ownership

Load **agent-governance-guide** to select the adoption mode and define executor, control, evidence, and enforcement semantics. This reference owns only the OpenCode mapping and product-specific caveats.

## OpenCode mappings

| Pattern                               | Native mapping                                                                | Caveat                                                           |
| ------------------------------------- | ----------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| Protected path / secret / tool policy | `tool.execute.before` plugin                                                  | Validate tool name and arguments; test failure behavior          |
| Formatting                            | `tool.execute.after` or `file.edited`                                         | Keep transformations idempotent under repeated events            |
| Validation                            | Session-idle or command integration                                           | Observation alone does not enforce stop                          |
| Audit                                 | Event subscriber plus structured app logging/provider                         | Minimize message, file, prompt, and tool content                 |
| Context injection                     | Supported message/session hooks; experimental compaction only when acceptable | Mark experimental support visibly                                |
| Custom capability                     | Plugin custom tool                                                            | Closed argument schema, unique name, cancellation, output bounds |
| Model evaluator                       | Explicit plugin runner                                                        | Validate closed result and declare provider/data transfer        |
| Specialist agent                      | Explicit bounded child process/plugin                                         | Depth, process group, authority, cwd, budgets, and schema        |
| Organization-managed                  | Pinned npm/local plugin delivered through external managed configuration      | Native organization precedence is not documented here            |

Because plugin code receives a shell API and SDK client, least privilege is a module design and external process boundary, not an automatic platform property.
