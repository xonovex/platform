# Agent Operator

Kubernetes operator for running AI coding agents (Claude, OpenCode) as Jobs with managed workspaces, provider secrets, and namespace-level defaults. Supports shared multi-agent workspaces where multiple agents coordinate via a common git checkout and shared config/state directories.

**API Group:** `agent.xonovex.com/v1alpha1`

## Custom Resources

### AgentRun

The primary workload resource. Each AgentRun creates a Job with an init container (git clone) and a main container (agent binary). Runs can be standalone (own PVC) or reference a shared AgentWorkspace.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: review-codebase
spec:
  agent: claude # "claude" or "opencode"
  providerRef: gemini-provider # references an AgentProvider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Review the codebase and suggest improvements"
  timeout: 30m
  resources:
    requests:
      cpu: "500m"
      memory: "512Mi"
    limits:
      cpu: "2"
      memory: "2Gi"
```

**Lifecycle phases:** `Pending` -> `Initializing` -> `Running` -> `Succeeded` | `Failed` | `TimedOut`

#### Full spec reference

| Field                             | Type     | Description                                                    |
| --------------------------------- | -------- | -------------------------------------------------------------- |
| `agent`                           | string   | Agent type: `claude` or `opencode`                             |
| `providerRef`                     | string   | Name of an AgentProvider in the same namespace                 |
| `provider`                        | object   | Inline provider config (mutually exclusive with `providerRef`) |
| `workspaceRef`                    | string   | Name of an AgentWorkspace (mutually exclusive with `repository`) |
| `repository.url`                  | string   | Git repository URL (required when no `workspaceRef`)           |
| `repository.branch`               | string   | Branch to checkout                                             |
| `repository.commit`               | string   | Specific commit to checkout (overrides branch)                 |
| `repository.credentialsSecretRef` | object   | Secret reference for git credentials                           |
| `worktree.branch`                 | string   | Create a git worktree with this branch name (required with `workspaceRef`) |
| `worktree.sourceBranch`           | string   | Source branch to create the worktree from                      |
| `prompt`                          | string   | Task prompt for headless execution                             |
| `resources`                       | object   | K8s resource requirements for the agent container              |
| `timeout`                         | duration | Max run duration (default: `1h`)                               |
| `env`                             | list     | Additional environment variables                               |
| `image`                           | string   | Container image override                                       |
| `nodeSelector`                    | map      | Node selector for pod scheduling                               |
| `tolerations`                     | list     | Tolerations for pod scheduling                                 |

### AgentWorkspace

Owns a shared git checkout (ReadWriteMany PVC) and optional shared volumes for agent config/state directories. Multiple AgentRuns reference the workspace via `workspaceRef`, each creating its own git worktree for isolation.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: my-workspace
spec:
  repository:
    url: https://github.com/org/repo.git
    branch: main
  storageClass: nfs-csi          # must support ReadWriteMany
  storageSize: 10Gi
  sharedVolumes:                 # optional — shared config/state dirs for agents
    - name: claude-config
      mountPath: /root/.claude
      storageSize: 1Gi
    - name: opencode-config
      mountPath: /root/.opencode
      storageSize: 512Mi
```

**Lifecycle phases:** `Pending` -> `Initializing` -> `Ready` | `Failed`

#### Full spec reference

| Field                    | Type   | Description                                          |
| ------------------------ | ------ | ---------------------------------------------------- |
| `repository.url`         | string | Git repository URL (required)                        |
| `repository.branch`      | string | Branch to checkout                                   |
| `storageClass`           | string | Storage class for workspace PVC (must support RWX)   |
| `storageSize`            | string | Storage size for workspace PVC (default: `10Gi`)     |
| `sharedVolumes[].name`   | string | Volume name (used as PVC suffix)                     |
| `sharedVolumes[].mountPath` | string | Mount path in agent containers                    |
| `sharedVolumes[].storageSize` | string | PVC size for this volume (default: `1Gi`)       |

#### Volume layout

```
workspace PVC (RWX):
  /workspace/              <- main git checkout (from init Job)
  /workspace/.git/         <- shared .git dir
  /workspace-wt/agent-1/   <- worktree for agent-1
  /workspace-wt/agent-2/   <- worktree for agent-2

shared volume PVCs (RWX, one per sharedVolumes entry):
  /root/.claude/           <- claude-config PVC
  /root/.opencode/         <- opencode-config PVC
```

### AgentProvider

Reusable provider configuration with Kubernetes-native secret management. Auth tokens are read from Secrets instead of environment variables.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentProvider
metadata:
  name: gemini-provider
spec:
  displayName: Google Gemini
  agentTypes:
    - claude
  authTokenSecretRef:
    name: gemini-credentials
    key: api-key
  environment:
    ANTHROPIC_BASE_URL: "http://litellm-proxy:8317"
    API_TIMEOUT_MS: "3000000"
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1"
    ANTHROPIC_DEFAULT_SONNET_MODEL: "gemini-3-flash-preview"
```

The controller validates that the referenced Secret exists and contains the specified key, reporting readiness via `.status.ready`.

### AgentConfig

Namespace-level defaults applied to all AgentRuns in the namespace. Create one per namespace (singleton by convention).

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentConfig
metadata:
  name: default
spec:
  defaultAgent: claude
  defaultProviders:
    claude: gemini-provider
    opencode: gemini-opencode-provider
  defaultImage: "node:trixie-slim"
  defaultTimeout: 1h
  storageClass: standard
  storageSize: 10Gi
  env:
    - name: LANG
      value: "en_US.UTF-8"
```

## Installation

### Prerequisites

- Kubernetes cluster (v1.28+)
- `kubectl` configured to access the cluster
- `kustomize` (or `kubectl` with built-in kustomize)

### Install CRDs

```bash
kubectl apply -k config/crd/
```

### Deploy the operator

```bash
# Build the Docker image (run from repo root)
docker build -f packages/agent/agent-operator-go/Dockerfile -t agent-operator:latest .

# Deploy with kustomize (uses the default namespace/RBAC configuration)
kubectl apply -k config/default/
```

### Run locally (for development)

```bash
# Install CRDs first
kubectl apply -k config/crd/

# Run the operator against your current kubeconfig
go run ./cmd/operator/ \
  --health-probe-bind-address=:8081 \
  --metrics-bind-address=:8080
```

## Usage

### 1. Create a Secret for your provider credentials

```bash
kubectl create secret generic gemini-credentials \
  --from-literal=api-key='your-api-key-here'
```

### 2. Create an AgentProvider

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentProvider
metadata:
  name: gemini-provider
spec:
  displayName: Google Gemini
  agentTypes: [claude]
  authTokenSecretRef:
    name: gemini-credentials
    key: api-key
  environment:
    ANTHROPIC_BASE_URL: "http://litellm-proxy:8317"
```

```bash
kubectl apply -f provider.yaml
kubectl get agentproviders
# NAME              DISPLAY NAME     READY   AGE
# gemini-provider   Google Gemini    true    5s
```

### 3. (Optional) Create namespace defaults

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentConfig
metadata:
  name: default
spec:
  defaultAgent: claude
  defaultProviders:
    claude: gemini-provider
  defaultTimeout: 1h
  storageSize: 10Gi
```

### 4. Run an agent (standalone)

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: fix-auth-bug
spec:
  agent: claude
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: feature/auth
  prompt: "Fix the authentication bug in the login handler"
  timeout: 30m
```

```bash
kubectl apply -f run.yaml
```

### 5. Run multiple agents on a shared workspace

Create a workspace with shared volumes, then launch concurrent agents:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: my-workspace
spec:
  repository:
    url: https://github.com/org/repo.git
    branch: main
  storageClass: nfs-csi
  storageSize: 10Gi
  sharedVolumes:
    - name: claude-config
      mountPath: /root/.claude
      storageSize: 1Gi
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: agent-1
spec:
  agent: claude
  workspaceRef: my-workspace
  worktree:
    branch: agent-1-work
  providerRef: gemini-provider
  prompt: "Fix the login bug"
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: agent-2
spec:
  agent: claude
  workspaceRef: my-workspace
  worktree:
    branch: agent-2-work
  providerRef: gemini-provider
  prompt: "Add unit tests for the auth module"
```

```bash
kubectl apply -f workspace.yaml
kubectl get agentworkspaces
# NAME           PHASE   AGE
# my-workspace   Ready   30s

kubectl get agentruns
# NAME      AGENT    PHASE     AGE
# agent-1   claude   Running   15s
# agent-2   claude   Running   15s
```

### 6. Monitor runs

```bash
# Watch AgentRun status
kubectl get agentruns -w
# NAME           AGENT    PHASE      AGE
# fix-auth-bug   claude   Running    30s

# Check the underlying Job and Pod
kubectl get jobs,pods -l agent.xonovex.com/agent-type=claude

# View agent logs
kubectl logs job/fix-auth-bug -c agent -f

# View init container logs (git clone)
kubectl logs job/fix-auth-bug -c git-clone
```

### Inline provider (no AgentProvider resource needed)

For one-off runs, you can specify the provider inline:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: quick-review
spec:
  agent: claude
  provider:
    name: anthropic
    authSecretRef:
      name: anthropic-credentials
      key: api-key
    environment:
      ANTHROPIC_BASE_URL: "https://api.anthropic.com"
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Review PR changes"
```

### Using worktrees (standalone)

Create agent runs that work in isolated git worktrees:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: feature-work
spec:
  agent: claude
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  worktree:
    branch: agent/feature-work
    sourceBranch: main
  prompt: "Implement the user settings page"
```

## Testing

```bash
# Unit tests
go test ./...

# Integration tests (envtest — real API server, no kubelet)
# Requires: setup-envtest (go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)
KUBEBUILDER_ASSETS=$(setup-envtest use -p path) go test -tags=integration -v -timeout=300s ./test/integration/

# E2E tests (Kind — full cluster with scheduling and garbage collection)
# Requires: kind, kubectl, Docker
go test -tags=e2e -v -timeout=600s ./test/e2e/

# E2E against an existing cluster (skips Kind creation, skips image deployment test)
USE_EXISTING_CLUSTER=true go test -tags=e2e -v -timeout=600s ./test/e2e/
```

### What the tests cover

- **Unit (68 tests):** Builders (PVC, Job, containers, env vars, workspace PVC/Job/worktree), webhooks (defaulting and validation for all 4 CRDs including workspaceRef rules), resolvers (config, provider, workspace).
- **Integration (20 tests):** Reconciler logic against a real API server — PVC/Job creation, phase transitions (Running, Succeeded, Failed, TimedOut), provider resolution, AgentConfig defaults, terminal phase skipping, AgentWorkspace PVC creation, workspace Ready/Failed transitions, AgentRun with workspaceRef waiting for workspace Ready, backward compatibility.
- **E2E (7 tests):** Full cluster behavior — Pod scheduling, PVC binding, init container failure propagation, owner reference garbage collection, Docker image deployment with health probe validation, multi-agent workspace with concurrent runs.

## Architecture

Each AgentRun triggers one of two paths:

**Standalone path** (no `workspaceRef`):
1. **Workspace PVC** (RWO) is created for persistent git storage
2. **Job** is created with init container (git clone) and main container (agent binary)
3. Controller watches Job status and updates AgentRun phase

**Workspace path** (with `workspaceRef`):
1. **AgentWorkspace** must be in `Ready` phase (requeue if not)
2. **Job** is created using the workspace's shared PVC (RWX) with init container (git worktree add) and main container (agent binary working in the worktree)
3. Shared volume PVCs are mounted at configured paths (e.g. `~/.claude/`)
4. Controller watches Job status and updates AgentRun phase

```
Standalone:                         Workspace:

AgentRun                            AgentWorkspace
    |                                   |
    +-> AgentConfig (defaults)          +-> PVCs (RWX): workspace + shared volumes
    +-> AgentProvider -> Secret         +-> Init Job (git clone) -> Ready
    |                                   |
    +-> PVC (RWO)                   AgentRun (workspaceRef)
    +-> Job                             |
          +-> Init: git clone           +-> AgentConfig (defaults)
          +-> Main: agent binary        +-> AgentProvider -> Secret
                                        |
                                        +-> Job (uses workspace PVC)
                                              +-> Init: git worktree add
                                              +-> Main: agent binary
                                                    workingDir: /workspace-wt/{run}
                                                    mounts: shared volumes
```

## Cleanup

```bash
# Delete a specific run (also cleans up its Job via owner references)
kubectl delete agentrun fix-auth-bug

# Delete a workspace (also cleans up its PVCs and init Job via owner references)
# Note: worktrees created by AgentRuns remain on disk in the PVC
kubectl delete agentworkspace my-workspace

# Uninstall the operator
kubectl delete -k config/default/

# Remove CRDs (deletes all AgentRun/AgentProvider/AgentConfig/AgentWorkspace resources)
kubectl delete -k config/crd/
```
