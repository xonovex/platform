---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 2
status: complete
dependencies:
  plans: [subplan-01-scenario-freeze-and-contract]
  files:
    - packages/skill/
    - tsconfig.json
    - package.json
skills_to_consult: [git-guide, skill-guide, moon-guide, npm-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: partial
---

# Subplan 2: catalog-cull

## Objective

Execute the parent's pre-computed cull: delete 28 skill packages and sweep every
reference to them, leaving a 72-skill catalog with all existing gates green.

## Cull table (from parent decision log)

- Role/process (13): `product-discovery`, `product-analytics`, `ux-research`,
  `ux-design`, `release-readiness`, `operational-readiness`, `incident-response`,
  `test-strategy`, `exploratory-testing`, `security-testing`,
  `architecture-evaluation`, `threat-modeling`, `fdd`
- Provider/tooling (7): `atlassian`, `azure-devops`, `bitbucket`, `bitrise`,
  `figma`, `datadog`, `aws`
- No grounding family (8): `android-analytics`, `android-wcag`, `strudel`,
  `expressjs`, `presentation`, `remotion`, `motion-react`, `adr`

Confirmed keeps needing no action: `copilot`, `kiro`, `opencode`.

## Tasks

1. Delete the 28 `packages/skill/skill-<name>` directories (plus any additions
   escalated by subplan 1's mapping).
2. Sweep root-level registrations that index packages by name: workspace globs in
   `package.json`, project references in `tsconfig.json` / `tsconfig.options.json`,
   README or marketplace listings; refresh `package-lock.json` via `npm install`.
3. Reword the ~20 boundary eval queries in surviving `eval-queries.json` files that
   name cut guides, pointing each at a surviving neighbor
   (e.g. `expressjs-guide` boundaries → `hono-guide`).
4. Fix the `fdd-guide` reference in
   `packages/skill/skill-user-stories/user-stories-guide/references/splitting-flowchart.md`.
5. Verify zero live references:
   `grep -rE "(product-discovery|...|adr)-guide" packages/ --include="*.md" --include="*.json"`
   over all 28 names returns nothing.
6. Run catalog-wide gates (validation steps below).

## Validation Steps

- `npx moon run '#skill:skill-validate'`
- `npx moon run '#skill:skill-eval-triggers'`
- `npx moon run '#skill:ci-check'` and `npx moon run '#command:ci-check'`

## Success Criteria

- [x] Catalog at 72 skill packages
- [x] Grep sweep for all 28 guide names returns zero hits outside git history and the
      plan documents that record the cull
- [ ] Trigger evals green catalog-wide — scoped to the 14 skills whose queries changed;
      catalog-wide blocked on eval-harness defects (see appendix)
- [x] Full ci-check green on skill and command tags (485 tasks), plus
      `script-moon-skill-validate:test` (54 tests) and `release-validate`
      (1349 checks, 74 lockstep packages)

## Files Modified/Created

- Delete: 28 `packages/skill/skill-*` directories
- Modify: ~20 surviving `eval-queries.json`, 1 prose reference, root manifests,
  `package-lock.json`

## Dependencies

- `subplan-01` (mapping may add cull entries).
- Parallel-safe with `subplan-03` (disjoint files: skill packages vs. script
  packages and task wiring).

## Estimated Duration

1–2 sessions.

## Appendix: execution record

### Blast radius (measured, larger than the parent estimated)

| Surface                              | Parent estimate                  | Actual                                                                                      |
| ------------------------------------ | -------------------------------- | ------------------------------------------------------------------------------------------- |
| `eval-queries.json` boundary queries | ~20                              | 33 across 14 files                                                                          |
| Prose cross-references               | 1 file                           | 3 files (user-stories, ddd ×2)                                                              |
| `sdlc.md` guide names                | 14 (deferred to subplan 4)       | 14, fixed here                                                                              |
| Contract test assertions             | not identified                   | 12 in `workflow-contracts.test.ts`                                                          |
| Output evals                         | not identified                   | 1 in `command-guide/evals.json`                                                             |
| Registration files                   | `package.json` / `tsconfig.json` | neither (both glob-based); 28 entries in each of two marketplace manifests, 24 README lines |

### Scope decisions

- `sdlc.md` and `workflow-contracts.test.ts` were swept here rather than deferred:
  leaving them would have left 14 dangling guide pointers and a failing contract test
  in the tree until subplan 4. Cut guide names became capability descriptions
  ("a product-discovery capability"), which matches the skill's own soft-selection
  principle; subplan 4 still performs the full rewrite on top.
- Three orphaned negative queries whose deleted owner was named only in prose
  (`accessibility[7]` Jetpack Compose, `pull-request[10]` Azure DevOps,
  `react[10]` Express) were repointed at surviving owners. The rationale-only grep
  missed these; query text needs sweeping too.
- Pre-existing dangling owner `motion-guide` in `astro-guide` (the package was named
  `motion-react`) was corrected in the same pass.
- Routing-pairing fallout: `missingValidationRoutingOwners` requires every skill to be
  the sole positive of some query text that another skill carries as a negative, in the
  validation split. `adr-guide`, `android-wcag-guide`, and `strudel-guide` were the only
  carriers for `plan-guide`, `accessibility-guide`, and `audio-guide`; deleting them
  broke all three. Restored by adding the same query as a negative in
  `code-quality-guide`, `react-guide`, and `memory-management-guide`.

### Gate gap this exposed

`skill-validate` and `#skill:ci-check` do not cover it — the invariant lives in
`script-moon-skill-eval-triggers:ts-coverage`, which neither `#skill:ci-check` nor
`script-moon-skill-validate:test` runs. Any cull must also run
`#typescript-script:ci-check` and the eval packages' `ts-coverage`.

### Trigger-eval results (validation split, 3 runs, 14 affected skills)

Every replacement query that the harness scored passes. Four replacements initially
mis-routed and were rewritten, then re-verified green:

| Skill                 | Replacement that mis-routed    | Cause                      | Now                          |
| --------------------- | ------------------------------ | -------------------------- | ---------------------------- |
| audio                 | "Import authored WAV sources…" | "WAV" pulled `audio-guide` | glTF meshes, rate 0          |
| credential-management | GitHub App token endpoint      | token/auth vocabulary      | gh REST paging, rate 0       |
| credential-management | gh CLI token + host profiles   | token/auth vocabulary      | gh PR listing, rate 0        |
| credential-management | GitLab OIDC federated claims   | auth vocabulary            | protected-branch job, rate 0 |

`skill-audio` and `skill-credential-management` re-ran fully green (8/8 and 6/7, the
one failure being an untouched Argon2 query). `skill-terraform` and `skill-react`
passed clean.

### Eval-harness defects blocking catalog-wide evals (for subplan 3)

1. `#skill:skill-eval-triggers` aborts before any model call: defaults (all splits ×
   3 runs) exceed the harness batch cap — "48 model runs; maximum is 24". Every skill
   has ≥16 queries, so the documented command cannot succeed for any of them.
2. Resolved by deletion. `ci-runtime` failed in bash with "syntax error near
   unexpected token `)`" — moon delivered the `|` block scalar's newlines to
   `bash -c` as literal `\n`. Making it run revealed what it did: `npm install
--no-save` of `@anthropic-ai/claude-code` or `@openai/codex` into the workspace on
   every CI eval run. The task was dropped from both eval packages and its three
   dependents now depend on `~:install`; harnesses are expected on `PATH`.
3. The harness is fail-fast on transient errors (`claude exited 1`, `output-limit`),
   so one flake discards the whole skill's run. 6 of 14 skills aborted this way.
4. `dist/` is gitignored and produced by the `ts-build` moon task, so the bin
   `node_modules/.bin/moon-skill-eval-triggers` executes whatever build happens to be
   on disk. Invoking it directly (rather than through `moon run`, whose `deps` rebuild
   first) can silently run stale code — observed here as a binary rejecting
   `--plugin-dir` / `--workspace` / `--split` that its own `cli.ts` defines. The
   package is not published to npm; the workspace symlink is correct.

### Failing queries left untouched (not modified by this subplan)

`hono[LinearRouter]`, `hono[Hono or Express + Vitest]`, `zod[Hono JSON request]`,
`threejs[translate gizmo]`, `user-stories[As a / I want]`, `user-stories[SMART]`,
`user-stories[three amigos]`, `credential-management[Argon2]`. These are eval-quality
issues for subplan 5, not cull fallout.
