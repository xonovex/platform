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

## GitHub Copilot mappings

| Pattern                               | Native mapping                                                        | Caveat                                               |
| ------------------------------------- | --------------------------------------------------------------------- | ---------------------------------------------------- |
| Protected path / secret / tool policy | `preToolUse` command or HTTPS handler                                 | Cloud and CLI inputs/environments differ             |
| Permission interception               | CLI `permissionRequest`                                               | Use `preToolUse` in cloud                            |
| Formatting                            | `postToolUse` command                                                 | Cloud filesystem is ephemeral                        |
| Validation before completion          | `agentStop` command/HTTP                                              | Bound continuation within job/session timeout        |
| Audit                                 | Post-event command/HTTP                                               | Cloud retention requires explicit external transfer  |
| Context injection                     | Supported event output or interactive `sessionStart` prompt           | Prompt handler does not cover resume/non-interactive |
| Model evaluator                       | Explicit command or HTTPS runner                                      | Validate schema and disclose content transfer        |
| Specialist agent                      | Explicit bounded runner or reviewed custom agent/subagent composition | Enforce depth, tools, and budget outside prose       |
| Organization-managed                  | CLI policy hooks plus managed plugin/skill distribution               | Policy hooks do not apply to cloud                   |

Repository skills and plugins are packaging mechanisms. Installation or discovery is not proof that a blocking hook ran.
