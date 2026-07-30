# OpenCode Capability Matrix

## Matrix identity

| Field                  | Value                                      |
| ---------------------- | ------------------------------------------ |
| Matrix version         | `1.0.0`                                    |
| Documentation snapshot | `2026-07-16`                               |
| Runtime probe          | `opencode --version`                       |
| Observed runtime       | `1.14.30`, probed `2026-07-16`             |
| Evidence status        | Documentation-verified; runtime-unverified |
| Refresh trigger        | OpenCode/plugin API release or 90 days     |

The probe observed the installed CLI version on a host with a working credentialed install (`opencode auth list` reports three stored credentials), and the walking skeleton exercised the deterministic guard contract (JSON event on stdin, exit 0 allow / exit 2 deny) locally on the same host. Native plugin registration was not exercised: hook-level rows below remain documentation-verified and must not be reported as runtime conformance. Re-run the probe and skeleton whenever the installed version differs from the observed one.

## Native capabilities

OpenCode loads npm plugins from global/project config and local JavaScript or TypeScript plugins from global/project directories. The documented order is global config, project config, global plugin directory, then project plugin directory, with hooks running sequentially.

| Semantic intent      | Native mapping                                             | Support            | Blocking/context                                           | Limits                                             |
| -------------------- | ---------------------------------------------------------- | ------------------ | ---------------------------------------------------------- | -------------------------------------------------- |
| Tool before use      | `tool.execute.before`                                      | Supported          | Plugin may mutate arguments or fail the call               | Not an independent sandbox                         |
| Tool after use       | `tool.execute.after`                                       | Supported          | Result observation/modification                            | Action already occurred                            |
| Permission lifecycle | `permission.asked` / `permission.replied`                  | Supported events   | Observation documented; decision guarantee not established | Do not claim blocking without runtime/API proof    |
| Session lifecycle    | `session.created`, `session.idle`, `session.error`, others | Supported events   | Observation                                                | Event-specific cancellation not documented         |
| File mutation        | `file.edited` / watcher updates                            | Supported events   | Post-event                                                 | File event coverage differs from tool interception |
| Shell environment    | `shell.env`                                                | Supported mutation | Adds environment/context                                   | Applies to AI and user shell execution             |
| Command              | `command.executed`                                         | Supported event    | Observation                                                | Post-command event                                 |
| Custom capability    | Plugin `tool` registration                                 | Supported          | Tool-defined                                               | Name collision overrides built-in                  |
| Context compaction   | `experimental.session.compacting`                          | Experimental       | Injects compaction context                                 | Cannot guarantee mandatory stable control          |

The event list is not a promise that every event can block or rewrite. Record event-specific handler types and probes rather than generalizing from `tool.execute.before`.
