---
type: plan
has_subplans: true
status: complete
approved_date: "2026-07-17"
completed_date: "2026-07-19"
dependencies:
  plans:
    - composable-workflow-phases
    - composable-workflow-implementations-merge
    - environment-hardening
  files:
    - package.json
    - packages/skill/skill-workflow/workflow-guide/scripts/*.mjs
    - packages/skill/skill-workflow/workflow-guide/scripts/*.ts
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.mjs
    - packages/skill/skill-agent-governance/agent-governance-guide/scripts/*.ts
    - packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/*
    - packages/skill/skill-agent-governance/agent-governance-guide/references/autonomy.md
    - packages/skill/skill-claude-code/code-harness-guide/references/capabilities.md
    - packages/agent/agent-operator-go/api/v1alpha1/*.go
    - packages/agent/agent-operator-go/internal/webhook/*.go
    - packages/agent/agent-operator-go/internal/controller/*.go
    - packages/agent/agent-cli-go/internal/**/*.go
    - packages/command/command-workflow/commands/*.md
  subplans:
    - governance-decision-point-fail-closed
    - claude-code-native-hook-block
    - walking-skeleton-runtime-real
    - a3-unattended-orchestration-runtime
    - operational-proof-drift-incident
proposed_subplans:
  - governance-decision-point-fail-closed
  - claude-code-native-hook-block
  - walking-skeleton-runtime-real
  - a3-unattended-orchestration-runtime
  - operational-proof-drift-incident
parallel_groups:
  - group: 1
    plans:
      - governance-decision-point-fail-closed
  - group: 2
    plans:
      - claude-code-native-hook-block
      - a3-unattended-orchestration-runtime
  - group: 3
    plans:
      - walking-skeleton-runtime-real
  - group: 4
    plans:
      - operational-proof-drift-incident
skills_to_consult:
  - plan-guide
  - hexagonal-pattern-guide
  - microkernel-pattern-guide
  - kubernetes-guide
  - typescript-guide
  - shell-scripting-guide
  - testing-guide
  - connascence-guide
  - git-guide
  - moon-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  e2e: passed
updated: "2026-07-19"
---

# Runtime Enforcement Completion

## Execution evidence — 2026-07-19

Phases 1–4 are implemented and locally validated: the shared decision point fails closed, the real
Claude Code `PreToolUse` hook blocks denied work, the lifecycle probe executes real discovery and
independent validation commands, and the operator supplies governed A3 schedules/triggers,
provenance, escalation and admission controls. Phase 5's telemetry, drift detector, containment and
authenticated remediation path are also implemented and pass deterministic, unit and envtest gates.
After the host upgraded from runc 1.4.0 to runc 1.4.3 (`bb14dabeb7185bb72c8c86735d090dcb20f36587`),
a minimal Kind cluster reached Ready in 18 seconds and the targeted Kind-tagged A3 proof passed. The
proof created a governed A3 run and live Job, induced oversight degradation, demoted the run to A2,
deleted the Job, recorded durable containment evidence, and left the run Paused. The scenario passed
in 2.78 seconds (30.568 seconds for the e2e package including cluster lifecycle), and Kind removed the
cluster after the run. All five subplans are complete.

## Objective

Close the gap between a rigorously-specified governance framework and an enterprise agentic
**solution that actually enforces at runtime**. Today the governance contracts are _validated
by fixtures, not enforced against real agent decisions_. This plan builds the missing runtime
surface: a decision point that invokes the reference validators on live lifecycle operations
and fails closed; one harness whose native hook is proven to BLOCK a denied action end-to-end;
the A3 unattended-orchestration runtime (triggers, admission control, escalation, per-run
provenance) that `autonomy.md` names but nothing implements; a walking skeleton made
runtime-real; and operational proof that drift is detected and an incident contained on a live
estate.

This is the **sibling** of the composition-completion plan. The composition plan completes the
_workflow / composition surface_ (capabilities, profiles, methods, provider-native results); this
plan completes the _runtime / enforcement surface_ (does the governance plane actually stop a
prohibited action). The two are complementary and must cross-reference, not absorb, each other.
The specification baseline both extend is [composable-workflow-phases.md](composable-workflow-phases.md)
and its descendants ([composable-workflow-implementations-merge.md](composable-workflow-implementations-merge.md),
[environment-hardening.md](environment-hardening.md)).

## Decisions (settled 2026-07-17)

1. **Language seam — option (a), in TypeScript.** The decision point is a long-running
   TypeScript policy-decision service with a versioned JSON request/response contract that
   imports the authoritative validators — never a reimplementation. The validators migrate
   `.mjs` → erasable-syntax `.ts`, executed directly by Node via native type stripping; the
   root `engines` floor rises from `>=20.0.0` to `>=22.18.0` (target: Node 24 LTS). The floor
   raise is a shared precursor: the composition plan's new validators are authored `.ts`
   against it — whichever plan executes first lands it first.
2. **Deployment topology — endpoint-configurable, sidecar default.** The webhook calls a
   configurable decision-service endpoint (flag/env, default: pod-local sidecar) and denies
   mandatory operations when it is unreachable, regardless of topology. Phase 5 proves the
   sidecar topology only; a separate Deployment + Service is documented as
   supported-by-contract with a NetworkPolicy example, not e2e-proven.
3. **Live-agent proofs are maintainer-run probes.** Phases 2–3 prove via bundled repeatable
   probe scripts whose dated, version-stamped results land as capability-matrix evidence,
   re-earned on the refresh trigger; no PR gate requires a live model session; CI runs only
   the deterministic checks. Probe scripts are structured so a scheduled CI job could adopt
   them later without change.
4. **Trigger surface — both CRDs.** Phase 4 ships `AgentSchedule` (time-based CronJob analog)
   and `AgentTrigger` (event-based), each minimal and hand-maintained.
5. **`AgentTrigger` v1 event source — authenticated webhook receiver.** A
   bearer-token-authenticated HTTP receiver hosted as a runnable in the existing operator
   manager; token from a Secret, NetworkPolicy-gated; an authenticated POST creates the
   templated `AgentRun`, unauthenticated or unmatched posts create nothing. Phase 5's drift
   detector reuses this receiver as its remediation path.

## Baseline state (verified before implementation)

### The contracts are validated by fixtures, not enforced at runtime

The reference validators are pure functions returning failure codes:

- `checkIndependence` — `packages/skill/skill-workflow/workflow-guide/scripts/independence-helpers.mjs:34`
- `validateEmergencyAccess` and `validatePrivilegedOperation` — `.../scripts/operational-lifecycle-helpers.mjs:186,247`
- `selectDevelopmentExecutor` and `validateDevelopment` — `.../scripts/development-assurance-helpers.mjs:33,65`

A repo-wide grep for these symbols across `*.go`, `*.ts`, `*.js`, `*.mjs`, `*.md` (excluding
`node_modules`/`dist`) finds importers **only** in:

- their own fixture validators — `validate-operational-lifecycle-fixtures.mjs`,
  `validate-development-assurance-fixtures.mjs`, and the sibling helpers that compose them; and
- one cross-skill conformance check — `packages/skill/skill-agent-governance/agent-governance-guide/scripts/validate-executor-vocabulary.mjs:109`,
  which dynamically imports `selectDevelopmentExecutor` and also reads the helper source with
  regex (`readFileSync` at line 1) to compare declared vocabularies.

There are **24 fixture JSON files** under `packages/skill/skill-workflow/workflow-guide/assets/fixtures`.
The grep for the validators in `packages/agent/**/*.go` returns **nothing** — no CLI or operator
code invokes them. The validators decide independence, emergency access, privileged operations
(`integration|transition|release|data-deletion|retirement`), and executor selection, but no
running program ever asks them about a real decision.

### The commands are prompt files, not a runtime

`packages/command/command-workflow/commands/*.md` are markdown delegation prompts. For example
`observe-run.md:24` reads "Load `workflow-guide` and perform **observe-run**"; `incident-run.md`
and `workflow-drift.md` likewise delegate to `workflow-guide` / `agent-governance-guide` skill
operations. They instruct an interactive agent; they enforce nothing.

### The runtime surface that DOES exist

**`packages/agent/agent-operator-go`** — a Kubernetes operator (`cmd/operator/main.go`) that wires
three reconcilers (AgentRun, AgentProvider, AgentWorkspace) and two admission webhooks (AgentRun,
AgentToolchain) into a controller-runtime manager. CRDs in `api/v1alpha1/`: `AgentRun`,
`AgentHarness`, `AgentProvider`, `AgentWorkspace`, `AgentToolchain`, `AgentPolicy`.

- The `AgentRunReconciler` (`internal/controller/agentrun_controller.go`) turns an `AgentRun` into
  a K8s `Job` running a hardened agent container, creates a default-deny egress `NetworkPolicy` and
  a zero-RBAC ServiceAccount, and tracks phases `Pending → Initializing → Running →
Succeeded/Failed/TimedOut` (`api/v1alpha1/agentrun_types.go:23-30`).
- The `AgentRunWebhook` is a **live** admission point (`main.go:79`,
  `SetupWebhookWithManager`). Its `enforcePolicy` (`internal/webhook/agentrun_webhook.go:216-304`)
  enforces the `AgentPolicy.Spec.Enforced` constraints — `runtimeClassName`, `RequireSecurityContext`,
  `RequireNetworkPolicy`, `MaxTimeout`, `MaxResources`, `AllowedImages`, `AllowedRuntimeClassNames`
  (`api/v1alpha1/agentpolicy_types.go:13-41`). `agentpolicy_verdict_test.go` freezes these
  accept/deny verdicts.

**This is the intended enforcement point — but it enforces only pod/infra security, never workflow
governance.** `enforcePolicy` does not call `checkIndependence`, `validateEmergencyAccess`,
`selectDevelopmentExecutor`, or `validatePrivilegedOperation`. It cannot deny a _release without an
independent approver_ or an _executor escalation_; it only denies a privileged _pod_.

**`packages/agent/agent-cli-go`** — a sandbox launcher. `internal/cmd/run.go` resolves three
orthogonal confinement axes (isolation × provision × network) and runs `claude`/`opencode` inside a
cell, failing closed when a requested guarantee (pinned/host-tools-unreachable/egress-restricted)
cannot be established. It invokes no governance validators and registers no hooks.

### Gap 1 — no runtime invokes the validators (VERIFIED above)

The live enforcement point (AgentRun admission webhook) enforces infrastructure, not the governance
contracts. There is no decision point that runs `checkIndependence` / `validateEmergencyAccess` /
executor selection on an actual lifecycle operation, records the outcome as evidence, and fails
closed.

### Gap 2 — harness hook enforcement is documented, not demonstrated (VERIFIED)

Of the six adapters, only **Claude Code** has an observed runtime: `code-harness-guide/references/capabilities.md`
records `Observed runtime: 2.1.211 (Claude Code), probed 2026-07-16`. But line 14 states plainly:
"The guard contract (JSON event on stdin, exit 0 allow / exit 2 deny) was exercised locally;
**native hook registration was not**. Hook-level rows below remain documentation-verified and must
not be reported as runtime conformance." The other five — codex, copilot, kiro, pi, opencode — each
carry `Observed runtime: Not installed in the validation environment` /
`Evidence status: Documentation-verified; runtime-unverified`. Only `skill-claude-code` uses the
`code-harness-guide` naming; the rest are `<name>-guide`. Claude Code is the natural lead harness:
nothing has ever proven that a semantic-intent → native-hook mapping actually FIRES and BLOCKS a
denied action end-to-end.

### Gap 3 — the A3 autonomy runtime does not exist (VERIFIED)

`agent-governance-guide/references/autonomy.md:9` is explicit: "**`A3` is the eventual goal, not a
description of what exists.** Its triggers, admission control, and escalation routing are targets an
adopter builds and proves." A3 requires (line 24, and coupling table line 53): "Schedules, sensors,
or humans trigger runs / Policy verdict + protected target + human sign-off / Isolated per-run jobs
under admission control / Automated + human-in-the-loop escalation" and "Enforced policy verdict at
a non-bypassable point, protected targets, escalation routing, and per-run provenance."

Verification of the operator against that bar:

- **No trigger surface.** Grep for `schedul|trigger|sensor|cron` in operator `*.go` finds only
  pod-scheduling fields (`NodeSelector`/`Tolerations`) and test-only reconcile triggers. Reconcile
  fires on `AgentRun` creation; there is no schedule/sensor CRD.
- **No escalation router.** Grep for `escalat` finds only `AllowPrivilegeEscalation` (a securityContext
  field), not escalation routing.
- **No anomaly/kill-switch/break-glass.** Grep for `anomaly|falco|kill.switch|emergency|break.glass`
  in operator `*.go` returns nothing — despite `packages/agent/AGENTS.md` describing anomaly
  detection, a tested kill-switch, and break-glass as governance targets.
- **No accountable owner, no run journal / AIBOM.** `AgentRunSpec` (`agentrun_types.go:204-258`) has
  no owner field, and no code writes the model/provider/prompt/tools/permissions "run journal"
  `AGENTS.md` claims each run records. These are aspirational.

`autonomy.md:57` states per-run provenance "is recorded by the runtime at every unattended level" —
the runtime that would record it does not exist.

### Gap 4 — no end-to-end demonstration (VERIFIED)

The walking skeleton is `agent-governance-guide/assets/walking-skeleton/{run-skeleton.sh,guard.sh}`.
`run-skeleton.sh:6-7` declares it "Self-contained: mutates only a temp workspace; the repository is
never touched." It writes a scratch `project-settings.json` describing a `PreToolUse` hook
(lines 76-77) and then invokes the guard **directly** (`run_guard` → `bash "$GUARD"`, line 39) —
never through a real Claude Code hook dispatch. The capability matrix it checks is a hardcoded literal
(`MATRIX='{"before-tool-use":...}'`, line 109). It faithfully proves the _lifecycle shape_
(discover → preview → consent → apply → verify → evidence → drift → rollback) and the _guard's
decision logic_, but it is a simulation: no agent takes a real task through the lifecycle, and no
real harness loads the applied config.

### Gap 5 — no operational proof (VERIFIED)

`observe-run`, `incident-run`, and `workflow-drift` are contracts (markdown delegation prompts, see
above). No OpenTelemetry / telemetry wiring exists in the agent runtime: grep for
`opentelemetry|otel|telemetry|trace.Span` across `packages/agent/**/*.go` returns nothing. Drift is a
prompt-driven "evaluate drift" operation, not a detector fed by live signals; no incident has ever
been contained on a running estate by these contracts.

## Out of scope

- **Publishing / release.** Version bumps, `:ci-publish`, npm/OCI publication, marketplace
  registration, and the release PR flow are explicitly excluded (maintainer decision). Packaging
  metadata changes only insofar as a new runtime component needs to build and test locally.
- **The composition / workflow surface** — capability contracts, profile composition, method skills,
  provider-native result envelopes, and the CI/enterprise-platform onboarding packs. Those belong to
  the sibling composition plan and the completed `composable-workflow-*` plans. This plan consumes
  the validators and contracts as-is and does not redesign them.
- **Rewriting the validator semantics.** The reference `.mjs` validators are treated as the
  authoritative decision logic; this plan wires them in, it does not re-litigate their rules.
- **Full six-harness runtime parity.** Only the lead harness (Claude Code) is proven end-to-end here;
  the other five stay documentation-verified until each is separately probed (their capability
  matrices already say so).
- **Multi-tenant / production-grade telemetry backend selection.** Operational proof uses an
  OTel-compatible pipeline sufficient to demonstrate detection and containment, not a hardened
  enterprise observability platform.

## Phases

Phases are ordered by dependency. Phase 1 (a runtime that invokes the validators and records a
verdict) and Phase 2 (one proven enforcing hook) are the load-bearing foundation; Phase 3 proves them
together on a real lifecycle; Phases 4 and 5 (A3 orchestration, operational proof) build only on a
verdict that is already enforced at a non-bypassable point.

### Phase 1 — Governance decision point that invokes the validators and fails closed

**Objective.** Make the reference validators load-bearing: a running decision point evaluates a real
lifecycle operation with `checkIndependence` / `validateEmergencyAccess` / `validatePrivilegedOperation` /
`selectDevelopmentExecutor`, returns an allow/deny verdict with the exact failure code, records the
verdict as provider-native evidence, and fails closed on any error or missing input.

**Tasks.**

- Per Decision 1, raise the root `engines` Node floor to `>=22.18.0` (target Node 24 LTS),
  migrate the validator `.mjs` files to erasable-syntax `.ts` run directly via native type
  stripping, and expose them as a long-running **TypeScript policy decision service** with a
  stable, versioned JSON request/response contract, keeping decision separate from enforcement
  per the framework (`composable-workflow-phases.md` plane 2). New scripts live beside the validators under
  `packages/skill/skill-workflow/workflow-guide/scripts/` and
  `packages/skill/skill-agent-governance/agent-governance-guide/scripts/`; do not fork the validator
  logic.
- Implement a thin **enforcement adapter** that calls the decision service — through a
  configurable endpoint (flag/env, default: the pod-local sidecar per Decision 2) — at a real gate. Primary
  target: extend the operator's live admission surface so a governed lifecycle operation (modelled on
  an `AgentRun` annotation/label or a new `AgentPolicy.Spec.Enforced` governance field) is checked by
  the verdict — `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go`,
  `api/v1alpha1/agentpolicy_types.go`. Keep the existing infra `enforcePolicy` intact; add the
  governance verdict as a second, independent deny path (defense in depth).
- Emit a **verdict evidence record** (decision, failure code, policy version, subject reference,
  correlation id) on every check, allow or deny — reusing the walking skeleton's evidence shape
  (`run-skeleton.sh` `record_evidence`) as the schema seed.
- Fail closed: any validator error, unreachable decision service, or absent required input denies a
  mandatory-profile operation (mirror `guard.sh:34-41`).
- Add tests: a table of governed operations (independent vs self-approved release; in-scope vs expired
  emergency access; mechanical vs adaptive executor request) asserting the exact failure code and the
  accept/deny verdict — the governance analogue of `agentpolicy_verdict_test.go`.

**Acceptance criteria.**

- A grep for the four validators shows at least one **non-fixture, non-vocabulary** importer: the new
  decision service.
- Running the decision point against a self-approved release returns deny with
  `emergency-access-independence-failed` (or the acceptance independence code); against a compliant
  one returns allow — both asserted by an automated test.
- The operator admission path denies a governed operation whose verdict is deny, with the failure code
  surfaced in the admission response, and a passing operation is admitted — proven by a webhook test.
- Every check writes exactly one verdict evidence record (idempotent per correlation id); a decision-
  service outage denies a mandatory operation and observes an advisory one.
- `go test ./...` (operator) and the workflow/governance `validate-*` scripts pass.

### Phase 2 — One harness native hook proven to block a denied action end-to-end

**Objective.** Turn Claude Code's capability matrix from documentation into runtime conformance: a
semantic-intent → native-hook mapping is actually **registered** with the installed runtime and is
observed to **block** a denied tool call end-to-end — the load-bearing proof that the governance plane
can enforce at the harness.

**Tasks.**

- Register the guard as a real Claude Code `PreToolUse` hook (not a scratch file the harness never
  loads): produce the actual settings entry and registration procedure under
  `packages/skill/skill-claude-code/code-harness-guide/` and its walking-skeleton assets, wiring the
  existing `guard.sh` (or the Phase 1 decision service) as the handler.
- Map the semantic intent `before-tool-use` → native `PreToolUse` and drive a real session: a denied
  operation (write to a protected path) must be **blocked by the harness** (tool call refused via exit
  `2`), and a permitted operation must proceed — observed against the probed runtime `2.1.211`.
- Have the hook call the **Phase 1 decision point** for its verdict, so the harness enforcement and the
  admission enforcement share one decision logic (single source of truth; connascence kept at the
  decision service, not duplicated in the hook).
- Upgrade `code-harness-guide/references/capabilities.md`: change the `PreToolUse` /
  `Tool before use` evidence from documentation-verified to **runtime-verified**, and update the
  matrix-identity `Evidence status` line to record that native hook registration was exercised — with
  the probe date and observed version. Do **not** upgrade rows that were not actually exercised.
- Add a repeatable end-to-end test/harness-probe script (bundled with the adapter) that a maintainer
  can re-run to reconfirm the block after a harness update, per the matrix `Refresh trigger`. Per
  Decision 3 this proof is a maintainer-run probe, not a CI gate — CI runs only the deterministic
  checks — and the script is structured so a scheduled CI job could adopt it unchanged.

**Acceptance criteria.**

- With the hook registered, a real Claude Code session is **denied** a write to a protected path (the
  action does not occur) and **allowed** a write to an unprotected path — captured in a reproducible
  transcript/log, not a simulation that invokes the guard directly.
- The hook obtains its verdict from the Phase 1 decision service (the same code that backs admission),
  demonstrated by a shared verdict evidence record for the blocked action.
- `capabilities.md` `Evidence status` and the two tool-use rows accurately state runtime-verified
  registration with the probe date/version, and no un-exercised row is upgraded.
- The refresh probe script reproduces the block on a clean checkout.

### Phase 3 — The walking skeleton made runtime-real

**Objective.** Replace the self-contained simulation with a real end-to-end run: an agent takes a task
through the lifecycle (discovery → … → acceptance → integration) with the Phase 1 decision point
enforcing at each governed gate and provider-native evidence recorded at each — the walking skeleton
turned into a live, observable run.

**Tasks.**

- Author a real end-to-end scenario driving the lead harness (Phase 2) through a minimal but genuine
  lifecycle, invoking the actual lifecycle commands/skill operations
  (`packages/command/command-workflow/commands/*.md`) rather than a bash mock, and gating the
  high-impact transitions (acceptance, integration) through the Phase 1 verdict.
- At each gate, record a provider-native evidence record with the exact-revision reference and the
  verdict; assert the run **cannot advance** past a gate whose verdict is deny.
- Keep the existing `run-skeleton.sh` as the deterministic contract-shape check but add a
  runtime-real counterpart under `agent-governance-guide/assets/walking-skeleton/` that exercises the
  real harness + real decision point (clearly labelled which is simulation vs live). The live
  counterpart is a maintainer-run probe per Decision 3; CI runs only the deterministic simulation.
- Prove one negative path: a self-approved acceptance or an executor escalation is blocked mid-run and
  produces an escalation/deny evidence record.

**Acceptance criteria.**

- A single command drives an agent from discovery to integration with a governed gate enforced by the
  Phase 1 verdict at acceptance and integration; the evidence trail shows a verdict record per gate
  tied to an exact revision.
- Injecting a governance violation (self-approval, out-of-scope emergency access, or executor
  escalation) halts the run at the offending gate and records the failure code — verified, not
  simulated.
- The live counterpart is distinguishable in-repo from the simulation, and both pass their checks.

### Phase 4 — A3 unattended-orchestration runtime coupled to the oversight invariant

**Objective.** Build the A3 components `autonomy.md` names as targets — non-human triggers, admission
control, escalation routing, and per-run provenance — so unattended orchestration is real and cannot
run without the oversight it depends on.

**Tasks.**

- Add a **trigger surface** to the operator per Decisions 4-5: two new CRDs in `api/v1alpha1/` with
  reconcilers in `internal/controller/` — `AgentSchedule` (time-based CronJob analog: cron
  expression, `AgentRun` template, `suspend`, concurrency policy) and `AgentTrigger` (event-based:
  a declared endpoint whose bearer token comes from a Secret, served by an HTTP receiver hosted as
  a runnable in the existing manager, NetworkPolicy-gated; an authenticated POST creates the
  templated `AgentRun`, unauthenticated or unmatched posts create nothing) — so "schedules,
  sensors, or humans trigger runs" is real, not just human-initiated reconcile.
- Add **per-run provenance**: an accountable-owner field on `AgentRunSpec`
  (`api/v1alpha1/agentrun_types.go`) that admission requires for triggered runs, and a run journal
  recording model/provider/prompt/tools/granted-permissions (the AIBOM `AGENTS.md` describes) written
  by the reconciler.
- Add an **escalation router**: an unattended run needing a human raises a bounded escalation with a
  declared window and a safe default (pause/abandon) per `autonomy.md:59-71`; an unanswered escalation
  falls back to the safe default on expiry and records the outcome. `A3` is unavailable where the
  route has no accountable recipient — encode that as a fail-closed admission check.
- Couple autonomy to oversight: gate the ability to run at A3 on the Phase 1 verdict being enforced at
  a **non-bypassable** point (the admission webhook), protected targets, escalation routing present,
  and per-run provenance recorded — and demote to the highest level whose oversight still holds when a
  control stops producing evidence (the `autonomy.md` demotion trigger).
- Tests: a triggered run without an accountable escalation recipient is refused; an expired escalation
  takes its declared safe default; a run missing provenance is denied admission at A3.

**Acceptance criteria.**

- A schedule/sensor creates an `AgentRun` with no human in the loop, and that run carries an
  accountable owner and a per-run provenance journal — both asserted by tests.
- An unattended run that needs a human raises an escalation with a window and safe default; on expiry
  the safe default is taken and recorded; silence never advances a gate.
- Admission refuses an A3 run when the governance verdict is not enforced at a non-bypassable point,
  when the escalation route has no accountable recipient, or when provenance is absent.
- `go test ./...` (operator, incl. new CRD/reconciler/webhook) passes.

### Phase 5 — Operational proof: drift detection and incident containment on a live estate

**Objective.** Wire `observe-run` / `incident-run` / `workflow-drift` to real telemetry so drift is
actually detected and an incident actually contained on a running estate — the operational contracts
made real.

**Tasks.**

- Emit OpenTelemetry-compatible signals from the runtime (operator reconciler + Phase 1 decision
  point + Phase 2 hook): per-run traces and verdict/enforcement metrics, correlated by the run's
  correlation id, with content minimization (no prompts/secrets by default) per
  `composable-workflow-phases.md` plane 7. New telemetry seam in
  `packages/agent/agent-operator-go/internal/` (and the decision service).
- Implement a **drift detector** fed by those signals that recomputes effective autonomy against
  required oversight (`autonomy.md` coupling) and raises drift when a control stops producing
  evidence, fails open, or diverges from the applied reference — wiring the `workflow-drift` contract
  to live data instead of a prompt.
- Implement **incident containment**: on a detected drift or anomaly, exercise a tested kill-switch /
  pause on the live run (building on the operator's Job/AgentRun lifecycle and the escalation router
  from Phase 4), and record the containment as evidence — wiring `incident-run`. Remediation runs
  are raised by POSTing to the Phase 4 `AgentTrigger` receiver (Decision 5) — no second trigger
  mechanism.
- Demonstrate on a live (kind/e2e) estate: induce drift (disable a required control) and an incident
  (a run breaching policy), and show detection + containment with evidence. The estate runs the
  sidecar decision-service topology (Decision 2) — the only topology e2e proves; document the
  separate-Deployment alternative with a NetworkPolicy example as supported-by-contract, not
  e2e-proven.

**Acceptance criteria.**

- The runtime emits OTel-compatible traces/metrics correlated per run without logging sensitive
  content by default.
- Disabling a required oversight control is detected as drift on the live estate and demotes the
  effective autonomy level, with a drift evidence record.
- A policy-breaching run is contained (paused/killed) by the incident path on the live estate, with a
  containment evidence record and no manual intervention.
- The e2e proof runs under the operator's existing `-tags=e2e` harness.

## Risks and unknowns

- **Language seam (Go operator vs Node validators) — DECIDED (Decision 1).** Option (a): a
  TypeScript policy-decision service the operator calls over a local contract, keeping decision and
  enforcement decoupled per the framework. Residual risk: the `.mjs` → `.ts` migration and the Node
  engine-floor raise touch every skill package that runs validators, and direct execution requires
  erasable-syntax-only TypeScript (no enums/namespaces/parameter properties).
- **Admission timing vs lifecycle semantics.** The AgentRun webhook fires at pod-creation admission,
  but governance gates (acceptance, integration, release) are lifecycle events that do not map 1:1 to
  a K8s object create. Modelling a "governed operation" as an admissible resource (or a distinct
  decision-service call from the harness/CI) needs design; the operator webhook may not be the right
  point for every gate — defense in depth across harness hook + admission + CI is the intended answer.
- **Native hook registration may differ from documented behavior.** Claude Code hooks are the only
  probed surface and `PreToolUse` blocking is documentation-verified; the runtime block in Phase 2 may
  reveal event/exit/permission-mode nuances (the matrix already warns "handler availability
  event-specific"). Budget for reconciling observed behavior with the matrix.
- **`agent` handler is experimental.** `capabilities.md:30` marks the `agent` handler "Experimental /
  Cannot satisfy mandatory profile"; Phase 2 must use `command`/decision-service enforcement, not the
  agent handler, for the mandatory block.
- **A3 CRD surface is greenfield.** No scheduler/sensor/escalation CRD exists; controller-gen is noted
  broken on Go 1.25+ (CRDs and deepcopy maintained by hand — `agent-operator-go/AGENTS.md`), so new
  CRDs carry manual-maintenance cost and review risk.
- **Telemetry privacy.** Operational proof must not regress the data-minimization posture; content
  capture stays off by default, which constrains how much drift signal is available.
- **Sibling coupling.** The composition plan and this plan share the validators, evidence shape, and
  autonomy contract. If the composition plan changes a validator's contract, Phase 1's decision
  service must track it — coordinate the seam, do not duplicate.

## Success criteria

Completion of the runtime / enforcement surface means:

- At least one running program invokes the reference validators on real lifecycle operations, returns
  an allow/deny verdict with the exact failure code, records verdict evidence, and fails closed — the
  validators are load-bearing, not fixture-only (Phase 1).
- One harness (Claude Code) is proven, against its probed runtime, to **block** a denied action
  end-to-end through a registered native hook that shares the Phase 1 decision logic, with
  `capabilities.md` honestly upgraded to runtime-verified for exactly the exercised rows (Phase 2).
- An agent takes a real task through discovery → acceptance → integration with governance enforced at
  each gate and provider-native evidence recorded, and a governance violation halts the run — the
  walking skeleton is runtime-real, not a simulation (Phase 3).
- Unattended orchestration is real: non-human triggers create runs, each run carries an accountable
  owner and per-run provenance, escalations are bounded with safe defaults, and A3 is refused wherever
  its required oversight is absent, unverified, or degraded (Phase 4).
- Drift is detected and an incident contained on a live estate from real telemetry, with the effective
  autonomy level demoted when oversight degrades (Phase 5).
- Defense in depth holds: a mandatory control fails closed at two independent enforcement points
  (harness hook + admission) in the reference composition.
- Validation gates pass without regression: operator `go test ./...` (unit + integration + e2e tags),
  the workflow/governance `validate-*` scripts, typecheck, lint, and build.
