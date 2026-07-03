---
type: roadmap
status: active
updated: 2026-07-03
repos:
  xonovex-platform: .
  drodan-platform: ../../drodan/drodan-platform
tracks:
  A: plans/agentic/platform-workflow-improvement.md
  B: plans/agentic/agentic-workflow-evolution.md
  S: plans/agent-security-hardening.md
  C: ../../drodan/drodan-platform/plans/drodan/drodan-multi-cluster-rollout.md
  D: (to be created) drodan-platform/plans/drodan/agentic-workflow-runtime.md
---

# Cross-Repo Roadmap: Catalog → Agentic Workflow → Fleet Runtime

## How to use this document

This is the ONE place that orders all in-flight and planned work across
xonovex-platform (this repo) and drodan-platform. Executors — human or
agent — read in this order:

1. This roadmap (order, dependencies, glossary, execution protocol).
2. The parent plan of the subplan you are executing (its Decision Log
   is normative whenever a subplan cites "decision N").
3. The subplan itself — each has a "Context (read this first)" section
   and is self-contained given that section + the parent decision log.
4. Every skill in the subplan's `skills_to_consult` frontmatter (in
   this repo under `packages/skill/skill-<name>/<name>-guide/SKILL.md`,
   or installed as plugins).

Never start a subplan whose Dependencies section lists an unmerged
prerequisite. When file/line anchors don't match (they are pinned to
`main @ 2b276a7f`, 2026-07-03), locate the named construct instead —
plans older than the code they describe are expected here.

## The five tracks

| Track | Plan | Repo | Status | Size |
|---|---|---|---|---|
| **A** | `plans/agentic/platform-workflow-improvement.md` — catalog gates, skill refs, command distillation, plumbing, dedup, backfill | xonovex | pending-approval | 6 subplans, ~1.5–2 wks |
| **B** | `plans/agentic/agentic-workflow-evolution.md` — step contract, workflow definitions, L1/L2/L3 interpreters, sensors | xonovex | pending-approval | 7 subplans, ~4–6 wks |
| **S** | `plans/agent-security-hardening.md` — operator webhooks, fail-closed enforcement, CLI hardening | xonovex | approved (in flight) | 7 subplans |
| **C** | `plans/drodan/drodan-multi-cluster-rollout.md` — five-cluster Talos/Cilium fleet (mgmt, obs, ci, main, staging) | drodan | approved | 9 subplans, ~8–11 wks |
| **D** | `agentic-workflow-runtime` — deploy the B agent stack to the C fleet's ci cluster | drodan | NOT YET CREATED (by design; see M5) | est. ~1–2 wks |

Unrelated but present: `plans/vite8.1-upgrade-hold.md` (blocked-upstream
runbook, no interaction with these tracks);
`plans/drodan/platform-infrastructure-modernization.md` (C's own
prerequisite, drodan-side).

## Milestones in order

**M0 — Ready to start (now).**
Approve A and B (`/xonovex-workflow:plan-revise --final`; done
2026-07-03). S and C are already approved. Commit the plan documents.

**M1 — Catalog & workflow foundation (Track A, all 6 subplans).**
Order inside A: group 1 = `01-quality-gates` + `02-skill-references` +
`05-catalog-dedup` in parallel (if 05's Branch A triggers, merge 01
first) → group 2 = `03-command-distillation` + `06-sources-eval-backfill`
→ group 3 = `04-workflow-plumbing`. Each group is releasable.
Value shipped: real CI gates, documented pr ops, thin commands,
plan-list/association plumbing. Release note: the 2026-07-03 lifecycle
rename (`plan-decide`, `plan-revise` — A decision 7) is breaking by
semver; the owner shipped it as the 4.1.0 lockstep bump instead
(applied 2026-07-03).

**M2 — Workflow becomes data (Track B, subplans 01–02).**
Starts once A's `04-workflow-plumbing` merges (A05/A06 need NOT be
done). `B01-step-result-contract` → `B02-workflow-definitions`.
Value shipped: machine-readable step results; `workflows/feature.yml`;
the workflow-guide skill.

**M3 — Assisted + local autonomy (Track B, subplans 03 ∥ 04).**
`B03-workflow-next-command` (L1) and `B04-agent-cli-workflow-run` (L2)
in parallel — disjoint files (markdown/commands vs Go).
Value shipped: `workflow-next` suggests/executes the next step; headless
supervised runs work on the local machine. This is the first big
usability payoff — reachable without ANY drodan dependency.

**M4 — Cluster engineering (Track B, subplans 05 ∥ 07, then 06).**
`B05-workflowrun-crd` requires Track S's operator webhook deployment
(S01) and fail-closed enforcement (S03) to be merged — check S's status
before starting. `B07-workflow-diagram-refresh` runs in parallel
(disjoint files). Then `B06-sensors-and-outer-loops`.
Value shipped: L3 fully engineered and tested (envtest/e2e) — not yet
deployed anywhere.

**Track C runs in parallel with M1–M4** (different repo, different
files): C01 cilium → C02 template/provisioning → [C03 obs, C04 mgmt,
C05 ci] → [C06 signing, C07 gateway] → C08 cutover → C09 docs. Nothing
in A/B/S waits for C before M5.

**M5 — Fleet runtime (Track D, drodan repo).**
Prerequisites: B05 + B06 merged (the engine exists) AND C03 (obs) +
C04 (mgmt) + C05 (ci cluster) applied (the fleet exists). At that
point, CREATE the plan `drodan-platform/plans/drodan/
agentic-workflow-runtime.md` (it deliberately does not exist yet —
creating it earlier would mean planning against an unbuilt fleet).
Its scope is already outlined in B's design decision 8: a
kubernetes-agent-operator package, flux tree entries on the ci
cluster, the deny-all-egress model-provider exception, the Alertmanager
sensor receiver route, GitLab webhook registration, bot-identity
credentials (B decision 12), and the AgentPolicy gate floor values.
Value shipped: L3 live — sensors turn CI/observability events into
paused, human-gated WorkflowRuns on the ci cluster.

**M6 — Promotion (operational, no plan document).**
After M5 soaks: flip eval reports from report-only to blocking (A
decision 3), promote individual gates assisted → auto one at a time (B
decision 3), each change via a reviewed `workflows/**` PR (B decision
10). Revisit: token budgets in the L2/L3 runners; a `go-guide` skill;
the deferred unified command-format migration (A deferred list).

## Dependency graph

```
xonovex-platform                                 drodan-platform
────────────────                                 ───────────────
A: improvement                                   C: rollout (approved)
 [A01 A02 A05] → [A03 A06] → [A04]                C01 → C02 → [C03 C04 C05]
                               │                        → [C06 C07] → C08 → C09
S: security-hardening          │                              │   │    │
 S01 → S03 (operator work) ─┐  │                              │   │    │
                            │  ▼                              │   │    │
B: evolution                ▼                                 │   │    │
 B01 → B02 → [B03 ∥ B04] → [B05 ∥ B07] → B06                  │   │    │
                             │             │                  ▼   ▼    ▼
                             └─────────────┴───────► D: agentic-workflow-runtime
                                                        (create at M5)
                                                              │
                                                              ▼
                                            L3 live on ci cluster → M6 promotion
```

Cross-repo rules:
- xonovex work (A, S, B) never blocks on drodan until M5.
- Every A/B group merge that touches plugin packages is releasable via
  a version-packages PR (lockstep bump — see glossary); release before
  starting work that depends on the published artifacts.
- The only two coordination points: B05 ↔ S (same operator codebase)
  and D ↔ C (deployment target).

## Glossary (read once; assumed by every plan)

- **Monorepo tooling**: `npm install` once at repo root. Tasks run via
  `npx moon run <project>:<task>` (e.g.
  `npx moon run agent-operator-go:test`) or by tag:
  `npx moon run '#skill:skill-validate'`. CI runs
  `npx moon ci :ci-check`. Tasks inherit from `.moon/tasks/tag-*.yml`
  based on each project's `tags` in its `moon.yml`.
- **Skill package**: `packages/skill/skill-<name>/<name>-guide/` with
  `SKILL.md` (frontmatter `name` + quoted `description`; body ends
  with a progressive-disclosure list) and `references/*.md` (detail
  loaded on demand). Optional: `SOURCES.md` (upstream provenance),
  `eval-queries.json` (trigger benchmarks). Each package also has
  `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` — these
  two files must stay byte-identical.
- **Command plugin**: `packages/command/<plugin>/commands/*.md`; each
  command is frontmatter (`description`, `allowed-tools`,
  `argument-hint`) + Arguments + a THIN Delegation section using the
  standard wording ("Load the `<skill>` skill (plugin `<plugin>`) and
  perform its **<operation>** operation… do not restate them."). A
  command never contains procedure; the skill reference does.
- **Marketplace / lockstep**: every plugin is registered in
  `.claude-plugin/marketplace.json`; ALL plugin packages and the
  marketplace share one version (currently 4.1.0) and are bumped
  together, only via a version-packages PR merged to main (triggers
  `.github/workflows/release.yml`). Never publish or push directly.
- **Plan documents**: `<dir>/<feature>.md` (parent) +
  `<dir>/<feature>/NN-name.md` (subplans), where `<dir>` is `plans/`
  or a category dir like `plans/agentic/` — resolve plans by
  frontmatter, not directory depth. Frontmatter: `status`
  (pending-approval → approved for parents; pending → in-progress →
  complete for subplans), `parallel_group` (subplans in the same group
  may run concurrently; groups execute in ascending order),
  `dependencies.plans/files`, `skills_to_consult`, `validation` (set
  each key to passed/failed as you verify). `plans/old/` is archive —
  never execute from it.
- **Workflow concepts (Track B)**: a *step* is a slash command with a
  machine-readable result written to the run journal
  (`plans/<plan>/runs/<run-id>-<step>.yaml`); a *gate* is a named
  decision point with policy `manual | assisted | auto`; *L0–L3* are
  maturity levels (manual → assisted → local headless → cluster) —
  same steps at every level, only the driver changes. Definitions live
  in `workflows/*.yml`; semantics are owned by the workflow-guide
  skill (created in B02).
- **Git rules** (from AGENTS.md): no feature branches or pushes unless
  the flow explicitly calls for them; the sanctioned flow is
  `/xonovex-workflow:plan-worktree-create` (branch + worktree +
  `branch.<branch>.plan` config) → work → `plan-worktree-merge`.
  Conventional commits. Release via PR only.

## Execution protocol (every subplan, human or agent)

1. Confirm the parent plan is `approved` and every entry in the
   subplan's `dependencies.plans` is `complete`; confirm external
   prerequisites in its Dependencies section (e.g. B05 → check Track S
   status; D → check Track C status).
2. Read: this roadmap → parent plan (Decision Log!) → the subplan →
   each `skills_to_consult` skill.
3. Create the worktree: `/xonovex-workflow:plan-worktree-create
   plans/<feature>/<subplan>.md`, `cd` into it, then
   `/xonovex-workflow:plan-continue`.
4. Execute tasks in order. Re-verify every file/line anchor before
   editing (drift rule). If reality contradicts the subplan, STOP and
   update the plan first (`/xonovex-workflow:plan-update`) — never
   improvise silently.
5. Run the subplan's Validation Steps; record results in its
   `validation` frontmatter; set `status: complete`; run
   `/xonovex-workflow:plan-update` so the parent rolls up.
6. Merge via `/xonovex-workflow:plan-worktree-merge`; groups touching
   plugin packages end with a version-packages release PR.

## Status snapshot (2026-07-03)

| Plan | Status | Next action |
|---|---|---|
| A platform-workflow-improvement (+6 subplans) | approved / pending | start group 1 (A01, A02, A05) |
| B agentic-workflow-evolution (+7 subplans) | approved / pending | starts after A04 merges |
| S agent-security-hardening (+7 subplans) | approved | in flight; B05 waits on S01+S03 |
| C drodan-multi-cluster-rollout (+9 subplans) | approved (drodan) | independent until M5 |
| D agentic-workflow-runtime | not created (by design) | create at M5 in drodan repo |
