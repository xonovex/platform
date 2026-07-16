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

## Codex mappings

| Pattern                               | Native mapping                                            | Caveat                                                 |
| ------------------------------------- | --------------------------------------------------------- | ------------------------------------------------------ |
| Protected path / secret / tool policy | `PreToolUse` command                                      | Only guarantees verified intercepted tool paths        |
| Permission interception               | `PermissionRequest` command                               | Only fires when native approval is requested           |
| Formatting                            | `PostToolUse` command for covered writes                  | Idempotent under concurrent matching commands          |
| Validation before completion          | `Stop` command                                            | Bound continuation and parse one closed result         |
| Audit                                 | Post-event command                                        | Minimize transcript and tool payload capture           |
| Context injection                     | `SessionStart`, `SubagentStart`, or supported hook output | Context is not authoritative evidence                  |
| Model evaluator                       | Explicit deterministic command runner                     | Native prompt handlers do not execute in this snapshot |
| Specialist agent                      | Explicit bounded command runner                           | Native agent handlers do not execute in this snapshot  |
| Organization-managed                  | `requirements.toml` hooks plus separately managed scripts | Use absolute paths and verify both artifacts           |

A command runner invoking a model or child agent declares that executor class in its module metadata. Do not label it deterministic merely because the outer handler is a command.
