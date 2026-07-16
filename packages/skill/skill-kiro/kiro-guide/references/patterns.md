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

Translate the deterministic templates for protected paths, secrets, tool policy, formatting, validation, audit, context injection, and privileged operations.

## Kiro mappings

| Pattern                               | Native mapping                                                                  | Caveat                                                     |
| ------------------------------------- | ------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| Protected path / secret / tool policy | `PreToolUse` command action                                                     | Exit `2` is the documented block; validate tool categories |
| Prompt policy                         | `UserPromptSubmit` command action                                               | Prompt is sensitive input                                  |
| Formatting                            | `PostFileSave` or `PostToolUse` command                                         | Keep file pattern narrow and action idempotent             |
| Validation before spec task           | `PreTaskExec` command                                                           | Covers native spec-task execution only                     |
| Audit                                 | Post-event command                                                              | Minimize prompt/tool/file content                          |
| Context injection                     | Successful command stdout on documented events                                  | Treat as model context                                     |
| Model evaluator                       | Agent action or explicit command runner                                         | Bound credits, purpose, output schema, and failure         |
| Specialist agent                      | Explicit bounded command runner                                                 | No hidden recursion or ambient authority                   |
| Organization-managed                  | Version-controlled workspace hook plus independent repository/endpoint controls | Native managed precedence is not documented here           |

A shared repository hook is executable project content. Establish repository trust and review before enabling it for contributors.
