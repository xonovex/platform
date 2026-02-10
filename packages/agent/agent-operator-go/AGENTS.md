# Agent Operator Go

Kubernetes operator for managing AI coding agent runs.

## CRDs

- **AgentRun** — creates Jobs for agent execution (standalone or workspace-based)
- **AgentWorkspace** — shared RWX PVC for multi-agent coordination (git worktrees, shared config volumes)
- **AgentProvider** — reusable provider config with K8s secret management
- **AgentConfig** — namespace-level defaults

## Testing

- `go test ./...` — unit tests (builders, resolvers, webhooks)
- `go test -tags=integration ./test/integration/` — envtest (requires `KUBEBUILDER_ASSETS`)
- `go test -tags=e2e ./test/e2e/` — Kind cluster (requires Docker, kind, kubectl)

## Docker

- Build → `docker build -f packages/agent/agent-operator-go/Dockerfile -t agent-operator:latest .`
- Moon → `npx moon run agent-operator-go:docker-build`
