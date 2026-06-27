# Agent Operator Go

- **AgentRun** — creates Jobs for agent execution (standalone or workspace-based); 4 concerns via ref/inline: harness, provider, workspace, toolchain
- **AgentHarness** — agent type defaults (image, timeout, runtimeClassName, env)
- **AgentProvider** — reusable provider config with K8s secret management
- **AgentWorkspace** — shared RWX PVC for multi-agent coordination (git worktrees, shared config volumes)
- **AgentToolchain** — toolchain config (e.g. Nix packages)

- **Nix toolchain = nix-built OCI image** (not a per-pod install): `NixSpec` carries a pin (`nixpkgsRev`), a source (`packages` XOR `flakeRef`/`shell`), and a pre-built digest-pinned `image`. The pod runs that image (built from the same `flake.lock` + `nix/agent-env.nix` as the CLI — the SAME content-addressed store-path closure, verified via `nix path-info -r`, not byte-identical layers). Build/push the image with `npx moon run agent-operator-go:agent-image-build` (→ `nix build .#legacyPackages.<sys>.agentImage` + skopeo push). The webhook requires a pinned image (`RequirePinnedProvision`); there is no `nixos/nix` init container or `nix-env` emptyDir.
- **Untrusted-pod hardening** (fail closed): a sandboxed `runtimeClassName` via the existing `DefaultRuntimeClassName`/`AllowedRuntimeClassNames` machinery (set it on the harness — `RequireKernelIsolation`, never default runc); a dedicated zero-RBAC ServiceAccount (`agent-runner`, created by the controller) with `automountServiceAccountToken=false`; default resource requests/limits + the namespace `LimitRange`/`ResourceQuota` in `config/agent/`; a default-deny egress `NetworkPolicy` per `AgentRun` mapped from `Network` (`none`=DNS-only, `proxy`=public-except-metadata/RFC1918/loopback + DNS, `host`=allow-all), FQDN-aware via Cilium `toFQDNs`/Squid as the upgrade; `readOnlyRootFilesystem=true` reconciled with a writable HOME `emptyDir` + `fsGroup=1000`.

- Unit: `go test ./...` (builders, resolvers, webhooks)
- Integration: `go test -tags=integration ./test/integration/` (requires `KUBEBUILDER_ASSETS`)
- E2E: `go test -tags=e2e ./test/e2e/` (requires Docker, kind, kubectl)
- E2E gVisor: `go test -tags=e2e_gvisor ./test/e2e-gvisor/` (downloads runsc, creates kind cluster)
- E2E Kata: `go test -tags=e2e_kata ./test/e2e-kata/` (downloads Kata, requires `/dev/kvm`; VM isolation test skips in unprivileged kind — use a real cluster or `USE_EXISTING_CLUSTER=true`)
- E2E CoCo: `go test -tags=e2e_coco ./test/e2e-coco/` (creates kind cluster with simulated kata-cc/kata-tdx RuntimeClasses; validates runtimeClassName propagation, harness defaults, full-cycle pipeline, workspace jobs)

- **controller-gen broken with Go 1.25+**: generates `_.yaml` (empty group name) and omits sub-type DeepCopyInto methods; CRDs and `zz_generated.deepcopy.go` must be maintained manually until controller-tools supports Go 1.25+

- `docker build -f packages/agent/agent-operator-go/Dockerfile -t ghcr.io/xonovex/agent-operator-go:latest .`
- Moon: `npx moon run agent-operator-go:docker-build`
- Multi-arch publish: `npx moon run agent-operator-go-docker:docker-publish`
