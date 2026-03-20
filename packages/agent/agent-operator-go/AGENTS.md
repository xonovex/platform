# Agent Operator Go

- **AgentRun** — creates Jobs for agent execution (standalone or workspace-based)
- **AgentWorkspace** — shared RWX PVC for multi-agent coordination (git worktrees, shared config volumes)
- **AgentProvider** — reusable provider config with K8s secret management
- **AgentConfig** — namespace-level defaults

- Unit: `go test ./...` (builders, resolvers, webhooks)
- Integration: `go test -tags=integration ./test/integration/` (requires `KUBEBUILDER_ASSETS`)
- E2E: `go test -tags=e2e ./test/e2e/` (requires Docker, kind, kubectl)
- E2E gVisor: `go test -tags=e2e_gvisor ./test/e2e-gvisor/` (downloads runsc, creates kind cluster)
- E2E Kata: `go test -tags=e2e_kata ./test/e2e-kata/` (downloads Kata, requires `/dev/kvm`; VM isolation test skips in unprivileged kind — use a real cluster or `USE_EXISTING_CLUSTER=true`)

- **controller-gen broken with Go 1.25+**: generates `_.yaml` (empty group name) and omits sub-type DeepCopyInto methods; CRDs and `zz_generated.deepcopy.go` must be maintained manually until controller-tools supports Go 1.25+

- `docker build -f packages/agent/agent-operator-go/Dockerfile -t agent-operator:latest .`
- Moon: `npx moon run agent-operator-go:docker-build`
