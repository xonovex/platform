# Claude Code Adapter Capability Matrix

## Matrix identity

| Field                  | Value                                                          |
| ---------------------- | -------------------------------------------------------------- |
| Matrix version         | `1.1.0`                                                        |
| Documentation snapshot | `2026-07-19`                                                   |
| Runtime probe          | `claude --version`                                             |
| Observed runtime       | `2.1.211 (Claude Code)`, probed `2026-07-19`                   |
| Evidence status        | `PreToolUse` command registration runtime-verified             |
| Refresh trigger        | Product update, hook schema change, handler change, or 90 days |

The project-scoped `Edit|Write` `PreToolUse` command hook was registered in a real `2.1.211` session on `2026-07-19`. The harness blocked `secrets/probe.key`, allowed `src/probe.txt`, and the shared decision service recorded both verdicts plus the paired minimized enforcement telemetry. Only that row is runtime-verified; every other hook-level row remains documentation-verified. After reviewing the probe's network and bypass-permissions behavior, re-run `scripts/refresh-pre-tool-use-probe.sh --confirm-dangerous-probe` whenever the installed version or hook contract changes.

## Native handlers and configuration

Handler types are `command`, `http`, `mcp_tool`, `prompt`, and `agent`, but availability and output fields are event-specific. Settings scopes are managed, user, shared project, local project, plugin, skill/agent component, and session/SDK registration. Managed restrictions can allow only managed and force-enabled vetted plugin hooks.

| Semantic intent          | Native mapping                                   | Support                                            | Blocking/context                                   | Ordering and limits                                               |
| ------------------------ | ------------------------------------------------ | -------------------------------------------------- | -------------------------------------------------- | ----------------------------------------------------------------- |
| Session start/resume     | `SessionStart`                                   | Supported command/MCP-tool                         | Context injection; not a general gate              | Matching handlers parallel                                        |
| Prompt before submission | `UserPromptSubmit`                               | Supported                                          | Exit `2` rejects; output may add context           | Matcher is not used                                               |
| Tool before use          | `PreToolUse`                                     | Runtime-verified command handler on Edit and Write | Exit `2` blocked the denied write before execution | Matching handlers parallel; deny does not roll back siblings      |
| Permission request       | `PermissionRequest`                              | Supported                                          | May allow, deny, ask, or defer within native rules | Native decision precedence applies                                |
| Tool after use/failure   | `PostToolUse` / `PostToolUseFailure`             | Supported                                          | Advisory/context after action                      | Cannot prevent completed action                                   |
| Context compaction       | `PreCompact` / session start with compact source | Supported                                          | Context preservation; event-specific blocking      | Treat injected text as model context                              |
| Subagent lifecycle       | Native subagent events                           | Supported                                          | Event-specific context and control                 | Separate from `agent` handler type                                |
| Model evaluator          | `prompt` handler on documented events            | Supported on selected events                       | Bounded model decision                             | Consumes model capacity; validate output                          |
| Agent verifier           | `agent` handler on documented events             | Experimental                                       | Event-specific                                     | Cannot satisfy mandatory profile                                  |
| HTTP/MCP integration     | `http` / `mcp_tool` on documented events         | Supported                                          | Event-specific                                     | URL/env allowlists and connected-server state constrain execution |

## Guarantee boundary

A hook is not an independent sandbox. Confirm the exact event, handler support, covered tools, output schema, exit behavior, permission mode, timeout, parallel sibling behavior, and active settings sources. Use independent enforcement for operations the native event cannot cover.
