# Claude Code Adoption Patterns

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

## Claude Code mappings

| Pattern                               | Native mapping                                                                | Caveat                                                                         |
| ------------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Protected path / secret / tool policy | Deterministic `PreToolUse` command                                            | Match the actual tool/input and return the documented denial shape or exit `2` |
| Formatting                            | `PostToolUse` command for covered write tools                                 | Idempotent because matching handlers run in parallel                           |
| Validation before completion          | `Stop` command                                                                | Bound continuation count and avoid recursive stop loops                        |
| Audit                                 | Post-event command, HTTP, or MCP-tool                                         | Minimize transcript, prompt, and file content                                  |
| Context injection                     | `SessionStart` or documented context-producing output                         | Treat as model-visible context, not evidence                                   |
| Model evaluator                       | `prompt` handler only on a supported event, otherwise explicit command runner | Validate a closed output schema                                                |
| Specialist agent                      | Experimental `agent` handler or explicit bounded runner                       | Native agent handler remains optional and experimental                         |
| Organization-managed                  | Managed hooks plus pinned device/package distribution                         | Test startup and lower-scope non-weakening behavior                            |

Do not hide a model evaluator inside a deterministic template. The preview must state when model input leaves the process, which content is sent, and what happens on timeout or invalid output.
