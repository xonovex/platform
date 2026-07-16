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

Record native event, handler, matcher, input schema, output/exit behavior, timeout, concurrency, permissions, data exposure, evidence, and rollback for each translated template.

## Pi mappings

| Pattern                               | Native mapping                                                          | Caveat                                                                                   |
| ------------------------------------- | ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Protected path / secret / tool policy | `tool_call` extension handler                                           | Validate mutable input; no automatic revalidation                                        |
| Formatting                            | `tool_result` or post-event handler                                     | Sibling completion order may interleave                                                  |
| Validation                            | `agent_settled`, session, or command integration                        | Distinguish low-level end from fully settled                                             |
| Audit                                 | Extension event to approved provider                                    | Full context may be sensitive; minimize before transfer                                  |
| Context injection                     | `before_agent_start`, `context`, or custom message                      | State whether content persists in session/model context                                  |
| Custom capability                     | `registerTool`                                                          | Closed schema, output bounds, cancellation, and full process authority                   |
| Model evaluator                       | Explicit extension runner                                               | Pin provider policy and validate output                                                  |
| Specialist agent                      | Explicit child `pi` process/extension                                   | Depth 1, trusted cwd, no recursion, attenuated tools/extensions, terminate process group |
| Organization-managed                  | Pinned package/settings distributed through an external managed channel | Pi project trust is not organization provenance                                          |

Permission extensions and subagent packages are optional modules, not built-in guarantees. Inspect their source and declare their actual authority rather than naming them as a platform property.
