---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-3
status: pending
dependencies:
  plans:
    [
      plans/hardening-branch-migration/subplan-02-infra-config-slice.md,
      plans/hardening-branch-migration/subplan-04-script-validation-slice.md,
    ]
  files: [packages/shared, packages/agent]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - moon-guide
  - typescript-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 05: Shared & Agent Slice

## Objective

Land the shared libraries (`packages/shared`: shared-core, shared-core-go,
shared-agent-go — 34 files) and the agent tooling (`packages/agent`:
agent-cli-go, agent-operator-go — 196 files) in one PR, respecting the
`config -> shared -> agent` dependency chain.

## Tasks

1. **Create the slice branch**:
   `git checkout -b migrate/shared-agent origin/main`
2. **Checkout donor paths**:
   `git checkout composable-workflow-platform-hardening -- packages/shared packages/agent`
3. **Audit added/removed files**: `git status --porcelain` — confirm every
   addition is intentional branch work and nothing deleted on main reappears
   (no mapped removals expected in these packages, but verify).
4. **Verify main-side intents survive** (parent plan intent map):
   - `022af353` — generated coverage output still ignored in agent-cli-go
   - `a2b8e0f4` — agent docs' egress and governance claims still match the code
     after the branch's doc changes (re-read the relevant docs against the
     migrated code, not against the branch's assumptions)
   - `9dfb7522` — prettier formatting holds for the shared/agent files it touched
5. **Regenerate lockfiles**: `npm install` for any TS package.json changes;
   `go mod tidy` in each Go module if go.mod/go.sum changed.
6. **Run validation**, commit, open the PR.

## Validation Steps

- `npx moon run :typecheck :lint :build :test` for projects under
  `packages/shared` and `packages/agent`
  (`moon query projects --tags "go|shared|agent"` to confirm ids)
- Go: module tests green, `go vet` clean via the moon go task templates
- Validators from subplan 04 stay green workspace-wide

## Success Criteria

- [ ] All three mapped main-commit intents verified
- [ ] Go and TS builds/tests green
- [ ] No resurrected or accidentally-dropped files
- [ ] Lockfiles/go.sum regenerated, not copied

## Files Modified/Created

- `packages/shared/**`, `packages/agent/**`, lockfiles (regenerated)

## Dependencies

Subplans 02 and 04 (task templates + validators on main first).

## Estimated Duration

3-5 hours.
