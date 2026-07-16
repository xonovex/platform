# Harness Module Patterns

## Adoption ladder

Choose the lowest-authority composition that meets the requirement:

| Mode                        | Mechanism                                                 | Guarantee                                                               |
| --------------------------- | --------------------------------------------------------- | ----------------------------------------------------------------------- |
| Knowledge-only              | `AGENTS.md` or an Agent Skill                             | Guidance and reusable procedure only                                    |
| Advisory hook               | Deterministic post-event observer                         | Visible evidence or feedback; never blocks                              |
| Enforcing hook              | Deterministic native pre-event handler                    | Blocks only the documented covered operation set                        |
| Script plus model evaluator | Deterministic runner invoking a bounded model             | Validated judgment; failure is visible and never silently authoritative |
| Specialist agent launcher   | Explicit bounded child process or native agent hook       | Adaptive work with attenuated authority and no hidden recursion         |
| Organization-managed        | Native managed policy plus pinned executable distribution | Mandatory only within the verified native and external guarantees       |

Lifecycle commands are optional consumers. Every pattern works for ordinary agent activity and refers to native files, modules, and diagnostics rather than assuming workflow commands are installed.

## Deterministic templates

Translate the semantic templates in `assets/harness-module-templates.json` into the selected harness's native mechanism:

- protected-path interception;
- secret-pattern and secret-file protection;
- tool allow/deny policy;
- post-write formatting;
- validation before stop or publication;
- minimized audit events;
- bounded context injection;
- privileged-operation authorization interception.

Each translated template declares inputs, outputs, covered native tools/events, timeout, side effects, concurrency, failure behavior, evidence, permissions, and rollback. Path or command matching is input validation, not a sandbox. Audit templates default to metadata and hashes; capturing prompt, file, tool output, or transcript content requires explicit authorization and retention rules.

## Bounded model evaluator

Use a native prompt/model handler only when the matrix states that it executes for the selected event. Otherwise use an explicit deterministic command runner. Require:

- fixed purpose and selected input references;
- pinned model/provider policy;
- time, token, cost, retry, and concurrency budgets;
- no filesystem, tool, network, secret, or transcript access beyond the declaration;
- a closed output schema such as `{outcome, reasons, evidenceRequests}`;
- strict parsing and validation before any decision is consumed;
- fail-visible behavior on timeout, provider error, invalid JSON, unknown fields, or budget exhaustion.

Model output is inference evidence. It does not replace authoritative inspection or directly expand authority.

## Bounded agent launcher

Require an explicit launch action and declare purpose, working directory, maximum depth, model/provider, time/token/cost budgets, tool/filesystem/network/secret/data scopes, result schema, cancellation, kill switch, and evidence. Child authority is a strict subset of the parent. Depth `0` means no child launch; depth `1` means a child cannot launch another child.

Native agent hooks remain optional and experimental where the matrix says so. A command-launched child must use an explicit process runner, pass task text without exposing secrets in process arguments where practical, validate its structured result, and terminate the whole process group on cancellation or timeout.

## Organization-managed composition

Managed configuration and executable distribution are separate concerns. Pin the configuration and the executable source/version/digest, verify both, record ownership and change control, and test that lower scopes cannot disable or bypass the selected control. If the native product only manages configuration but does not distribute scripts, use an organization-controlled package or device-management channel and retain both references.
