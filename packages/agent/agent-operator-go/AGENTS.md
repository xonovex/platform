# Agent Operator Go

- **AgentRun** — creates Jobs for agent execution (standalone or workspace-based)
- **AgentWorkspace** — shared RWX PVC for multi-agent coordination (git worktrees, shared config volumes)
- **AgentProvider** — reusable provider config with K8s secret management
- **AgentConfig** — namespace-level defaults

- Unit: `go test ./...` (builders, resolvers, webhooks)
- Integration: `go test -tags=integration ./test/integration/` (requires `KUBEBUILDER_ASSETS`)
- E2E: `go test -tags=e2e ./test/e2e/` (requires Docker, kind, kubectl)

- `docker build -f packages/agent/agent-operator-go/Dockerfile -t agent-operator:latest .`
- Moon: `npx moon run agent-operator-go:docker-build`
