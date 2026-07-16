# Kiro Capability Matrix

## Matrix identity

| Field                  | Value                                       |
| ---------------------- | ------------------------------------------- |
| Matrix version         | `1.0.0`                                     |
| Documentation snapshot | `2026-07-16`                                |
| Runtime probe          | `kiro-cli --version`                        |
| Observed runtime       | Not installed in the validation environment |
| Hook schema            | `v1` standalone workspace JSON              |
| Evidence status        | Documentation-verified; runtime-unverified  |

## Native capabilities

Kiro v1 hooks use files under `.kiro/hooks/` and support command or agent actions. Command success can add context; event-specific exit `2` blocks. Agent actions run an agent loop, and prompt-submit agent actions append to the user prompt.

| Semantic intent             | Native mapping                                     | Support                 | Blocking/context                           | Limits                                        |
| --------------------------- | -------------------------------------------------- | ----------------------- | ------------------------------------------ | --------------------------------------------- |
| Session start               | `SessionStart`                                     | Supported               | Context only                               | No blocking                                   |
| Prompt before submission    | `UserPromptSubmit`                                 | Supported command/agent | Command exit `2` blocks; context supported | Prompt content exposed to action              |
| Session/turn completion     | `Stop`                                             | Supported               | Advisory                                   | No blocking in v1 table                       |
| Tool before use             | `PreToolUse`                                       | Supported command/agent | Command exit `2` blocks                    | Match tool names/categories precisely         |
| Tool after use              | `PostToolUse`                                      | Supported command/agent | Advisory/context                           | Action already occurred                       |
| Workspace file mutation     | `PostFileCreate`, `PostFileSave`, `PostFileDelete` | Supported               | Post-event                                 | Agent-originated file events in defining root |
| Capability before execution | `PreTaskExec`                                      | Supported               | Command exit `2` blocks                    | Native spec-task scope only                   |
| Capability after result     | `PostTaskExec`                                     | Supported               | Advisory/context                           | Native spec-task scope only                   |
| Model/agent evaluation      | Agent action                                       | Supported               | Event-specific                             | Consumes credits and starts agent loop        |

Ordering and concurrency guarantees are unknown in this documentation snapshot. A profile requiring serial handlers fails until a runtime probe establishes it.
