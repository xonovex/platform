---
type: plan
has_subplans: false
parent_plan: environment-hardening
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
  - packages/agent/agent-operator-go/moon.yml
  - packages/agent/agent-operator-go/README.md
  - .github/workflows/e2e.yml
  - plans/composable-workflow-phases/VALIDATION.txt (shared with siblings — own section only)
skills_to_consult:
- moon-guide
- github-guide
- shell-scripting-guide
- kubernetes-guide
- testing-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 01: E2E Remediation and CI Preparation

## Objective

Make the operator's envtest integration suite self-sufficient, wire the missing fourth Kind suite into moon, author the dormant CI workflow that will run all of it whenever the owner pushes, and time-box a host-runtime investigation — recording every disposition honestly.

## Tasks

1. **Fix `go-test-integration` wiring.** In `packages/agent/agent-operator-go/moon.yml`, convert the task from `command:` list to `script:` form so it wires its own environment:
   ```yaml
   go-test-integration:
     script: KUBEBUILDER_ASSETS=$(setup-envtest use -i -p path) go test -tags=integration -v -timeout=300s ./test/integration/
   ```
   preserving the existing `deps`, `inputs`, and `cache: false`. Note `setup-envtest` comes from the nix devshell (`nix/k8s.nix`) — document that requirement in the task-adjacent README section.
2. **Add the missing `go-test-e2e-coco` task** mirroring its three siblings (`-tags=e2e_coco`, per the operator AGENTS.md), so all four Kind suites are moon-visible.
3. **Author the dormant CI workflow** `.github/workflows/e2e.yml`: `workflow_dispatch` trigger (plus optional `schedule`), hosted ubuntu runner, nix devshell setup matching `ci.yml`'s pattern, jobs for `go-test-integration` and the four e2e suites, and per-suite skipped-vs-executed test counts surfaced in the job summary (kata/coco self-skip in unprivileged Kind — a pass-by-skip must be visible). A header comment states: dormant until first push, first-run expectations, and known suite-skip behavior.
4. **Validate the workflow locally**: `actionlint` (available via nix or npx) plus YAML parse; run the four moon tasks locally expecting the three runnable ones to fail only with the documented runc `setns` error and `go-test-e2e-coco` to behave identically (proving the task wiring, not the suite).
5. **Time-boxed host investigation (≤ half a day).** Reproduce the `setns: exec: already started` failure minimally (`docker exec --privileged` against a Kind node), check docker/runc versions against known kernel-7.x issues, try the one-or-two cheapest mitigations (runc update if trivially available, Kind node-image variant). Whatever the outcome, write it up in `packages/agent/agent-operator-go/README.md` under the e2e section: root cause or best hypothesis, what was tried, and the working alternative (CI workflow).
6. **Update VALIDATION.txt (own section)**: replace the host-exclusion caveat with the current disposition — four suites moon-wired, integration self-sufficient, CI coverage prepared-but-dormant, no-push deferral with reactivation conditions. Full `:ci-check` gate; one conventional commit.

## Validation Steps

- `nix develop --command npx moon run agent-operator-go:go-test-integration` green **without** any exported `KUBEBUILDER_ASSETS`.
- `moon query tasks` shows `go-test-e2e-coco`; running it fails only with the documented runc error (same signature as siblings).
- `actionlint .github/workflows/e2e.yml` clean; workflow header documents dormancy.
- Full `:ci-check` green; nothing pushed (remote still `166c4f26`).

## Success Criteria

- [ ] Integration task self-sufficient (no manual env), `script:` form, nix requirement documented
- [ ] All four e2e suites have moon tasks with identical shape
- [ ] `e2e.yml` exists, parses clean, surfaces skip-vs-run counts, states dormancy and first-run expectations
- [ ] Host investigation written up in the operator README with outcome and time-box honored
- [ ] VALIDATION.txt e2e section reflects the new disposition and the owner's no-push deferral
- [ ] Full gate green; one conventional commit; remote untouched

## Files Modified/Created

- Modified: `packages/agent/agent-operator-go/{moon.yml,README.md}`, `plans/composable-workflow-phases/VALIDATION.txt`
- Created: `.github/workflows/e2e.yml`

## Dependencies

None (group 1). Shares only VALIDATION.txt with siblings — each edits its own section; single-worktree execution serializes writes.

## Estimated Duration

0.5–1 day (half of it the time-boxed investigation).
