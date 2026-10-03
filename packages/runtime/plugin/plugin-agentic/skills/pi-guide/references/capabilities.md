# Pi Capability Matrix

## Matrix identity

| Field                  | Value                                      |
| ---------------------- | ------------------------------------------ |
| Matrix version         | `1.0.0`                                    |
| Documentation snapshot | `2026-07-16`                               |
| Runtime probe          | `pi --version`                             |
| Observed runtime       | `0.80.2`, probed `2026-07-16`              |
| Documentation channel  | `latest`                                   |
| Evidence status        | Documentation-verified; runtime-unverified |

The probe observed the installed CLI version on a host with a working credentialed install (the agent auth store `~/.pi/agent/auth.json` is present), and the walking skeleton exercised the deterministic guard contract (JSON event on stdin, exit 0 allow / exit 2 deny) locally on the same host. Native extension registration was not exercised: extension-level rows below remain documentation-verified and must not be reported as runtime conformance. Re-run the probe and skeleton whenever the installed version differs from the observed one.

## Native capabilities

Pi extensions are TypeScript modules with lifecycle events, custom tools, commands, session state, and provider access; packages bundle extensions, skills, templates, and themes. Project-local resources load only after project trust, and none of this is a built-in sandbox.

| Semantic intent                | Native mapping                               | Support                         | Blocking/context                          | Ordering and limits                                                 |
| ------------------------------ | -------------------------------------------- | ------------------------------- | ----------------------------------------- | ------------------------------------------------------------------- |
| Workspace before activation    | `project_trust`                              | Supported                       | First yes/no decision owns trust          | User/global or CLI extensions only before trust                     |
| Session start/end              | `session_start` / shutdown events            | Supported                       | Observation/context                       | Session replacement has lifecycle caveats                           |
| Prompt/input before processing | `input`, `before_agent_start`                | Supported                       | Transform/handle or inject prompt/context | Skill/template expansion order is explicit                          |
| Model before call              | `context` and provider-request events        | Supported                       | Can modify model context/request          | Sensitive full context may be visible                               |
| Tool before use                | `tool_call`                                  | Supported blocking              | Mutate input or return block              | Preflight sequential; sibling execution concurrent                  |
| Tool after use                 | `tool_result`                                | Supported                       | Modify result/context                     | Middleware chain; completion order may interleave                   |
| Context compaction             | `session_before_compact` / `session_compact` | Supported                       | Customize or cancel before compaction     | Extension chain                                                     |
| User shell                     | `user_bash`                                  | Supported interception          | Replace operations or result              | Full process authority                                              |
| Custom tools                   | `registerTool`                               | Supported                       | Tool-defined                              | Must validate inputs and truncate output                            |
| Specialist child agent         | Extension-defined process/package            | No universal built-in guarantee | Extension-defined                         | Launcher owns trust, recursion, isolation, permissions, and budgets |

The exact extension API is tied to the installed release. Re-probe event names, ordering, and handler types before runtime claims.
