## Adoption modes

Keep the semantic contract in **agent-governance-guide** and translate only the selected mode:

| Mode                 | Native result                                                | Enforcement claim                                   |
| -------------------- | ------------------------------------------------------------ | --------------------------------------------------- |
| Knowledge-only       | `AGENTS.md` or an Agent Skill                                | Guidance only                                       |
| Advisory             | Post-event observer or non-blocking handler                  | Evidence or feedback only                           |
| Enforcing            | Deterministic pre-event handler with verified coverage       | Only the documented covered operations              |
| Model evaluator      | Native supported model handler or explicit command runner    | Bounded validated inference, never silent authority |
| Specialist agent     | Explicit bounded launcher                                    | Attenuated authority, depth and budgets enforced    |
| Organization-managed | Native managed configuration plus pinned executable delivery | Only within verified native and external guarantees |

Translate the deterministic templates for protected paths, secrets, tool policy, formatting, validation, audit, context injection, and privileged operations. Record native event, handler, matcher, input schema, output/exit behavior, timeout, concurrency, permissions, data exposure, evidence, and rollback.

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
