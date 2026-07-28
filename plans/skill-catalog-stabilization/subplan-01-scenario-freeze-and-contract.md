---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/skill/skill-workflow/workflow-guide/SKILL.md
    - packages/skill/skill-workflow/workflow-guide/references/contract.md
skills_to_consult: [skill-guide, workflow-guide, instruction-guide]
validation:
  type_check: not_applicable
  lint: pass
  build: pass
  tests: not_applicable
  integration: pass
---

# Subplan 1: scenario-freeze-and-contract

## Objective

Produce the authoritative one-page ground truth — frozen scenarios plus the seven
architecture-contract clauses — that every other subplan cites and every future
catalog change must trace to.

## Context carried from parent

- Frozen families: Xonovex platform, Drodan/CruiseReviews product,
  native/game-engine, infra/ops.
- Seven clauses (parent "Architecture Contract"): two axes, warm/cold transitions,
  context graph, minimal handoff, effects and declared conventions,
  executor-agnostic operations, positive execute definition.
- The contract is itself ratcheted: one page forever; the vocabulary lint applies
  to it.

## Tasks

1. Enumerate concrete agent-driven scenarios per frozen family — from actual
   practice, not aspiration; 3–6 per family. Record each as
   `<family>: <trigger> → <operations/skills used> → <artifact produced>`.
2. Write the seven contract clauses in final normative wording, each ≤ 4 sentences,
   defining every term of art it uses.
3. Create `packages/skill/skill-workflow/workflow-guide/references/contract.md`
   containing scenarios + clauses; add it to the Progressive Disclosure section of
   `packages/skill/skill-workflow/workflow-guide/SKILL.md`.
4. Build the mapping table: every keep-tier skill (72) and all 12 commands mapped to
   ≥ 1 scenario. Escalate any unmappable artifact to the parent's cull table (feeds
   subplan 2). Store the table in this subplan file under an appendix heading.
5. Verify the one-page budget (~650 words) and self-contained vocabulary; trim until
   both hold.

## Validation Steps

- `npx moon run skill-workflow:skill-validate`
- `npx moon run skill-workflow:format-check`

## Success Criteria

- [x] `contract.md` exists at the stated path, ≤ 1 page (649 words)
- [x] All 72 keep-tier skills and 12 commands mapped; zero unmappable keeps
- [x] No clause uses a term the document does not define
- [x] Skill validation and format checks green for `skill-workflow`

## Files Modified/Created

- Create: `packages/skill/skill-workflow/workflow-guide/references/contract.md`
- Modify: `packages/skill/skill-workflow/workflow-guide/SKILL.md`
- Modify: this subplan (mapping-table appendix)

## Dependencies

None — blocks all other subplans.

## Estimated Duration

1 session.

## Appendix: scenario mapping

Scenario ids are defined below. They record the practice the catalog was frozen
against; they are a planning record, not skill content, because the workflow guide
selects capabilities from installed routing descriptions rather than from a fixed
inventory. Every keep-tier skill and every workflow command maps to at least one; no
artifact was unmappable, so nothing was escalated to the parent cull table.

### Frozen scenarios

Each line reads `trigger -> operations and skills -> artifact`.

**Xonovex platform** — agent tooling.

- XP1 capability gap or repeated correction -> reflect, skill, command, instruction ->
  amended catalog artifact with evals
- XP2 multi-step change needs scoping -> plan, decide; architecture-pattern guides ->
  plan document with subplans
- XP3 package implementation -> execute, validate; typescript, vitest, moon -> source
  with green tasks
- XP4 change ready to land -> review, publish; git, pull-request, versioning, npm ->
  pull request and published release

**Drodan and CruiseReviews** — multilingual product and public sites.

- DR1 editorial or destination guide requested -> create, revise, publish; content,
  accessibility -> published multilingual entry
- DR2 increment needs slicing -> decide, validate; user-stories, bdd, tdd -> stories
  with agreed examples
- DR3 web surface built -> execute; astro, react, accessibility -> rendered page or
  component
- DR4 endpoint and its data added -> execute, validate; hono, zod, sql-postgresql ->
  endpoint with migration and tests

**Native and game engine** — C99 runtimes and editors.

- NG1 throughput-critical subsystem -> execute; data-oriented-design, ecs, c99,
  memory-management -> subsystem and build target
- NG2 rendering or editor feature -> execute; gpu-rendering, editor-viewport, imgui,
  node-graph -> viewport feature
- NG3 authored content or gameplay script -> execute; asset-pipeline, audio, lua ->
  pipeline stage and scripted behaviour
- NG4 new target or reported crash -> execute, abandon; cross-platform, debugging ->
  platform layer or minimal repro and fix

**Infrastructure and operations** — clusters, pipelines, repository state.

- IO1 service must run in a cluster -> create, publish; docker, kubernetes, terraform
  -> GitOps manifests
- IO2 automation or check must run -> execute, validate; github, gitlab, moon,
  shell-scripting -> pipeline definition and scripts
- IO3 machine credential stored, injected, or rotated -> execute;
  credential-management -> secret-manager-backed configuration
- IO4 parallel work isolated or integrated -> workspace create, merge, abandon,
  cleanup -> worktree lifecycle

### Keep-tier skills (72)

| Skill                 | Scenarios | Skill               | Scenarios     |
| --------------------- | --------- | ------------------- | ------------- |
| accessibility         | DR1, DR3  | lock-free           | NG1           |
| asset-pipeline        | NG3       | lua                 | NG3           |
| astro                 | DR3       | lua-opinionated     | NG3           |
| audio                 | NG3       | memory-management   | NG1, NG4      |
| bdd                   | DR2       | microkernel-pattern | XP2           |
| c99                   | NG1       | moon                | XP3, IO2      |
| c99-game-opinionated  | NG1       | node-graph          | NG2           |
| c99-opinionated       | NG1       | npm                 | XP4           |
| claude-code           | XP1       | oop                 | XP3           |
| cmake                 | NG1, NG4  | opencode            | XP1           |
| code-quality          | XP3       | orthogonal-pattern  | XP2           |
| code-review           | XP4       | pi                  | XP1           |
| codex                 | XP1       | plan                | XP2           |
| command               | XP1       | pull-request        | XP4           |
| connascence           | XP2, XP3  | python              | XP3, IO2      |
| content               | DR1       | react               | DR3           |
| copilot               | XP1       | reflect             | XP1           |
| credential-management | IO3       | shell-scripting     | IO2           |
| cross-platform        | NG4       | skill               | XP1           |
| data-model            | NG2       | sql-postgresql      | DR4           |
| data-oriented-design  | NG1       | tdd                 | XP3, DR2      |
| ddd                   | XP2, DR4  | terraform           | IO1           |
| debugging             | NG4       | testing             | XP3, DR2      |
| docker                | IO1       | threejs             | NG2           |
| ecs                   | NG1       | typescript          | XP3, DR3, DR4 |
| editor-viewport       | NG2       | typescript-to-lua   | NG3           |
| fp                    | XP3       | user-stories        | DR2           |
| game-networking       | NG3       | versioning          | XP4           |
| git                   | XP4, IO4  | vitest              | XP3           |
| github                | XP4, IO2  | workflow            | all           |
| gitlab                | XP4, IO2  | zod                 | XP3, DR4      |
| gpu-rendering         | NG2       |                     |               |
| gpu-rendering-vulkan  | NG2       |                     |               |
| hexagonal-pattern     | XP2, DR4  |                     |               |
| hono                  | DR4       |                     |               |
| hono-opinionated      | DR4       |                     |               |
| imgui                 | NG2       |                     |               |
| instruction           | XP1       |                     |               |
| kiro                  | XP1       |                     |               |
| kubernetes            | IO1       |                     |               |
| llmstxt               | DR1       |                     |               |

### Workflow commands (12)

| Command           | Scenarios              |
| ----------------- | ---------------------- |
| create            | XP1, DR1, IO1          |
| review            | XP4, DR2               |
| revise            | XP1, DR1               |
| decide            | XP2, DR2               |
| execute           | XP3, NG1-NG4, IO2, IO3 |
| validate          | XP3, DR2, DR4, IO2     |
| publish           | XP4, DR1, IO1          |
| abandon           | NG4                    |
| workspace-create  | IO4                    |
| workspace-merge   | IO4                    |
| workspace-abandon | IO4                    |
| workspace-cleanup | IO4                    |

### Scenario grounding

Families were grounded against working trees, not aspiration: `xonovex-platform`
(skills, commands, Go agents, moon, npm) for XP; `drodan-platform` packages
`cruisereviews-*`, `website-*` (astro, react, hono, zod, vitest) for DR;
`drodan-platform` packages `game-*` plus `vn` and `voxellandneo` (C99, CMake,
typescript-to-lua, three.js) for NG; `drodan-platform` packages `kubernetes-*`,
`terraform-*`, plus `drodan-cluster` and `drodan-infrastructure` for IO.
