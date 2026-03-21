# Agent Operator

Kubernetes operator for running AI coding agents (Claude, OpenCode) as Jobs with managed workspaces, provider secrets, and namespace-level defaults. Supports shared multi-agent workspaces where multiple agents coordinate via a common git checkout and shared config/state directories. Supports sandboxed execution via gVisor or Kata Containers runtime classes. Supports [Jujutsu (jj)](https://github.com/jj-vcs/jj) as an alternative VCS for automatic snapshotting and operation-log based undo.

**API Group:** `agent.xonovex.com/v1alpha1`

## Custom Resources

### AgentRun

The primary workload resource. Each AgentRun creates a Job with an init container (git clone) and a main container (agent binary). Runs can be standalone (own PVC) or reference a shared AgentWorkspace. Supports sandboxed runtimes via `runtimeClassName`.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: review-codebase
spec:
  agent: claude # e.g. "claude", "opencode"
  configRef: default # references an AgentConfig for defaults
  providerRef: gemini-provider # references an AgentProvider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Review the codebase and suggest improvements"
  timeout: 30m
  runtimeClassName: gvisor # optional — run in a gVisor sandbox
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
| `agent`                           | string   | Agent type (e.g. `claude`, `opencode`)                         |
| `configRef`                       | string   | Name of an AgentConfig in the same namespace for defaults      |
| `config`                          | object   | Inline config (mutually exclusive with `configRef`)            |
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
| `runtimeClassName`                | string   | Pod runtime class for sandboxed execution (e.g. `gvisor`, `kata`) |
| `vcs`                             | string   | Version control system (e.g. `git`, `jj`; default: `git`)     |
| `nix.packages`                    | list     | Nixpkgs attribute names to install (e.g. `nodejs_22`, `python3`) |
| `nix.image`                       | string   | Nix container image for init container (default: `nixos/nix:latest`) |
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

Reusable configuration with namespace-level defaults. AgentRuns reference an AgentConfig via `configRef` to inherit defaults. Multiple configs can coexist in the same namespace for different workloads.

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
  defaultRuntimeClassName: gvisor # all runs in this namespace use gVisor by default
  defaultTimeout: 1h
  storageClass: standard
  storageSize: 10Gi
  env:
    - name: LANG
      value: "en_US.UTF-8"
```

#### Full spec reference

| Field                      | Type     | Description                                                  |
| -------------------------- | -------- | ------------------------------------------------------------ |
| `defaultAgent`             | string   | Default agent type for new runs                              |
| `defaultProviders`         | map      | Map of agent type to default provider name                   |
| `defaultImage`             | string   | Default container image                                      |
| `defaultRuntimeClassName`  | string   | Default pod runtime class (e.g. `gvisor`, `kata`)            |
| `defaultVCS`               | string   | Default version control system (e.g. `git`, `jj`)             |
| `defaultNix.packages`      | list     | Default Nix packages for all runs referencing this config     |
| `defaultNix.image`         | string   | Default Nix container image                                   |
| `defaultTimeout`           | duration | Default timeout for agent runs                               |
| `storageClass`             | string   | Default storage class for workspace PVCs                     |
| `storageSize`              | string   | Default storage size for workspace PVCs                      |
| `env`                      | list     | Default environment variables for all runs                   |

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

### Standalone agent run (full workflow)

The typical workflow: create a Secret, an AgentProvider, optionally an AgentConfig for namespace defaults, then run an agent.

```bash
# 1. Create a Secret for your provider credentials
kubectl create secret generic gemini-credentials \
  --from-literal=api-key='your-api-key-here'
```

```yaml
# 2. Create an AgentProvider
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

```yaml
# 3. (Optional) Create a reusable config with defaults
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

```yaml
# 4. Run the agent (referencing config and provider)
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: fix-auth-bug
spec:
  agent: claude
  configRef: default          # inherit defaults from AgentConfig
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: feature/auth
  prompt: "Fix the authentication bug in the login handler"
  timeout: 30m
```

```bash
kubectl apply -f provider.yaml -f config.yaml -f run.yaml
kubectl get agentproviders
# NAME              DISPLAY NAME     READY   AGE
# gemini-provider   Google Gemini    true    5s

kubectl get agentruns -w
# NAME           AGENT    PHASE         AGE
# fix-auth-bug   claude   Pending       0s
# fix-auth-bug   claude   Initializing  1s
# fix-auth-bug   claude   Running       5s
# fix-auth-bug   claude   Succeeded     45s
```

### Sandboxed agent run (gVisor)

Run agents inside a gVisor sandbox for syscall-level isolation. Requires the `gvisor` RuntimeClass to be configured on your cluster.

```yaml
# RuntimeClass (cluster setup — once per cluster)
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: gvisor
handler: runsc
---
# AgentRun with gVisor sandbox
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: sandboxed-review
spec:
  agent: claude
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Review the codebase for security issues"
  runtimeClassName: gvisor
```

The `runtimeClassName` is applied to the Job's PodSpec. Both the init container (git clone) and the main container (agent binary) run inside the gVisor sandbox.

### Sandboxed agent run (Kata Containers)

Run agents inside a Kata Containers VM for hardware-level isolation. Requires the `kata` RuntimeClass and hardware virtualization support (`/dev/kvm`).

```yaml
# RuntimeClass (cluster setup — once per cluster)
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: kata
handler: kata
---
# AgentRun with Kata VM isolation
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: isolated-agent
spec:
  agent: claude
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Implement the payment processing module"
  runtimeClassName: kata
  timeout: 1h
```

### Sandbox default via config

Use AgentConfig to provide sandboxed runtime defaults that runs inherit via `configRef`:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentConfig
metadata:
  name: sandboxed
spec:
  defaultAgent: claude
  defaultProviders:
    claude: gemini-provider
  defaultRuntimeClassName: gvisor  # runs referencing this config use gVisor
  storageSize: 10Gi
---
# This run inherits runtimeClassName=gvisor from the config
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: auto-sandboxed
spec:
  agent: claude
  configRef: sandboxed
  repository:
    url: https://github.com/org/repo.git
  prompt: "Add input validation to all API endpoints"
```

### Agent run with Jujutsu (jj)

Run agents with [Jujutsu](https://github.com/jj-vcs/jj) for automatic working copy snapshotting, operation-log based undo, and first-class conflict handling. The operator clones with git, then initializes jj in colocated mode (`jj git init --colocate`) so both tools work side-by-side.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: jj-agent
spec:
  agent: claude
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Refactor the error handling"
  vcs: jj  # clone with git, then init jj colocated
```

The agent container image must include the `jj` binary. When `vcs: jj` is set:
- **Standalone clone**: `git clone ... && jj git init --colocate`
- **Standalone worktree**: `jj workspace add` instead of `git worktree add`
- **Workspace init**: clone + `jj git init --colocate`
- **Workspace worktree**: `jj workspace add` instead of `git worktree add`

### Jujutsu default via config

Use AgentConfig to provide jj defaults that runs inherit via `configRef`:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentConfig
metadata:
  name: jj-config
spec:
  defaultAgent: claude
  defaultProviders:
    claude: gemini-provider
  defaultVCS: jj  # runs referencing this config use jj
  storageSize: 10Gi
---
# This run inherits vcs=jj from the config
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: auto-jj
spec:
  agent: claude
  configRef: jj-config
  repository:
    url: https://github.com/org/repo.git
  prompt: "Add input validation to all API endpoints"
```

### Agent run with Nix packages

Provision reproducible tool environments using [Nix](https://nixos.org/). The operator adds a `nix-env` init container that installs packages from nixpkgs into a shared volume, making them available in the agent's PATH.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: nix-agent
spec:
  agent: claude
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
    branch: main
  prompt: "Set up the CI pipeline"
  nix:
    packages:
      - nodejs_22
      - python3
      - ripgrep
      - jujutsu
```

When `nix.packages` is set, the operator:
1. Adds an emptyDir volume (`nix-env`)
2. Adds a `nix-env` init container (using `nixos/nix:latest`) that bootstraps the Nix store to the volume and installs the packages via `nix profile install`
3. Mounts the volume at `/nix` in the main container
4. Prepends `/nix/var/nix/profiles/agent/bin` to `PATH`

Package names are [nixpkgs](https://search.nixos.org/packages) attributes — the same names you'd use with `nix profile install nixpkgs#<name>`.

### Nix defaults via config

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentConfig
metadata:
  name: nix-config
spec:
  defaultAgent: claude
  defaultProviders:
    claude: gemini-provider
  defaultNix:
    packages:
      - nodejs_22
      - python3
      - git
  storageSize: 10Gi
---
# Inherits nix packages from config
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: auto-nix
spec:
  agent: claude
  configRef: nix-config
  repository:
    url: https://github.com/org/repo.git
  prompt: "Fix the build"
```

### Nix + jj + sandbox (combined)

All features compose. An agent can use Nix packages, Jujutsu VCS, and gVisor sandboxing simultaneously:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: full-stack-agent
spec:
  agent: claude
  configRef: default
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
  prompt: "Implement the payment processing module"
  vcs: jj
  runtimeClassName: gvisor
  nix:
    packages:
      - nodejs_22
      - python3
      - postgresql
```

### Multi-agent shared workspace

Create a workspace with shared volumes, then launch concurrent agents. Each agent gets an isolated git worktree but shares the same checkout and config directories.

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

### Multi-agent shared workspace with jj

Combine workspace-based runs with Jujutsu. The workspace init job clones with git and initializes jj colocated. Each agent run creates a jj workspace instead of a git worktree.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: jj-workspace
spec:
  repository:
    url: https://github.com/org/repo.git
    branch: main
  storageClass: nfs-csi
  storageSize: 10Gi
  vcs: jj  # init job runs: git clone + jj git init --colocate
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: jj-agent-1
spec:
  agent: claude
  workspaceRef: jj-workspace
  worktree:
    branch: agent-1-work
  providerRef: gemini-provider
  vcs: jj  # uses: jj workspace add (instead of git worktree add)
  prompt: "Implement the search feature"
```

### Multi-agent shared workspace with sandbox

Combine workspace-based runs with runtime sandboxing. The workspace init Job runs with the default runtime, while each agent Job runs in the sandbox.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: sandboxed-ws
spec:
  repository:
    url: https://github.com/org/repo.git
    branch: main
  storageClass: nfs-csi
  storageSize: 10Gi
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: sandboxed-agent-1
spec:
  agent: claude
  workspaceRef: sandboxed-ws
  worktree:
    branch: agent-1-work
  providerRef: gemini-provider
  runtimeClassName: gvisor  # agent Job runs in gVisor; workspace init Job does not
  prompt: "Refactor the database layer"
```

### Inline config (no AgentConfig resource needed)

For one-off runs, you can specify the config inline instead of creating a separate AgentConfig resource:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: quick-run
spec:
  agent: claude
  config:
    defaultImage: "node:22-slim"
    defaultTimeout: 30m
    defaultRuntimeClassName: gvisor
    defaultNix:
      packages:
        - nodejs_22
        - ripgrep
  providerRef: gemini-provider
  repository:
    url: https://github.com/org/repo.git
  prompt: "Fix the tests"
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

### Monitoring runs

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

# E2E gVisor tests (creates Kind cluster, installs runsc, runs full workflow in gVisor sandbox)
# Requires: kind, kubectl, Docker, internet access (downloads runsc binary)
go test -tags=e2e_gvisor -v -timeout=600s ./test/e2e-gvisor/

# E2E Kata tests (creates Kind cluster, installs Kata, runs full workflow in Kata VM)
# Requires: kind, kubectl, Docker, /dev/kvm, internet access (downloads Kata release)
# Note: VM isolation tests skip gracefully in unprivileged kind (vsock/QEMU errors)
go test -tags=e2e_kata -v -timeout=600s ./test/e2e-kata/
```

### What the tests cover

- **Unit (68 tests):** Builders (PVC, Job, containers, env vars, workspace PVC/Job/worktree), webhooks (defaulting and validation for all 4 CRDs including workspaceRef rules), resolvers (config, provider, workspace).
- **Integration (20 tests):** Reconciler logic against a real API server — PVC/Job creation, phase transitions (Running, Succeeded, Failed, TimedOut), provider resolution, AgentConfig defaults, terminal phase skipping, AgentWorkspace PVC creation, workspace Ready/Failed transitions, AgentRun with workspaceRef waiting for workspace Ready, backward compatibility.
- **E2E (7 tests):** Full cluster behavior — Pod scheduling, PVC binding, init container failure propagation, owner reference garbage collection, Docker image deployment with health probe validation, multi-agent workspace with concurrent runs, full-cycle pipeline (git clone + fake agent binary → Succeeded).
- **E2E gVisor (5 tests):** Sandbox isolation verification (dmesg gVisor banner), runtimeClassName propagation to Job/Pod, AgentConfig default inheritance, full workflow (Secret + Provider + Config + git clone + agent → Succeeded inside gVisor), workspace-based run (init Job has no runtimeClassName, agent Job does).
- **E2E Kata (4 tests):** VM isolation verification (guest kernel differs from host, /dev/pmem0), runtimeClassName propagation to Job/Pod, AgentConfig default inheritance, full workflow (Secret + Provider + Config + git clone + agent → Succeeded inside Kata VM, skips in unprivileged kind).

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

**RuntimeClassName** is applied to the Job's PodSpec when set on the AgentRun or inherited from the referenced AgentConfig. Both init and main containers run in the sandboxed runtime. Workspace init Jobs do *not* inherit runtimeClassName — only agent Jobs do.

```
Standalone:                         Workspace:

AgentRun                            AgentWorkspace
    |                                   |
    +-> AgentConfig (via configRef)     +-> PVCs (RWX): workspace + shared volumes
    +-> AgentProvider -> Secret         +-> Init Job (git clone) -> Ready
    |                                   |
    +-> PVC (RWO)                   AgentRun (workspaceRef)
    +-> Job                             |
          +-> Init: git clone           +-> AgentConfig (via configRef)
          +-> Main: agent binary        +-> AgentProvider -> Secret
          +-> runtimeClassName?         |
                                        +-> Job (uses workspace PVC)
                                              +-> Init: git worktree add
                                              +-> Main: agent binary
                                                    workingDir: /workspace-wt/{run}
                                                    mounts: shared volumes
                                              +-> runtimeClassName?
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
