---
type: plan
has_subplans: false
parent_plan: plans/agent-orthogonal-axis-reorg.md
parallel_group: 6
status: pending
dependencies:
  plans: [06-operator-axis-relocation.md]
  files:
    - packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go
    - packages/agent/agent-operator-go/internal/controller/
    - packages/agent/agent-operator-go/api/v1alpha1/zz_generated.deepcopy.go
    - packages/agent/agent-operator-go/config/rbac/role.yaml
    - packages/shared/shared-agent-go/pkg/policy/
skills_to_consult:
  - hexagonal-pattern-guide
  - microkernel-pattern-guide
  - kubernetes-guide
  - connascence-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Operator Policy Symmetry + Goldens + Codegen

## Objective

Route the operator's admission webhook through the same shared `pkg/policy.EnforcePolicy`
the CLI already calls, deleting the operator's duplicated admission algorithm so both
consumers share the policy only through the neutral `Capabilities` struct. Lock the
relocation and the `BuildJob`+`BuildWorkspaceJob` merge from changing pod output or
admission verdicts with golden tests, and reconcile the hand-maintained codegen (deepcopy)
and RBAC against the moved packages. Closes Phase 2.

## Tasks

1. **Capture baseline goldens BEFORE any rewire** — in
   `packages/agent/agent-operator-go/internal/builder/` (or the relocated per-axis dirs from
   subplan 06), add table-driven golden tests that marshal the produced `*batchv1.Job` /
   pod spec for representative `AgentRun` specs and compare against committed
   `testdata/*.golden.yaml`. Generate the goldens from the pre-merge code first so the
   merge is proven output-preserving.
   ```go
   func TestBuildJobGolden(t *testing.T) {
       for _, tc := range jobCases { // bwrap, docker, nix-provision, workspace-git
           got := mustMarshalYAML(t, BuildJob(tc.run))
           golden.Assert(t, got, tc.name+".golden.yaml") // -update regen flag
       }
   }
   ```

2. **Add admission-verdict goldens BEFORE the webhook rewire** — in
   `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook_test.go`, snapshot the
   accept/deny outcome (and denial reason string) for representative specs against the
   *current* `enforcePolicy`. These freeze behavior so the engine swap in Task 3 is proven
   verdict-preserving.
   ```go
   wantVerdicts := map[string]string{
       "pinned-ok": "", "unpinned-strict-deny": "provisioner not pinned",
       "egress-open-deny": "egress not restricted",
   }
   ```

3. **Rewire `webhook.enforcePolicy` onto the shared engine** — in
   `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go`, compute a
   `policy.Capabilities` from the CRD spec and call shared `policy.EnforcePolicy`; delete the
   operator's inline admission logic (connascence: the shared `Capabilities` struct is the
   only seam, not a copied algorithm).
   ```go
   caps := policy.Capabilities{
       Pinned:         spec.Provision.Pinned(),
       EgressRestricted: spec.Network.Mode == v1alpha1.NetworkProxy,
       KernelIsolated: spec.Isolation.KernelIsolated(),
   }
   if err := policy.EnforcePolicy(caps, spec.SandboxPolicy); err != nil {
       return admission.Denied(err.Error())
   }
   ```
   Confirm Task 2 goldens stay green.

4. **Regenerate deepcopy for the new typed enums** — for the typed CRD enums introduced in
   subplan 06 (e.g. `NetworkMode`), hand-update
   `packages/agent/agent-operator-go/api/v1alpha1/zz_generated.deepcopy.go` to match (string
   enum types need no deep copy beyond value assignment; verify any new slice/map fields get
   `DeepCopyInto` entries). controller-gen is hand-maintained on this toolchain — hand-check
   the file compiles and round-trips.

5. **Reconcile RBAC with the moved `+kubebuilder:rbac` markers** — after the package moves,
   collect every `+kubebuilder:rbac` marker and reconcile
   `packages/agent/agent-operator-go/config/rbac/role.yaml` so its rules exactly match
   (role.yaml is already drifted per the parent plan risk list). Remove rules no relocated
   marker grants; add any missing.
   ```bash
   grep -rn '+kubebuilder:rbac' packages/agent/agent-operator-go/internal | sort
   ```

6. **Add the cross-module axis-symmetry fitness test** — add a Go test (operator side, e.g.
   `internal/arch/symmetry_test.go`) asserting the top-level axis dir-name sets under
   `agent-cli-go/internal` and `agent-operator-go/internal` intersect on
   `{isolation, network, provision, workspace, harness, provider}`; fail if an axis is present
   in one consumer but absent in the other without an entry in a documented exceptions map.
   ```go
   want := []string{"isolation", "network", "provision", "workspace", "harness", "provider"}
   for _, axis := range want {
       if !hasDir(cliInternal, axis) || !hasDir(opInternal, axis) {
           if _, ok := documentedExceptions[axis]; !ok { t.Errorf("axis %q asymmetric", axis) }
       }
   }
   ```

7. **Full three-module validation + operator envtest** — run the moon typecheck/lint/build/test
   across all three modules plus the operator envtest/integration suite; confirm Job/pod and
   admission goldens are green (only the intended merge differs) and no RBAC drift remains.

## Validation Steps

- `cd packages/agent/agent-operator-go && go build ./...`
- `cd packages/agent/agent-operator-go && go test ./...` (includes golden + verdict + symmetry tests; envtest where configured)
- `npx moon run agent-operator-go:typecheck`
- `npx moon run agent-operator-go:lint`
- `npx moon run agent-operator-go:build`
- `npx moon run agent-operator-go:test`
- `npx moon run agent-cli-go:test` and `npx moon run shared-agent-go:test` (confirm shared `pkg/policy` consumed by both stays green)
- Manual: diff `config/rbac/role.yaml` against `grep -rn '+kubebuilder:rbac'` output — zero drift.

## Success Criteria

- [ ] `webhook.enforcePolicy` calls shared `policy.EnforcePolicy` via a computed `Capabilities`; the operator's duplicated admission algorithm is deleted.
- [ ] Job/pod-spec goldens unchanged across the relocation + `BuildJob`/`BuildWorkspaceJob` merge (no unintended output drift).
- [ ] Admission-verdict goldens unchanged across the engine swap (same accept/deny + reasons).
- [ ] `zz_generated.deepcopy.go` compiles and round-trips for the new typed enums.
- [ ] `config/rbac/role.yaml` exactly matches the relocated `+kubebuilder:rbac` markers.
- [ ] Cross-module symmetry test passes; axis dir-name sets intersect on the agreed axes.
- [ ] moon typecheck/lint/build/test green across all three modules; operator envtest green.

## Files Modified/Created

- `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook.go` (rewired)
- `packages/agent/agent-operator-go/internal/webhook/agentrun_webhook_test.go` (verdict goldens)
- `packages/agent/agent-operator-go/internal/builder/*_test.go` + `testdata/*.golden.yaml` (pod-spec goldens, in relocated dirs)
- `packages/agent/agent-operator-go/internal/arch/symmetry_test.go` (created)
- `packages/agent/agent-operator-go/api/v1alpha1/zz_generated.deepcopy.go` (regenerated/hand-updated)
- `packages/agent/agent-operator-go/config/rbac/role.yaml` (reconciled)

## Dependencies

- **06-operator-axis-relocation.md** — the per-axis dirs, merged Job builders, and typed CRD
  enums must exist first; this subplan locks their behavior with goldens and reconciles the
  codegen/RBAC the relocation invalidates. Transitively depends on
  `01-shared-per-axis-split` for `pkg/policy.EnforcePolicy` + `Capabilities`.

## Estimated Duration

Medium — 0.5–1 day. Mechanical: webhook rewire is small, the bulk is golden authoring,
hand-checking deepcopy/RBAC, and running the full three-module + envtest validation.
