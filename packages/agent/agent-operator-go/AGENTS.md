# Agent Operator Go

Kubernetes operator for managing AI coding agent runs.

## CRDs

- **AgentRun**: Primary workload resource - creates Jobs for agent execution (standalone or workspace-based)
- **AgentWorkspace**: Shared workspace with RWX PVC for multi-agent coordination via git worktrees and shared config volumes
- **AgentProvider**: Reusable provider configuration with K8s secret management
- **AgentConfig**: Namespace-level defaults

## Testing

- `go test ./...` — unit tests (builders, resolvers, webhooks)
- `go test -tags=integration ./test/integration/` — envtest (requires `KUBEBUILDER_ASSETS`)
- `go test -tags=e2e ./test/e2e/` — Kind cluster (requires Docker, kind, kubectl)

## Docker

Build from repo root: `docker build -f packages/agent/agent-operator-go/Dockerfile -t agent-operator:latest .`

Moon task: `npx moon run agent-operator-go:docker-build`
