---
type: plan
has_subplans: true
status: in_progress
dependencies:
  plans: []
  subplans:
    - plans/skill-catalog-stabilization/subplan-01-scenario-freeze-and-contract.md
    - plans/skill-catalog-stabilization/subplan-02-catalog-cull.md
    - plans/skill-catalog-stabilization/subplan-03-anti-drift-lints-warn.md
    - plans/skill-catalog-stabilization/subplan-04-workflow-core-simplification.md
    - plans/skill-catalog-stabilization/subplan-05-process-skill-grounding.md
    - plans/skill-catalog-stabilization/subplan-06-anti-drift-enforce.md
    - plans/skill-catalog-stabilization/subplan-07-release-and-dogfood.md
parallel_groups:
  group-1: [subplan-01-scenario-freeze-and-contract]
  group-2: [subplan-02-catalog-cull, subplan-03-anti-drift-lints-warn]
  group-3:
    [
      subplan-04-workflow-core-simplification,
      subplan-05-process-skill-grounding,
    ]
  group-4: [subplan-06-anti-drift-enforce]
  group-5: [subplan-07-release-and-dogfood]
proposed_subplans:
  - scenario-freeze-and-contract
  - catalog-cull
  - anti-drift-lints-warn
  - workflow-core-simplification
  - process-skill-grounding
  - anti-drift-enforce
  - release-and-dogfood
skills_to_consult:
  - skill-guide
  - command-guide
  - workflow-guide
  - instruction-guide
  - plan-guide
  - reflect-guide
  - typescript-guide
  - vitest-guide
  - moon-guide
  - versioning-guide
  - npm-guide
  - git-guide
research_sources:
  documentation:
    - packages/skill/skill-workflow/workflow-guide/references/handoffs.md
    - packages/skill/skill-workflow/workflow-guide/references/sdlc.md
    - packages/command/command-workflow/commands/
    - .moon/tasks/tag-skill.yml
    - .moon/tasks/tag-command.yml
    - packages/agent/agent-dispatcher-go/internal/domain/domain.go
  versions: {}
---

# Plan: Stabilize the Skill and Command Catalog

Approved by the author on 2026-07-25 (status is descriptive; the acceptance was the
authorizing act). Revision 3 content, restructured to the plan-guide document shape.
Branch context: `composable-workflow-platform-hardening`.

## Overview

Bring `packages/skill` and `packages/command` to a stable state: a settled two-axis
architecture, an internally consistent catalog where every artifact traces to a real
usage scenario, and quality gates that make abstraction drift mechanically
impossible to merge. Future changes become driven by evals and transcripts instead
of open-ended improvement passes.

## Goals

- Freeze a one-page ground truth (scenarios + architecture contract) that all
  catalog content must trace to.
- Cut the catalog from 100 to 72 skills by deleting everything without a grounding
  scenario.
- Rewrite the workflow core to the minimal cold-boundary handoff, a single
  governance source, and a safety-core flag surface.
- Ground surviving process skills in checkable artifacts with output evals.
- Enforce anti-drift lints (ratchet + caps, vocabulary ownership, duplication).
- Release so the installed catalog matches the repo, then prove stability by
  dogfooding to an evidence-based exit criterion.

## Current State

- 100 skill packages in `packages/skill`; 12 workflow commands in
  `command-workflow` plus utility commands in `command-utility`.
- Gate infrastructure exists and runs via moon tags: `skill-validate`,
  `skill-audit-sources`, `skill-eval-triggers`, `skill-eval-outputs`,
  `command-validate`, backed by `packages/script/script-moon-*` packages, with
  per-skill `eval-queries.json` (train/test splits, pinned evaluator models).
- Release flows through a reviewed version-packages PR;
  `.github/workflows/release.yml` publishes on merge to `main`.
- The installed catalog lags the repo: sessions still expose the pre-refactor
  command set, and installed commands (`/plan-decide`, `/plan-accept`) delegate to
  operations the repo's `plan-guide` no longer defines.
- `packages/agent` is healthy and out of scope; the dispatcher
  (`agent-dispatcher-go`) is on hold with a provider-native effect vocabulary
  (verified in `internal/domain/domain.go`).

## Research Findings

### Spiral symptoms (measured)

- `handoffs.md` demands 16-field context entries (sha256 digests, version
  counters); `sdlc.md` is a 22-phase/20-role matrix citing 14 guides; the
  untrusted-data invariant is restated across 40 files; 21 distinct flags exist
  across 12 commands; `decide.md` is predominantly caveats.
- Word counts: SKILL.md p50 438 / p75 633 / p90 873 / max 1786; commands p50 140 /
  max 224; references p50 213 / p90 604 / max 2046.

### Cull blast radius (measured)

- `sdlc.md` holds 14 references to cut guides (absorbed by its rewrite); ~20
  surviving `eval-queries.json` files mention cut guides in boundary queries; one
  prose file (`skill-user-stories/.../splitting-flowchart.md` → `fdd-guide`). No
  surviving manifest dependency on any cut skill.

### Original design intent (from the author)

Handoffs are placement into native systems (file, ticket, comment, PR) carrying
enough context for referencing sessions — canonically: coding-session decisions
enabling a later session's PR description and anchored inline comments. The
16-field protocol was another agent's over-generalization of this intent.

### Decision log

1. Ground-truth scenarios: all four project families — Xonovex platform,
   Drodan/CruiseReviews product, native/game-engine, infra/ops.
2. Cull fate: delete outright; git history is the archive; resurrection requires a
   scenario plus an output eval.
3. Size budgets: ratchet (per-file ceiling in a checked-in budget manifest, seeded
   after the correction pass; growth requires an explicit bump in the same PR) plus
   absolute p90 caps for new files: SKILL.md ≤ ~900 words, references ≤ ~650,
   commands ≤ ~250.
4. Flag surface: shrink to the safety core (~8 flags): `subject`, `--request`,
   `--context`, revision pinning, `--effect` / `--idempotency-key` on effectful
   operations; taxonomy flags (`--criterion`, `--method`, `--perspective`,
   `--option`, `--outcome`, `--evidence`) removed.
5. Dogfood exit: coverage + quiet period — all 12 operations used across ≥ 2
   families, plus two consecutive weeks with zero transcript-evidenced failures.
6. Cull table: 28 deletions (below); `copilot`, `kiro`, `opencode` confirmed keeps.
7. Dispatcher: on hold, in-tree, out of scope.

### Cull table — 28 deletions

- Role/process sweep (13): `product-discovery`, `product-analytics`, `ux-research`,
  `ux-design`, `release-readiness`, `operational-readiness`, `incident-response`,
  `test-strategy`, `exploratory-testing`, `security-testing`,
  `architecture-evaluation`, `threat-modeling`, `fdd`
- Provider/tooling sweep (7): `atlassian`, `azure-devops`, `bitbucket`, `bitrise`,
  `figma`, `datadog`, `aws`
- No grounding family (8): `android-analytics`, `android-wcag`, `strudel`,
  `expressjs`, `presentation`, `remotion`, `motion-react`, `adr`

## Proposed Approach

### Architecture contract (subplan 1 produces the authoritative one-page version)

1. **Two axes.** Skills are knowledge-first and entry-point-agnostic. Entry points
   (interactive sessions, workflow commands; future: crons, observation alerts,
   dispatcher phase 4) decide when an agent wakes and with what task; the catalog
   decides what it knows. Stream-shaped work gets its own entry point; it is never
   forced through the operations.
2. **Warm/cold transitions.** Operations chain freely within a session. A handoff
   exists only at a cold boundary (session or role change) and consists of
   placement into a native system — a file being one native system among several.
3. **Context graph.** Native artifacts (tickets, kanban cards, comments, files,
   PRs) carry context as nodes; links are deferred context resolved on demand by
   the skill matching the linked service; phases read and write across systems.
   Unresolvable links degrade visibly, never silently.
4. **Minimal handoff.** Subject + revision, what was done, decisions
   (what / why / where — code anchors), references and links, open issues. Regular
   headings (script-producible and consumable), no digests, versions, or audience
   taxonomies.
5. **Effects.** Fetched provider content informs work but never becomes
   instructions or authority (single governance reference). Effect sets extend only
   by declared team convention (team AGENTS.md or opinionated overlay skill) —
   never by inference; convention effects are overridable and always reported.
6. **Executor-agnostic operations.** Defined by inputs, effect boundary, and
   handoff — not by who executes. Agent → script+LLM → script migration is expected
   and interface-preserving (`workspace-*` operations are the anticipated first
   candidates).
7. **Execute is positive, not residual.** Execute carries out previously specified
   work and expects that antecedent. Requests fitting no operation get no command —
   freeform sessions are the sanctioned outside.

### Components

1. Scenario freeze and contract — the one-page ground truth (subplan 1).
2. Catalog cull — execute the 28-deletion table with reference sweep (subplan 2).
3. Anti-drift gates — lints in warn mode first (subplan 3), seeded and enforced
   after the rewrites (subplan 6).
4. Workflow-core simplification — handoff, governance, sdlc, flags, execute
   (subplan 4).
5. Process-skill grounding — artifact anchoring plus output evals including the
   golden end-to-end scenario (subplan 5).
6. Release and dogfood — close the installed gap, prove stability (subplan 7).

### Exclusions and reserved extension points

- `packages/agent` untouched; dispatcher on hold; no phase-4 work.
- Documented but not built: the team-convention layer and the script / script+LLM
  executor migration.
- No new operations, commands, or skills.

### Constraints

- Repo rules: no feature branches or pushes unless asked; release only via reviewed
  version-packages PR; conventional commits; delete rather than deprecate.
- Rewrites are bounded by the budgets; wording changes that are neither deletions
  nor concretizations are out of scope.
- The dispatcher's Go model is not a constraint on the handoff format.

## Risk Assessment

- Over-culling — mitigated by git history and the scenario+eval resurrection bar; a
  live unresolvable link is itself the resurrection signal for service skills.
- Contract document becoming the new spiral surface — mitigated by its own one-page
  ratchet and the vocabulary lint applying to it.
- Execute catch-all — mitigated fourfold: positive definition, freeform escape
  hatch, antecedent expectation, dogfood metrics feeding boundary evals.
- Eval subjectivity for process artifacts — mitigated by pinned evaluators and
  rubric-based golden scenarios testing observable outputs, not taste.
- Dogfood discipline decay — partially mitigated by the ratchet lint making silent
  growth fail CI.

## Proposed Child Plans

| Group | Subplan                                   | Depends on                               |
| ----- | ----------------------------------------- | ---------------------------------------- |
| 1     | `subplan-01-scenario-freeze-and-contract` | —                                        |
| 2     | `subplan-02-catalog-cull`                 | 01                                       |
| 2     | `subplan-03-anti-drift-lints-warn`        | 01 (parallel with 02; disjoint files)    |
| 3     | `subplan-04-workflow-core-simplification` | 02 (parallel with 05; disjoint packages) |
| 3     | `subplan-05-process-skill-grounding`      | 02 (parallel with 04)                    |
| 4     | `subplan-06-anti-drift-enforce`           | 03, 04, 05                               |
| 5     | `subplan-07-release-and-dogfood`          | 02, 04, 05, 06                           |

## Success Criteria

- [x] Contract document exists, ≤ 1 page (649 words), seven clauses; every surviving
      artifact maps to a frozen scenario
- [~] Catalog at 72 skills; zero references to cut guides (grep-verifiable);
  trigger evals green catalog-wide — catalog and grep done; catalog-wide trigger
  evals still unrun (hours of wall clock plus API spend)
- [x] Governance normative in exactly 1 file (from 40); `handoffs.md` 5 field groups
      with `file:line` anchors and no digest/version/audience machinery; command flag
      count = 8; `execute.md` contains the positive definition
- [x] Golden end-to-end output eval exists and passes (0.833, gate PASS); zero
      keep-tier process skills without output evals
- [x] Anti-drift lints enforcing; budget manifest seeded (555 files); full `ci-check`
      green across skill, command, and typescript-script tags (522 tasks)
- [ ] Release published; fresh-session installed catalog matches repo exactly
- [ ] Definition of Done: all above, plus dogfood exit met (12/12 operations,
      ≥ 2 families, 2 quiet weeks) with every in-window change citing an eval or
      transcript

## Progress

Subplans 1–6 are complete; subplan 7 (release and dogfood) is pending and is the only
remaining child. Each subplan document carries its own execution record, including the
deviations taken and why.

Measured outcome so far: 100 → 72 skills, workflow-core owned scope 8890 → 7008 words,
`sdlc.md` 1345 → 417, `context-forwarding.md` 849 → 270, command flags 21 → 8,
governance restatements 40 → 1, drift lints enforcing with 0 findings over 555 files.

Two catalog-wide eval sweeps remain unrun — trigger and output — because each is hours
of wall clock plus API spend rather than because anything blocks them; the harness
defects that did block them are fixed.

## Estimated Effort

- Subplans 1–6: roughly 8–12 working sessions (groups 2 and 3 parallelizable).
- Subplan 7: 1 session for the release, then an evidence-bounded dogfood window
  (minimum ~3 calendar weeks).
