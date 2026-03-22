# Agent Operator Go

- **AgentRun** — creates Jobs for agent execution (standalone or workspace-based); 4 concerns via ref/inline: harness, provider, workspace, toolchain
- **AgentHarness** — agent type defaults (image, timeout, runtimeClassName, env)
- **AgentProvider** — reusable provider config with K8s secret management
- **AgentWorkspace** — shared RWX PVC for multi-agent coordination (git worktrees, shared config volumes)
- **AgentToolchain** — toolchain config (e.g. Nix packages)

- Unit: `go test ./...` (builders, resolvers, webhooks)
- Integration: `go test -tags=integration ./test/integration/` (requires `KUBEBUILDER_ASSETS`)
- E2E: `go test -tags=e2e ./test/e2e/` (requires Docker, kind, kubectl)
- E2E gVisor: `go test -tags=e2e_gvisor ./test/e2e-gvisor/` (downloads runsc, creates kind cluster)
- E2E Kata: `go test -tags=e2e_kata ./test/e2e-kata/` (downloads Kata, requires `/dev/kvm`; VM isolation test skips in unprivileged kind — use a real cluster or `USE_EXISTING_CLUSTER=true`)
- E2E TEE: `go test -tags=e2e_tee ./test/e2e-tee/` (creates kind cluster with simulated kata-cc/kata-tdx RuntimeClasses and AKS CC node label; validates confidentialComputing spec, runtimeClassName selection, node affinity, harness defaults, workspace jobs)

- **controller-gen broken with Go 1.25+**: generates `_.yaml` (empty group name) and omits sub-type DeepCopyInto methods; CRDs and `zz_generated.deepcopy.go` must be maintained manually until controller-tools supports Go 1.25+

- `docker build -f packages/agent/agent-operator-go/Dockerfile -t ghcr.io/xonovex/agent-operator-go:latest .`
- Moon: `npx moon run agent-operator-go:docker-build`
- Multi-arch publish: `npx moon run agent-operator-go-docker:docker-publish`
