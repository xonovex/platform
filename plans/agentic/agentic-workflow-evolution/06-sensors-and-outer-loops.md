---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 5
status: pending
dependencies:
  plans:
    - plans/agentic/agentic-workflow-evolution/05-workflowrun-crd.md
  files:
    - packages/agent/agent-operator-go/cmd/sensor/**
    - packages/agent/agent-operator-go/internal/sensor/**
    - packages/agent/agent-operator-go/internal/controller/**
    - workflows/incident.yml
    - workflows/maintenance.yml
skills_to_consult:
  - kubernetes-guide
  - testing-guide
  - fp-guide
  - gitlab-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 06 — Sensors & Outer Loops: Events Become Proposed Workflows

## Objective

Turn runtime signals into supervised work (parent decisions 5, 11;
boundary: signals return ONLY through sensors): a webhook receiver for
GitLab pipeline events and Alertmanager alerts with fingerprint dedup
and cooldown, creating WorkflowRuns paused at an entry gate; workflow
definitions for the first two outer-loop agents (CI Sentinel,
Validation Auditor); Fleet Supervisor rules in the controller (stuck
detection, budget cap, kill-switch) escalating via Discord.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); subplans 01–05 land
first — re-read everything. Deployment (Flux tree, egress exception,
GitLab webhook registration, Alertmanager receiver config) is the
DRODAN-SIDE runtime plan; here everything must be testable with
simulated payloads (parent success criterion 6).

1. Receiver shape: a small HTTP service in agent-operator-go
   (`cmd/sensor`) — separate binary, same module, so it deploys
   independently of the controller and holds only
   create-WorkflowRun RBAC. Endpoints: `/gitlab` (pipeline events;
   token-checked) and `/alertmanager` (webhook format; Mimir ruler,
   Falco via Falcosidekick, Flagger events all arrive through the obs
   Alertmanager per parent research).
2. Storm control (parent risk): fingerprint = source + stable identity
   (GitLab: project+ref+failure signature; Alertmanager: its
   fingerprint field). Dedup rule: at most one OPEN WorkflowRun per
   fingerprint; cooldown window after closure. State lives in the
   created WorkflowRuns' labels — no extra store; the receiver queries
   by label before creating.
3. Every sensor-created WorkflowRun starts PAUSED at an entry gate
   (`triage-review`, policy assisted) — autonomy to *propose*, human
   authority to *start*, per the maturity ladder. The proposal's first
   step is the analyst agent (Sentinel/Auditor) whose triage brief
   becomes the gate's review material.
4. Workflow definitions: `workflows/incident.yml` (obs alerts → triage
   → gate → research → plan → fix → PR) and `workflows/maintenance.yml`
   (audit/eval reports → triage → gate → backfill/fix steps). Sentinel
   and Auditor are STEPS (agents executing one operation) — not
   standing services; their prompts live as workflow-guide reference
   material or command operations, consistent with the steps-are-
   commands model.
5. Fleet Supervisor (parent role list) is mostly deterministic and
   belongs in the WorkflowRun controller: no-progress-for-N detection,
   global concurrent-run cap, kill-switch (a cluster-scoped annotation
   or config flag that pauses all dispatch), each escalating via the
   decision-11 Discord webhook.

## Tasks

1. **Receiver** — `cmd/sensor` + `internal/sensor`: payload parsing for
   the two endpoints, token auth, fingerprint computation, label-query
   dedup + cooldown, WorkflowRun creation (definition snapshot from
   main per the trusted-ref rule, entry gate pre-set). Unit tests with
   recorded GitLab/Alertmanager payload fixtures.
2. **Definitions** — `workflows/incident.yml` and
   `workflows/maintenance.yml` per Context 4, validated by
   script-moon-workflow-validate; triage steps reference the analyst
   operations.
3. **Analyst operations** — CI Sentinel (classify pipeline failure:
   flaky / regression / environment; propose resume-fix vs new fix
   workflow) and Validation Auditor (consume the weekly audit / eval
   reports; propose backfill or trigger-tuning work) as workflow-guide
   reference operations, read-only toward the cluster.
4. **Fleet Supervisor** — controller additions: stuck detection
   (no status transition in N minutes → condition + notify), global
   run cap (config), kill-switch flag halting new AgentRun creation;
   Discord escalation messages with direct links.
5. **Simulated e2e** — envtest/kind: POST a recorded GitLab
   pipeline-failed payload twice → exactly ONE WorkflowRun exists,
   paused at triage-review (parent success criterion 6); flip the
   kill-switch → dispatch halts; stuck run → notification recorded
   (webhook stubbed).

## Validation Steps

- Sensor unit tests (fixtures, dedup, cooldown) green;
  `npx moon run agent-operator-go:test` green incl. new envtest cases.
- `script-moon-workflow-validate` accepts both new definitions.
- `:lint :typecheck :build` green.

## Success Criteria

- [ ] Duplicate simulated events yield one paused WorkflowRun (dedup +
      cooldown verified).
- [ ] incident.yml / maintenance.yml validate; entry gates assisted;
      analyst operations documented and read-only.
- [ ] Kill-switch, run cap, and stuck detection work in envtest with
      Discord escalation stubbed.
- [ ] Nothing here mutates a cluster beyond WorkflowRun CRs (parent
      decision 5 boundary).

## Files Modified/Created

- Created: `cmd/sensor/**`, `internal/sensor/**`,
  `workflows/incident.yml`, `workflows/maintenance.yml`, analyst
  operation references in workflow-guide
- Modified: WorkflowRun controller (supervisor rules), operator config

## Dependencies

Requires 05. Final group; the drodan-side runtime plan consumes this.

## Estimated Duration

~1 week.
