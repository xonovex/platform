# Codex Capability Matrix

## Matrix identity

| Field                  | Value                                         |
| ---------------------- | --------------------------------------------- |
| Matrix version         | `1.0.0`                                       |
| Documentation snapshot | `2026-07-22`                                  |
| Runtime probe          | `codex --version`                             |
| Observed runtime       | Not installed in the validation environment   |
| Evidence status        | Documentation-verified; runtime-unverified    |
| Refresh trigger        | Codex release, hook schema change, or 90 days |

## Configuration and execution

Codex loads `hooks.json` or inline `hooks` tables next to active configuration layers, plus plugin-bundled hooks. Useful native scopes include managed/system, user, project, session, and plugin. Multiple sources combine rather than replacing lower layers. Project hooks require project trust; non-managed command definitions require review of the current hash.

| Semantic intent          | Native mapping               | Support                             | Blocking/context                                     | Limits                                                                                                                  |
| ------------------------ | ---------------------------- | ----------------------------------- | ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Session start/resume     | `SessionStart`               | Supported command                   | Adds developer context                               | Matcher covers startup/resume/clear/compact                                                                             |
| Prompt before submission | `UserPromptSubmit`           | Supported command                   | Event-specific stop/context                          | Matcher ignored                                                                                                         |
| Tool before use          | `PreToolUse`                 | Supported command, partial coverage | Deny or exit `2` for covered call; context supported | Shell/unified exec, patch/edit/write, MCP, and most local function tools; hosted and specialized opt-out paths excluded |
| Permission request       | `PermissionRequest`          | Supported command, partial coverage | Allow, deny, or decline for native approval request  | Does not fire for operations requiring no approval                                                                      |
| Tool after use           | `PostToolUse`                | Supported command                   | Can stop turn or add warning/context                 | Action already occurred                                                                                                 |
| Compaction               | `PreCompact` / `PostCompact` | Supported command                   | Context and event-specific stop                      | Trigger matcher manual/auto                                                                                             |
| Subagent before launch   | `SubagentStart`              | Supported command                   | Adds child context; cannot stop launch               | Agent type matcher                                                                                                      |
| Subagent completion      | `SubagentStop`               | Supported command                   | Event-specific                                       | Matching commands concurrent                                                                                            |
| Model evaluator handler  | `type: prompt`               | Unsupported execution               | None                                                 | Parsed then skipped                                                                                                     |
| Agent handler            | `type: agent`                | Unsupported execution               | None                                                 | Parsed then skipped                                                                                                     |
| Async command            | `async: true`                | Unsupported execution               | None                                                 | Parsed then skipped                                                                                                     |

## Guardrail boundary

A Codex hook cannot claim complete mandatory coverage unless the requested operation set is restricted to verified intercepted paths. `write_stdin` does not emit a second `PreToolUse` event for a command that already passed the hook. Managed requirements can restrict hook sources and pin features, but do not distribute scripts from the managed directory.
