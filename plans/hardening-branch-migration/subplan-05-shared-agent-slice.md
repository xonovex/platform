---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-3
status: complete
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
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 05: Shared & Agent Slice

## Objective

Land the shared libraries (`packages/shared`: shared-core, shared-core-go,
shared-agent-go — 34 files) and the agent tooling (`packages/agent`:
agent-cli-go, agent-operator-go — 196 files) in one PR, respecting the
`config -> shared -> agent` dependency chain.

## Gate Inherited from Subplan 02

Subplan 02 removed `coverage` from `.moon/tasks/tag-go.yml`'s `ci-check` deps.
The template's `GO_COVERAGE_MIN` floor is 55 and `agent-cli-go` covers 44.8% on
main, while the donor's own `agent-cli-go/moon.yml` sets 81 because its tests
arrive with this slice. Once `packages/agent` is migrated, return `coverage` to
that `ci-check` deps list, carry each Go project's `GO_COVERAGE_MIN` override
from the donor, and confirm `npx moon run :ci-check` stays green.

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

- [x] All three mapped main-commit intents verified — see below
- [x] Go and TS builds/tests green — `:ci-check --force`, 790 tasks, exit 0,
      nothing cached, with the go coverage gate back in the closure
- [x] No resurrected files; three donor removals the checkout missed were
      applied — see below
- [x] Lockfile and every `go.sum` regenerated via `npm install` and `go mod tidy`

## Intent Verification

- `022af353` — generated coverage output still ignored: `coverage.out` resolves
  to `.gitignore`'s `*.out` rule. Present.
- `9dfb7522` — prettier pass: **regressed and repaired**. The donor left
  `shared-agent-go/README.md` and `agent-operator-go/README.md` unformatted,
  both of which main had formatted. Reformatted, along with
  `agent-operator-go-docker/package.json`, which main had already missed.
- `a2b8e0f4` — egress and governance claims match the code: **superseded, not
  dropped**. The donor rewrites `packages/agent/AGENTS.md` entirely, so the
  overclaims that commit corrected (a kill-switch, anomaly detection and an AI
  bill-of-materials described as existing) are simply absent rather than
  restated. The donor's replacement claims were checked against the migrated
  code: `proxy` is documented as reserved and failing closed, and
  `shared/network.go` defines `ErrProxyEnforcementUnavailable` to that effect.

## Donor Removals the Checkout Missed

`git checkout <branch> -- <path>` adds and modifies but never deletes, so a
file the donor removed survives on `main` unless it is removed explicitly. All
three predate the merge base, so each is a deliberate donor removal rather than
newer work on `main`:

| Path                                                       | Donor commit | Why it had to go                                                                                                                                                          |
| ---------------------------------------------------------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `agent-cli-go/internal/network/proxy/`                     | `861264a0`   | Its only surviving import still named the old `packages/cli/agent-cli-go` module path, which no longer resolves after the donor renamed the module to match its directory |
| `shared-core-go/pkg/scriptlib/`                            | `7ab28721`   | No tests and no consumers, so the per-package coverage check failed with "declares no coverage floor"                                                                     |
| `agent-operator-go/config/samples/agentconfig_sample.yaml` | `73c70ca8`   | `AgentConfig` was replaced by the four-concern resource model, so the sample no longer decodes against the API scheme                                                     |

## Findings

- The donor commits `os` and `cpu` onto the five `agent-cli-go-<platform>`
  packages **and** drops them from the root `workspaces` list. The two changes
  are inseparable: with the platform fields committed but the packages still
  globbed in as workspace members, `npm install` fails with `EBADPLATFORM` on
  any host that is not that platform. Moon's project discovery is independent of
  npm workspaces, so all five remain in the task graph.
- `format-check` does not cover markdown or JSON in Go- and docker-tagged
  projects — the go tag checks `gofmt` only. That is how two unformatted READMEs
  passed `ci-check` while a direct prettier run flagged them. Worth closing
  repo-wide, but it is a change to `.moon/tasks`, not to this slice.

## Files Modified/Created

- `packages/shared/**`, `packages/agent/**`, lockfiles (regenerated)

## Dependencies

Subplans 02 and 04 (task templates + validators on main first).

## Estimated Duration

3-5 hours.
