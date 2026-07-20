# Autonomy Levels and Oversight Coupling

## Who this ladder is for

This ladder is guidance for a team adopting these contracts. It grades an adopter's own environment and is not a claim about the posture of the environment that publishes it; adopting these contracts confers no level. A level is something an adopter asserts about an environment and backs with the oversight evidence below.

**Start at `A1`.** It is the recommended starting point for an adopting team: it needs only an interactive session and an independent critique, so it is reachable without a headless runner, a run journal, or admission control, and it establishes the adversarial reading that every higher level assumes.

The workflow runtime can launch a workflow-skill agent at `A1`, `A2`, or `A3` only after a registered verifier returns observed evidence for that exact level. The Kubernetes `AgentRun` adapter implements `A0` and the concrete `A3` path; it rejects `A1` and `A2` instead of silently accepting levels whose controls it does not verify. Teams can reach `A1`–`A2` with the workflow runtime or another interactive/headless runner.

The operator's `A3` building blocks are cron and authenticated HTTP triggers, admission-minted decision and enforcement evidence, content-minimized per-run provenance, an authenticated runtime blocked signal, provider-routed escalation, authenticated exact-revision approve/reject responses, and pause/abandon defaults. Block reporting and accountable response use distinct non-equal credentials, so the runtime cannot answer its own escalation. The recipient router, tokens, persistent evidence storage, admission webhooks, and protected targets must all be deployed and observed; installing the operator alone does not confer `A3`. Where a required control is absent, the level is unavailable — see [Oversight coupling](#oversight-coupling).

## What autonomy grades

Autonomy grades a **posture**: how far a run advances before a human must act. It is an aggregate, not a property any single capability owns. Each task still selects its executor independently — see [execution.md](execution.md) for the executor classes and the selection rules. Autonomy describes what that population of per-task choices, plus gate policy and runner, adds up to for an environment or profile.

Levels use an `A` prefix and are cited as `A0`–`A3`. Cite the prefix: a bare "level 2" is ambiguous wherever another ladder numbers its own levels.

## Levels

| Level | Advances the run                                                    | Gates                                              | Runner                                        | Review                                          |
| ----- | ------------------------------------------------------------------- | -------------------------------------------------- | --------------------------------------------- | ----------------------------------------------- |
| `A0`  | A human invokes every capability                                    | Human decides inline, synchronously                | Interactive session                           | Human only                                      |
| `A1`  | A human invokes every capability                                    | Human decides, informed by independent critique    | Interactive session                           | Human + independent critique of exact revisions |
| `A2`  | Headless session runs unattended to the next human gate, then stops | Asynchronous human approval on exact revisions     | Sandboxed headless session                    | Automated findings + human sign-off             |
| `A3`  | Schedules, sensors, or humans trigger runs                          | Policy verdict + protected target + human sign-off | Isolated per-run jobs under admission control | Automated + human-in-the-loop escalation        |

`A1` adds an adversarial reader, not unattended execution: an independent critique of an exact revision, run in a fresh context rather than by the author. `A2` is the first level where work proceeds with nobody watching, so it is the first level that requires a run journal and asynchronous approval on exact revisions. `A3` permits non-human triggers, which is what makes admission control and escalation mandatory rather than optional. A hook is not required: a schedule, sensor, API, CI/CD event, provider webhook, agent-harness hook, another agent, or a human can originate the same exact-revision run.

A level is a ceiling for an environment, not a badge. Different profiles in one organization may sit at different levels, and a single profile may run high-impact capabilities at `A0` while routine ones sit at `A2`.

## What changes and what does not

Only three things change across levels: **gate policy**, **runner**, and **who advances the run**.

Unchanged at every level:

- the same capabilities, results, exact-revision references, and gate definitions;
- executor selection — a work shape still selects the least adaptive class that fits, and authority or ownership still selects `human` or `external`. Raising autonomy never widens a task's permitted executor class, and never converts a `human` executor into a `model` or `agent` one;
- authority attenuation and the absence of hidden agent launch;
- high-impact gates stay human. A gate is **low-impact** only when a wrong outcome is both detectable without a human and reversible without an external system of record; every other gate is **high-impact** and non-delegable at every level. Gate policy may auto-advance low-impact gates only. Each lifecycle guide names its own high-impact gates — see **plan-guide** for the planning and delivery gates.

Raising a level is therefore a change to **who is watching and when**, never a relaxation of what a task is permitted to do.

## Oversight coupling

**Never raise autonomy without matching oversight.** Autonomy and oversight are separate axes, and a level increase is only valid when the oversight that level requires is already in place and observed to work. This is a do-not-deploy coupling: where the required oversight is absent, unverified, or degraded, the level is not available and the run does not proceed at it.

Each level names the oversight it depends on:

| Level | Required oversight before the level is available                                                                 |
| ----- | ---------------------------------------------------------------------------------------------------------------- |
| `A1`  | Independent critique that the author cannot suppress                                                             |
| `A2`  | Run journal, asynchronous approval bound to exact revisions, cancellation, and a kill switch                     |
| `A3`  | Enforced policy verdict at a non-bypassable point, protected targets, escalation routing, and per-run provenance |

Oversight degradation is a demotion trigger, not a warning: when a control that a level depends on stops producing evidence, fails open, or drifts from intent, the effective level drops to the highest level whose oversight still holds. Detect this through [drift.md](drift.md) and act on it through [operations-and-learning.md](operations-and-learning.md).

Per-run provenance — model, prompt, tools, and permissions — is recorded by the runtime at every unattended level; see [data-and-telemetry.md](data-and-telemetry.md) for classification, routing, and retention.

## Escalation and safe defaults

An unattended run raises an escalation only after the runtime reports an authenticated, content-addressed blocked observation for the exact run revision. A desired-state boolean does not stand in for that observation. The runner supplies the endpoint, while the agent harness still needs a concrete blocked-state adapter, access only to the blocked-signal credential, network reachability, and an authenticated status wait before it resumes. The accountable-response credential remains outside the runtime. An escalation is bounded, never open-ended:

- declare the window and the safe default before the run starts, not when the escalation fires;
- **an unanswered escalation falls back to a safe default — pause or abandon — when the window expires.** Silence is never approval, and a timeout never advances a gate;
- `pause` preserves state for resumption and suits a reversible, non-expiring subject; `abandon` releases the workspace and suits an escalation whose subject, evidence, or authorization will be stale by the time a human answers. Choose per capability, not globally;
- record the escalation, the window, the expiry, and the safe default taken as evidence — an expired escalation is an outcome to review, not a silent non-event;
- an abandoned run preserves partial state, reason, and cleanup for retry.

Escalation routing is oversight, so `A3` is unavailable where the provider router cannot return a delivery reference, the response endpoint/token is unavailable, or the route has no accountable recipient. Approval and rejection responses retain responder identity and provider-native response reference; silence remains pause or abandon.

Advisory oversight does not satisfy the coupling. A control that is installed but not observed to execute leaves the level unavailable, however mature the surrounding configuration is — running unattended against advisory enforcement is the case this coupling exists to forbid.
