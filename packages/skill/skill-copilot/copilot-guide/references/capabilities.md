# GitHub Copilot Capability Matrix

## Matrix identity

| Field                  | Value                                       |
| ---------------------- | ------------------------------------------- |
| Matrix version         | `1.0.0`                                     |
| Documentation snapshot | `2026-07-16`                                |
| Runtime probe          | `copilot --version`                         |
| Observed runtime       | Not installed in the validation environment |
| Hook schema            | JSON version `1`                            |
| Surfaces               | Copilot CLI and cloud agent                 |
| Evidence status        | Documentation-verified; runtime-unverified  |

## Handler correction

The current hook reference is not command-only. Command handlers are supported across hook types; HTTP handlers are documented with TLS and environment allowlist rules; prompt handlers are limited to `sessionStart` and interactive CLI startup behavior. Preserve this correction when revisiting older plans.

| Semantic intent       | Native mapping                | CLI                          | Cloud agent             | Guarantee notes                                    |
| --------------------- | ----------------------------- | ---------------------------- | ----------------------- | -------------------------------------------------- |
| Session start/end     | `sessionStart` / `sessionEnd` | Supported                    | Surface-specific subset | Prompt handler only on new interactive CLI session |
| Prompt submitted      | `userPromptSubmitted`         | Supported                    | Supported subset        | Observation/context; use event-specific output     |
| Tool before use       | `preToolUse`                  | Supported blocking           | Supported blocking      | Cloud uses bash/command and repository file        |
| Permission request    | `permissionRequest`           | Supported decision           | Unsupported/no effect   | Cloud tools are pre-approved; use `preToolUse`     |
| Tool after use        | `postToolUse`                 | Supported                    | Supported subset        | Post-event result/context                          |
| Main agent completion | `agentStop`                   | Supported continuation block | Supported               | Block forces another turn and consumes budget      |
| Subagent completion   | `subagentStop`                | Supported                    | Surface-specific        | Event-specific continuation                        |
| Error                 | `errorOccurred`               | Supported                    | Supported               | Observation, no output decision                    |
| Notification          | `notification`                | Async, non-blocking          | Does not fire           | Fire-and-forget                                    |
| HTTP integration      | HTTPS handler                 | Supported                    | Firewall-constrained    | Permission decisions require TLS                   |

CLI configuration combines policy, user, project, inline settings, and plugin hooks. Cloud loads repository `.github/hooks/*.json` by default. Policy hooks are CLI-only, load before other sources, and cannot be disabled by lower scopes.
