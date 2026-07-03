---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - plans/agentic/agentic-workflow-evolution/02-workflow-definitions.md
  files:
    - packages/shared/shared-agent-go/pkg/workflow/**
    - packages/agent/agent-cli-go/internal/cmd/**
    - packages/agent/agent-cli-go/internal/config/loader.go
    - packages/agent/agent-cli-go/internal/sandbox/**
skills_to_consult:
  - fp-guide
  - testing-guide
  - git-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 04 — agent-cli workflow run: The L2 Interpreter

## Objective

Supervised autonomy on the local machine: a Go interpreter in
`shared-agent-go/pkg/workflow` (the reference implementation of 02's
semantics, reused by L3 in 05) and an `agent-cli workflow run` loop —
fresh sandboxed session per step with the pinned plugin set provisioned
(decision 9), commits under the bot identity (decision 12), Discord
notification at gates (decision 11), synthesized failures with bounded
retries, budget/iteration enforcement.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); re-read cited files —
the improvement plan and subplans 01–02 land first. NOTE: no Go
guideline skill exists (parent risk); apply fp-guide principles
(module-level functions, explicit state) and the existing package style
in `shared-agent-go/pkg/*`.

1. `agent-cli-go` today: `run [agent-args...]` with isolation ×
   provision × network axes; headless mode = no terminal (stdin closed,
   logs to stdout); config at `internal/config/loader.go`
   (`homeDir`, `bindPaths`, `roBindPaths`, `customEnv`); Cobra CLI,
   `internal/cmd/run.go`. `internal/sandbox/plugins/plugins.go` is the
   ISOLATOR/PROVISIONER registry — naming collision with Claude Code
   plugins; do not confuse them.
2. **Claude-plugin provisioning gap (decision 9, critique blocking
   finding)**: sandboxed sessions get an isolated home — the xonovex
   command/skill plugins do NOT exist inside. This subplan adds
   provisioning: a pinned plugin set (source: local marketplace cache
   or repo checkout — decide; pin by version) materialized into the
   sandbox home's harness config before the session starts. Design it
   in `shared-agent-go` so 05 reuses the manifest for image baking.
3. Steps are invoked headless as
   `<agent> -p "/xonovex-workflow:<step> <args>"` (claude supports
   prompt mode; the agent registry in `shared-agent-go/pkg/agents`
   knows per-agent invocation shape — extend it with a
   headless-prompt builder rather than hardcoding claude).
4. Interpreter semantics come from workflow-guide (02) — table-driven
   tests in pkg/workflow are the drift guard between this and L1.
5. Results: the step's session writes the journal file (01); the RUNNER
   validates it after session exit and synthesizes
   `failed + synthesized: true` if missing/invalid; retry per-step cap
   from the definition.
6. Bot identity (decision 12): loop commits journal/status under the
   bot user with a `Workflow-Run: <run-id>` trailer; key from the local
   keychain; pushes only `agent/*` branches. Budgets: wall-clock +
   max-steps from the definition (token budgets deferred — record as a
   follow-up, parent risk accepted).
7. Parallel groups: one worktree per subplan in the group
   (`pkg/workspace` has worktree naming); sessions run concurrently
   bounded by a `--concurrency` flag.

## Tasks

1. **`shared-agent-go/pkg/workflow`** — types (Definition, Step, Gate,
   Policy), YAML parse + validate (schema from 02), `Next(state)`
   computation (foreach/loop/gate rules), gate policy evaluation with
   trusted-ref loading and most-restrictive-wins hook (floor arrives in
   05), journal read/write + result validation + synthesis. Table-driven
   unit tests mirroring 03's fixture walkthrough cases.
2. **Plugin-set provisioning** — `pkg/` support: a pinned plugin
   manifest (name@version list) + materializer that installs the set
   into a target home dir; agent-cli wires it into sandbox home setup;
   config gains the manifest path with a sane default.
3. **`agent-cli workflow run` command** — Cobra subcommand: flags
   `--workflow`, `--plan-file`, `--concurrency`, `--budget-minutes`,
   `--max-steps`, plus existing sandbox axes; loop = Next → spawn
   headless sandboxed session for the step (per-agent prompt builder) →
   await exit → validate/synthesize result → commit journal (bot
   identity + trailer) → repeat; halt at manual/assisted gates and on
   budget/cap exhaustion.
4. **Gate halt + notify** — Discord webhook (URL from config/env):
   message = workflow, plan, gate, verdict-so-far, direct link (repo
   file URL or PR); also plain stdout for terminal supervision.
5. **Worktree fan-out** — parallel group execution: worktree per
   subplan, bounded concurrency, group barrier before merge steps (the
   definition's foreach semantics).
6. **e2e-style test** — scripted fake agent (provision: command) that
   writes results; drive a 2-subplan toy plan through the loop: halts
   at the manual gate; killing a step mid-run synthesizes failed +
   retries once (parent success criterion 4 mechanics, minus real LLM).

## Validation Steps

- `npx moon run shared-agent-go:test agent-cli-go:test` green
  (pkg/workflow table tests + loop e2e with fake agent).
- `:lint :typecheck :build` green for both Go packages.
- Live smoke (real claude, cheap model, tiny toy plan): plugin set
  present inside the sandbox (step command resolves), journal complete,
  Discord message received at the gate.

## Success Criteria

- [ ] pkg/workflow implements 02's semantics with table-driven tests
      (shared with 03's fixtures — same cases, same outcomes).
- [ ] Sandboxed sessions contain the pinned plugin set; steps resolve.
- [ ] Loop halts at manual/assisted gates with Discord notification;
      budget/max-steps enforced; kill-mid-step → synthesized failed +
      bounded retry.
- [ ] Journal commits carry bot identity + `Workflow-Run:` trailer.

## Files Modified/Created

- Created: `shared-agent-go/pkg/workflow/**`, plugin-manifest support,
  `agent-cli-go/internal/cmd/workflow*.go`
- Modified: `agent-cli-go/internal/config/loader.go`, sandbox home
  setup, agent registry (headless prompt builder)

## Dependencies

Requires 02 (semantics + schema). Parallel with 03 (disjoint files).
Blocks 05.

## Estimated Duration

~1.5 weeks.
