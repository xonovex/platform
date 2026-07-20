# Agent Operator Go

- `AgentRun` is the only execution request API; external callers create runs directly through Kubernetes. Scheduling, event ingress, and trigger interpretation stay outside the operator.
- Keep harness, provider, workspace, and toolchain concerns selectable by reference or inline configuration; workspaces use RWX PVCs for coordination.
- A Nix toolchain is a pre-built, digest-pinned OCI image, never a per-pod install. `NixSpec` requires `nixpkgsRev`, `packages` XOR `flakeRef`/`shell`, and `image`; the webhook enforces `RequirePinnedProvision`. Build it with `npx moon run agent-operator-go:agent-image-build` (`nix build .#legacyPackages.<sys>.agentImage` plus skopeo push).
- Build operator images from the same `flake.lock` and `nix/agent-env.nix` closure as the CLI; verify store paths with `nix path-info -r`, not byte-identical layers. Do not add a `nixos/nix` init container or `nix-env` emptyDir.
- Harden untrusted pods fail-closed: require `RequireKernelIsolation` through a sandboxed `runtimeClassName` in `DefaultRuntimeClassName`/`AllowedRuntimeClassNames`, never default runc; use the controller-created zero-RBAC `agent-runner` ServiceAccount with `automountServiceAccountToken=false`, apply resource defaults plus `config/agent/` limits/quotas, and keep `readOnlyRootFilesystem=true` with a writable HOME `emptyDir` and `fsGroup=1000`.
- Create a default-deny egress `NetworkPolicy` per run: `none` is DNS-only, `proxy` allows public destinations except metadata/RFC1918/loopback plus DNS, and `host` allows all; use Cilium `toFQDNs`/Squid for FQDN-aware enforcement.
- Unit: `go test ./...` (builders, resolvers, webhooks)
- Integration: `go test -tags=integration ./test/integration/` (requires `KUBEBUILDER_ASSETS`)
- E2E: `go test -tags=e2e ./test/e2e/` (requires Docker, kind, kubectl)
- E2E gVisor: `go test -tags=e2e_gvisor ./test/e2e-gvisor/` (downloads runsc, creates kind cluster)
- E2E Kata: `go test -tags=e2e_kata ./test/e2e-kata/` (downloads Kata, requires `/dev/kvm`; VM isolation test skips in unprivileged kind — use a real cluster or `USE_EXISTING_CLUSTER=true`)
- E2E CoCo: `go test -tags=e2e_coco ./test/e2e-coco/` (creates kind cluster with simulated kata-cc/kata-tdx RuntimeClasses; validates runtimeClassName propagation, harness defaults, full-cycle pipeline, workspace jobs)
- `controller-gen` is broken with Go 1.25+: it generates `_.yaml` and omits subtype `DeepCopyInto` methods. Maintain CRDs and `zz_generated.deepcopy.go` manually until controller-tools supports Go 1.25+.
- `docker build -f packages/agent/agent-operator-go/Dockerfile -t ghcr.io/xonovex/agent-operator-go:latest .`
- Moon: `npx moon run agent-operator-go:docker-build`
- Multi-arch publish: `npx moon run agent-operator-go-docker:docker-publish`
