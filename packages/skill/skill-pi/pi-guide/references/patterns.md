## Governance ownership

Load **agent-governance-guide** to select the adoption mode and define executor, control, evidence, and enforcement semantics. This reference owns only the Pi mapping and product-specific caveats.

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
