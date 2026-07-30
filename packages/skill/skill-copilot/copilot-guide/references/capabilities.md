# GitHub Copilot Capability Matrix

## Matrix identity

| Field                  | Value                                        |
| ---------------------- | -------------------------------------------- |
| Matrix version         | `1.0.0`                                      |
| Documentation snapshot | `2026-07-16`                                 |
| Runtime probe          | `copilot --version`                          |
| Observed runtime       | `0.0.377 (Copilot CLI)`, probed `2026-07-16` |
| Hook schema            | JSON version `1`                             |
| Surfaces               | Copilot CLI and cloud agent                  |
| Evidence status        | Documentation-verified; runtime-unverified   |

The probe observed the installed CLI version on a host with a working credentialed install (local CLI auth/session state present; the GitHub CLI on the same host reports a signed-in account), and the walking skeleton exercised the deterministic guard contract (JSON event on stdin, exit 0 allow / exit 2 deny) locally on the same host. The cloud-agent surface and native hook registration were not exercised: hook-level rows below remain documentation-verified and must not be reported as runtime conformance. Re-run the probe and skeleton whenever the installed version differs from the observed one.

## Handler correction

The current hook reference is not command-only. Command handlers are supported across hook types; HTTP handlers are documented with TLS and environment allowlist rules; prompt handlers are limited to `sessionStart` and interactive CLI startup behavior.

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
