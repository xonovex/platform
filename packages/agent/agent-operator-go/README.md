# Agent Operator

Kubernetes operator for running AI coding agents (Claude, OpenCode) as Jobs with managed workspaces, provider secrets, and namespace-level defaults.

**API Group:** `agent.xonovex.com/v1alpha1`

## Custom Resources

### AgentRun

The primary workload resource. Each AgentRun creates a Job with an init container (git clone) and a main container (agent binary).

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
| `repository.url`                  | string   | Git repository URL (required)                                  |
| `repository.branch`               | string   | Branch to checkout                                             |
| `repository.commit`               | string   | Specific commit to checkout (overrides branch)                 |
| `repository.credentialsSecretRef` | object   | Secret reference for git credentials                           |
| `worktree.branch`                 | string   | Create a git worktree with this branch name                    |
| `worktree.sourceBranch`           | string   | Source branch to create the worktree from                      |
| `prompt`                          | string   | Task prompt for headless execution                             |
| `resources`                       | object   | K8s resource requirements for the agent container              |
| `timeout`                         | duration | Max run duration (default: `1h`)                               |
| `env`                             | list     | Additional environment variables                               |
| `image`                           | string   | Container image override                                       |
| `nodeSelector`                    | map      | Node selector for pod scheduling                               |
| `tolerations`                     | list     | Tolerations for pod scheduling                                 |

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

### 4. Run an agent

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

### 5. Monitor the run

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

### Using worktrees

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

- **Integration (15 tests):** Reconciler logic against a real API server — PVC/Job creation, phase transitions (Running, Succeeded, Failed, TimedOut), provider resolution, AgentConfig defaults, terminal phase skipping.
- **E2E (6 tests):** Full cluster behavior — Pod scheduling, PVC binding, init container failure propagation, owner reference garbage collection, and Docker image deployment with health probe validation.

## Architecture

Each AgentRun triggers the following:

1. **Workspace PVC** is created for persistent git storage
2. **Job** is created with:
   - **Init container**: clones the git repository (and optionally sets up a worktree)
   - **Main container**: runs the agent binary (claude/opencode) with resolved provider environment variables
3. The controller watches Job/Pod status and updates the AgentRun phase accordingly
4. On timeout, the Job is terminated and the AgentRun is marked `TimedOut`

```
AgentRun (created by user)
    |
    +---> AgentConfig (namespace defaults)
    +---> AgentProvider ---> Secret (auth token)
    |
    +---> PVC (workspace)
    +---> Job
            +---> Init Container (git clone)
            +---> Main Container (agent binary)
```

## Cleanup

```bash
# Delete a specific run (also cleans up its Job and PVC via owner references)
kubectl delete agentrun fix-auth-bug

# Uninstall the operator
kubectl delete -k config/default/

# Remove CRDs (deletes all AgentRun/AgentProvider/AgentConfig resources)
kubectl delete -k config/crd/
```
