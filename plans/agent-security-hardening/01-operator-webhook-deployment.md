---
type: plan
has_subplans: false
parent_plan: plans/agent-security-hardening.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/agent/agent-operator-go/cmd/operator/main.go
    - packages/agent/agent-operator-go/config/**
    - packages/agent/agent-operator-go/test/e2e/suite_test.go
skills_to_consult:
  - kubernetes-guide
  - testing-guide
  - code-quality-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 01 — Operator: Make the Admission Layer Deployable

## Objective

Make every admission webhook in `agent-operator-go` actually run: registered in
the manager, reachable by the API server through `*WebhookConfiguration`
resources, served over TLS with cert-manager certificates, and exercised by the
e2e suite. Today none of this exists, so all "the webhook rejects X" behavior
is unit-test-only fiction.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-operator-go` is a Kubernetes operator built on
controller-runtime v0.24.1 (Go 1.26). It defines CRDs AgentRun, AgentWorkspace,
AgentHarness, AgentToolchain, AgentProvider, AgentPolicy under
`api/v1alpha1/`. Deployment manifests live under `config/` (kustomize).

Current defects, verified against source:

1. `cmd/operator/main.go:79` registers ONLY `AgentRunWebhook`. Four more
   webhook implementations exist in `internal/webhook/` —
   `agentworkspace_webhook.go`, `agentharness_webhook.go`,
   `agenttoolchain_webhook.go`, `agentprovider_webhook.go` — each with a
   `SetupWebhookWithManager` method that is never called. They are dead code in
   production.
2. There is NO `ValidatingWebhookConfiguration` or
   `MutatingWebhookConfiguration` manifest, no webhook `Service`, and no
   certificate setup anywhere under `config/`.
   `config/default/kustomization.yaml` includes only `crd`, `rbac`, `manager`.
   So even the one registered webhook is never invoked by the API server.
3. Because a webhook IS registered in main.go, controller-runtime starts its
   webhook server on `mgr.Start()`. That server needs TLS certs at
   `/tmp/k8s-webhook-server/serving-certs`, and `config/manager/manager.yaml`
   mounts none — the deployed operator most likely crash-loops on start.
4. `test/e2e/suite_test.go` runs controllers in-process via envtest and never
   registers webhooks, so nothing catches any of the above.

## Tasks

1. **Register all webhooks in the manager.**
   - File: `cmd/operator/main.go` (around line 79).
   - Read `internal/webhook/*.go` first to get the exact exported type names.
   - Register each of the five webhooks the same way `AgentRunWebhook` is
     registered today, with explicit error handling (log + exit) per webhook.
   - Gate webhook setup behind the existing pattern the project uses if one
     exists (check for an `ENABLE_WEBHOOKS`-style env guard used by kubebuilder
     scaffolds; if none exists, add one so envtest-based unit suites that do
     not install certs can still start the manager).

2. **Create `config/webhook/`.**
   - Check the webhook Go files for `//+kubebuilder:webhook:...` markers.
     - If markers exist: generate manifests with controller-gen (see how other
       generated manifests in `config/crd/` are produced; mirror that flow).
     - If markers are missing: add markers to each webhook type first
       (path, mutating/validating, failurePolicy, sideEffects=None,
       admissionReviewVersions=v1), then generate.
   - Add a `Service` (port 443 → targetPort 9443) selecting the manager pod.
   - `failurePolicy: Fail` on every webhook configuration entry. This is a
     deliberate decision from the parent plan: admission must fail closed.

3. **Certificates via cert-manager.**
   - Add `config/certmanager/` with a self-signed `Issuer` and a `Certificate`
     for the webhook Service DNS names (standard kubebuilder layout).
   - Annotate the webhook configurations for CA injection
     (`cert-manager.io/inject-ca-from`).
   - Mount the certificate secret in `config/manager/manager.yaml` at
     `/tmp/k8s-webhook-server/serving-certs` (read-only).

4. **Wire kustomize.**
   - `config/default/kustomization.yaml`: include `../webhook` and
     `../certmanager`, plus the standard patches (webhook manager patch, CA
     injection vars).
   - Verify rendering: `kubectl kustomize config/default` (or
     `kustomize build config/default`) must succeed and contain the Service,
     both WebhookConfigurations, Certificate, and the cert volume mount.

5. **Make e2e exercise admission.**
   - File: `test/e2e/suite_test.go`.
   - Use envtest's `WebhookInstallOptions` to install the webhook
     configurations into the test API server and start the manager's webhook
     server with the envtest-provisioned certs (controller-runtime supports
     this; read the existing suite setup and extend it, do not rewrite it).
   - Add one smoke assertion per webhook: an obviously-invalid object of each
     kind is REJECTED by the API server (not by calling the webhook function
     directly).

6. **Deployment resilience note.**
   - With `failurePolicy: Fail`, the operator becomes a hard dependency for
     AgentRun admission. Add a `PodDisruptionBudget` for the manager in
     `config/manager/` and ensure replicas ≥ 1 is explicit in the Deployment.

## Validation Steps

Prerequisites: the Go toolchain is nix-managed — if `go` is not on PATH, run
inside `nix develop` from the repo root. The integration/e2e tasks need
envtest binaries (controller-runtime `setup-envtest`; check how
`test/e2e/suite_test.go` locates them today) and `kubectl` for the kustomize
render check. cert-manager is NOT needed for envtest-based e2e — only for
real-cluster deploys.

```bash
npx moon run agent-operator-go:go-build
npx moon run agent-operator-go:go-lint
npx moon run agent-operator-go:go-test
npx moon run agent-operator-go:go-test-integration
npx moon run agent-operator-go:go-test-e2e
kubectl kustomize packages/agent/agent-operator-go/config/default > /dev/null
```

## Success Criteria

- [ ] All five webhooks registered in `main.go`; no `SetupWebhookWithManager`
      implementation is uncalled.
- [ ] `config/webhook/` + `config/certmanager/` exist; `config/default` renders
      with Service, WebhookConfigurations (`failurePolicy: Fail`), Certificate,
      and the manager cert mount.
- [ ] e2e suite installs webhook configurations and proves at least one
      API-server-level rejection per webhook.
- [ ] Manager starts cleanly in e2e with the webhook server serving.
- [ ] All validation commands above pass.

## Files Modified/Created

- Modified: `cmd/operator/main.go`, `config/default/kustomization.yaml`,
  `config/manager/manager.yaml`, `test/e2e/suite_test.go`,
  `internal/webhook/*.go` (markers only, if missing)
- Created: `config/webhook/*`, `config/certmanager/*`

## Dependencies

None — this is the gate for subplans 03/04/05.

## Estimated Duration

1–1.5 days.
