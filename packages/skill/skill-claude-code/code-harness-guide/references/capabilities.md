# Claude Code Adapter Capability Matrix

## Matrix identity

| Field                  | Value                                                          |
| ---------------------- | -------------------------------------------------------------- |
| Matrix version         | `1.0.0`                                                        |
| Documentation snapshot | `2026-07-16`                                                   |
| Runtime probe          | `claude --version`                                             |
| Observed runtime       | `2.1.211 (Claude Code)`, probed `2026-07-16`                   |
| Evidence status        | Runtime version observed; guard contract exercised locally     |
| Refresh trigger        | Product update, hook schema change, handler change, or 90 days |

The probe observed the installed CLI version, and the walking skeleton exercised the deterministic guard contract (JSON event on stdin, exit 0 allow / exit 2 deny) against that runtime locally. Native hook registration was not exercised: hook-level rows below remain documentation-verified and must not be reported as runtime conformance. Re-run the probe and skeleton whenever the installed version differs from the observed one.

## Native handlers and configuration

Handler types are `command`, `http`, `mcp_tool`, `prompt`, and `agent`, but availability and output fields are event-specific. Settings scopes are managed, user, shared project, local project, plugin, skill/agent component, and session/SDK registration. Managed restrictions can allow only managed and force-enabled vetted plugin hooks.

| Semantic intent          | Native mapping                                   | Support                                        | Blocking/context                                   | Ordering and limits                                               |
| ------------------------ | ------------------------------------------------ | ---------------------------------------------- | -------------------------------------------------- | ----------------------------------------------------------------- |
| Session start/resume     | `SessionStart`                                   | Supported command/MCP-tool                     | Context injection; not a general gate              | Matching handlers parallel                                        |
| Prompt before submission | `UserPromptSubmit`                               | Supported                                      | Exit `2` rejects; output may add context           | Matcher is not used                                               |
| Tool before use          | `PreToolUse`                                     | Supported, handler availability event-specific | Native decision or exit `2` can deny covered call  | Matching handlers parallel; deny does not roll back siblings      |
| Permission request       | `PermissionRequest`                              | Supported                                      | May allow, deny, ask, or defer within native rules | Native decision precedence applies                                |
| Tool after use/failure   | `PostToolUse` / `PostToolUseFailure`             | Supported                                      | Advisory/context after action                      | Cannot prevent completed action                                   |
| Context compaction       | `PreCompact` / session start with compact source | Supported                                      | Context preservation; event-specific blocking      | Treat injected text as model context                              |
| Subagent lifecycle       | Native subagent events                           | Supported                                      | Event-specific context and control                 | Separate from `agent` handler type                                |
| Model evaluator          | `prompt` handler on documented events            | Supported on selected events                   | Bounded model decision                             | Consumes model capacity; validate output                          |
| Agent verifier           | `agent` handler on documented events             | Experimental                                   | Event-specific                                     | Cannot satisfy mandatory profile                                  |
| HTTP/MCP integration     | `http` / `mcp_tool` on documented events         | Supported                                      | Event-specific                                     | URL/env allowlists and connected-server state constrain execution |

## Guarantee boundary

A hook is not an independent sandbox. Confirm the exact event, handler support, covered tools, output schema, exit behavior, permission mode, timeout, parallel sibling behavior, and active settings sources. Use independent enforcement for operations the native event cannot cover.
