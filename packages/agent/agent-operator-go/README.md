# Agent Operator

Kubernetes operator for running AI coding agents (Claude, OpenCode) as Jobs with managed workspaces, provider secrets, and namespace-level defaults. Supports shared multi-agent workspaces where multiple agents coordinate via a common git checkout and shared config/state directories. Supports sandboxed execution via gVisor, Kata Containers, or Confidential Containers (CoCo) with AMD SEV-SNP / Intel TDX runtime classes. Supports [Jujutsu (jj)](https://github.com/jj-vcs/jj) as an alternative VCS for automatic snapshotting and operation-log based undo.

**API Group:** `agent.xonovex.com/v1alpha1`

## Custom Resources

AgentRun references four concerns via ref or inline: **harness**, **provider**, **workspace**, and **toolchain**.

### AgentRun

The primary workload resource. Each AgentRun creates a Job with an init container (git clone) and a main container (agent binary). Runs can be standalone (own PVC) or reference a shared AgentWorkspace.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: review-codebase
spec:
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: main
  prompt: "Review the codebase and suggest improvements"
  timeout: 30m
  runtimeClassName: gvisor
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

| Field | Type | Description |
| --- | --- | --- |
| `harnessRef` | string | Name of an AgentHarness in the same namespace |
| `harness` | object | Inline harness config (mutually exclusive with `harnessRef`) |
| `providerRef` | string | Name of an AgentProvider in the same namespace |
| `provider` | object | Inline provider config (mutually exclusive with `providerRef`) |
| `workspaceRef` | string | Name of an AgentWorkspace for shared workspace support |
| `workspace` | object | Inline workspace config (mutually exclusive with `workspaceRef`) |
| `toolchainRef` | string | Name of an AgentToolchain in the same namespace |
| `toolchain` | object | Inline toolchain config (mutually exclusive with `toolchainRef`) |
| `prompt` | string | Task prompt for headless execution |
| `resources` | object | K8s resource requirements for the agent container |
| `timeout` | duration | Max run duration (default: `1h`) |
| `env` | list | Additional environment variables |
| `image` | string | Container image override |
| `runtimeClassName` | string | Pod runtime class for sandboxed execution (e.g. `gvisor`, `kata`) |
| `nodeSelector` | map | Node selector for pod scheduling |
| `tolerations` | list | Tolerations for pod scheduling |

### AgentHarness

Agent type defaults (image, timeout, runtimeClassName, env). Multiple harnesses can coexist in a namespace for different agent types.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentHarness
metadata:
  name: claude-harness
spec:
  type: claude
  defaultProvider: gemini-provider
  defaultImage: "node:trixie-slim"
  defaultRuntimeClassName: gvisor
  defaultTimeout: 1h
  env:
    - name: LANG
      value: "en_US.UTF-8"
```

#### Full spec reference

| Field | Type | Description |
| --- | --- | --- |
| `type` | string | Agent type (`claude`, `opencode`) |
| `defaultProvider` | string | Default provider name |
| `defaultImage` | string | Default container image |
| `defaultResources` | object | Default resource requirements |
| `defaultTimeout` | duration | Default timeout for agent runs |
| `defaultRuntimeClassName` | string | Default pod runtime class (e.g. `gvisor`, `kata`) |
| `env` | list | Default environment variables |

### AgentProvider

Reusable provider configuration with Kubernetes-native secret management. Auth tokens are read from Secrets instead of environment variables.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentProvider
metadata:
  name: gemini-provider
spec:
  displayName: Google Gemini
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

#### Full spec reference

| Field | Type | Description |
| --- | --- | --- |
| `type` | string | Provider type (e.g. `anthropic`, `openai`) |
| `displayName` | string | Human-readable name |
| `authTokenSecretRef` | object | Secret reference for auth token |
| `environment` | map | Environment variables to set |
| `cliArgs` | list | Additional CLI arguments |

### AgentWorkspace

Owns a shared git checkout (ReadWriteMany PVC) and optional shared volumes for agent config/state directories. Multiple AgentRuns reference the workspace via `workspaceRef`, each creating its own git worktree for isolation.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: my-workspace
spec:
  type: git
  repository:
    url: https://github.com/org/repo.git
    branch: main
  storageClass: nfs-csi
  storageSize: 10Gi
  sharedVolumes:
    - name: claude-config
      mountPath: /root/.claude
      storageSize: 1Gi
    - name: opencode-config
      mountPath: /root/.opencode
      storageSize: 512Mi
```

**Lifecycle phases:** `Pending` -> `Initializing` -> `Ready` | `Failed`

#### Full spec reference

| Field | Type | Description |
| --- | --- | --- |
| `type` | string | Workspace type (`git` or `jj`) |
| `repository.url` | string | Git repository URL (required) |
| `repository.branch` | string | Branch to checkout |
| `storageClass` | string | Storage class for workspace PVC (must support RWX) |
| `storageSize` | string | Storage size for workspace PVC (default: `10Gi`) |
| `sharedVolumes[].name` | string | Volume name (used as PVC suffix) |
| `sharedVolumes[].mountPath` | string | Mount path in agent containers |
| `sharedVolumes[].storageSize` | string | PVC size for this volume (default: `1Gi`) |
| `git.worktree` | object | Git worktree configuration |
| `jj.revision` | string | Jujutsu revision |

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

### AgentToolchain

Reusable toolchain configuration (e.g. Nix packages). The operator adds init containers to provision tools.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentToolchain
metadata:
  name: nix-tools
spec:
  type: nix
  nix:
    packages:
      - nodejs_22
      - python3
      - ripgrep
      - jujutsu
```

#### Full spec reference

| Field | Type | Description |
| --- | --- | --- |
| `type` | string | Toolchain type (`nix`) |
| `nix.packages` | list | Nixpkgs attribute names to install |
| `nix.image` | string | Nix container image (default: `nixos/nix:latest`) |

When `nix` is configured, the operator:
1. Adds an emptyDir volume (`nix-env`)
2. Adds a `nix-env` init container that bootstraps the Nix store and installs packages via `nix profile install`
3. Mounts the volume at `/nix` in the main container
4. Prepends `/nix/var/nix/profiles/agent/bin` to `PATH`

Package names are [nixpkgs](https://search.nixos.org/packages) attributes, the same names you'd use with `nix profile install nixpkgs#<name>`.

## Installation

### Prerequisites

- Kubernetes cluster (v1.28+)
- `kubectl` configured to access the cluster
- `kustomize` (or `kubectl` with built-in kustomize)

### Install CRDs

```bash
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/crd
```

### Deploy the operator

```bash
# Deploy with kustomize (pulls from GHCR)
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/default
```

The manager deployment uses `ghcr.io/xonovex/agent-operator-go:latest`.

To build locally:

```bash
docker build -f packages/agent/agent-operator-go/Dockerfile -t ghcr.io/xonovex/agent-operator-go:latest .
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

The typical workflow: create a Secret, an AgentProvider, optionally an AgentHarness for defaults, then run an agent.

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
  authTokenSecretRef:
    name: gemini-credentials
    key: api-key
  environment:
    ANTHROPIC_BASE_URL: "http://litellm-proxy:8317"
```

```yaml
# 3. (Optional) Create a harness with defaults
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentHarness
metadata:
  name: claude-harness
spec:
  type: claude
  defaultProvider: gemini-provider
  defaultTimeout: 1h
```

```yaml
# 4. Run the agent (referencing harness and provider)
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: fix-auth-bug
spec:
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: feature/auth
    storageSize: 10Gi
  prompt: "Fix the authentication bug in the login handler"
  timeout: 30m
```

```bash
kubectl apply -f provider.yaml -f harness.yaml -f run.yaml
kubectl get agentproviders
# NAME              DISPLAY NAME     READY   AGE
# gemini-provider   Google Gemini    true    5s

kubectl get agentruns -w
# NAME           PHASE         AGE
# fix-auth-bug   Pending       0s
# fix-auth-bug   Initializing  1s
# fix-auth-bug   Running       5s
# fix-auth-bug   Succeeded     45s
```

### Sandboxed agent run (gVisor)

Run agents inside a gVisor sandbox for syscall-level isolation. Requires the `gvisor` RuntimeClass to be configured on your cluster.

```yaml
# RuntimeClass (cluster setup, once per cluster)
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
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: git
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
# RuntimeClass (cluster setup, once per cluster)
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
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: main
  prompt: "Implement the payment processing module"
  runtimeClassName: kata
  timeout: 1h
```

### Confidential Computing (Kata + CoCo)

Run agents inside a Trusted Execution Environment (TEE) using [Confidential Containers (CoCo)](https://github.com/confidential-containers) with AMD SEV-SNP or Intel TDX hardware. This works across cloud providers (AKS, EKS, GKE) and on-prem clusters with TEE-capable nodes.

Use the existing `runtimeClassName` and `nodeSelector` fields — no special configuration is needed:

```yaml
# RuntimeClass (cluster setup — cloud providers may create these automatically)
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: kata-cc
handler: kata-cc
---
# AgentRun with AMD SEV-SNP confidential computing
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: confidential-agent
spec:
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: main
  prompt: "Process sensitive data"
  runtimeClassName: kata-cc
  nodeSelector:
    # Use your cluster's label for TEE-capable nodes. Examples:
    # AKS: kubernetes.azure.com/confidential-computing: "true"
    # Generic: node.kubernetes.io/tee: "sev-snp"
    kubernetes.azure.com/confidential-computing: "true"
```

Common runtime classes: `kata-cc` (AMD SEV-SNP), `kata-tdx` (Intel TDX). The exact names depend on your cluster's CoCo installation. Use `nodeSelector` to target nodes with TEE hardware. Both fields can also be set as defaults on an AgentHarness via `defaultRuntimeClassName`.

### Sandbox default via harness

Use AgentHarness to provide sandboxed runtime defaults that runs inherit via `harnessRef`:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentHarness
metadata:
  name: sandboxed-harness
spec:
  type: claude
  defaultProvider: gemini-provider
  defaultRuntimeClassName: gvisor
---
# This run inherits runtimeClassName=gvisor from the harness
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: auto-sandboxed
spec:
  harnessRef: sandboxed-harness
  workspace:
    type: git
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
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: jj
    repository:
      url: https://github.com/org/repo.git
      branch: main
  prompt: "Refactor the error handling"
```

The agent container image must include the `jj` binary. When workspace type is `jj`:
- **Standalone clone**: `git clone ... && jj git init --colocate`
- **Standalone worktree**: `jj workspace add` instead of `git worktree add`
- **Workspace init**: clone + `jj git init --colocate`
- **Workspace worktree**: `jj workspace add` instead of `git worktree add`

### Agent run with Nix packages

Provision reproducible tool environments using [Nix](https://nixos.org/). Use an AgentToolchain or inline toolchain config.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: nix-agent
spec:
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: main
  toolchain:
    type: nix
    nix:
      packages:
        - nodejs_22
        - python3
        - ripgrep
        - jujutsu
  prompt: "Set up the CI pipeline"
```

Or reference a reusable AgentToolchain:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentToolchain
metadata:
  name: nix-tools
spec:
  type: nix
  nix:
    packages:
      - nodejs_22
      - python3
      - git
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: nix-agent
spec:
  harnessRef: claude-harness
  providerRef: gemini-provider
  toolchainRef: nix-tools
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
  prompt: "Fix the build"
```

### Combined: Nix + jj + sandbox

All features compose. An agent can use Nix toolchain, Jujutsu VCS, and gVisor sandboxing simultaneously:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: full-stack-agent
spec:
  harnessRef: claude-harness
  providerRef: gemini-provider
  workspace:
    type: jj
    repository:
      url: https://github.com/org/repo.git
  toolchain:
    type: nix
    nix:
      packages:
        - nodejs_22
        - python3
        - postgresql
  prompt: "Implement the payment processing module"
  runtimeClassName: gvisor
```

### Multi-agent shared workspace

Create a workspace with shared volumes, then launch concurrent agents. Each agent gets an isolated git worktree but shares the same checkout and config directories.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: my-workspace
spec:
  type: git
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
  harnessRef: claude-harness
  workspaceRef: my-workspace
  workspace:
    git:
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
  harnessRef: claude-harness
  workspaceRef: my-workspace
  workspace:
    git:
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
# NAME      PHASE     AGE
# agent-1   Running   15s
# agent-2   Running   15s
```

### Multi-agent shared workspace with jj

Combine workspace-based runs with Jujutsu. The workspace init job clones with git and initializes jj colocated. Each agent run creates a jj workspace instead of a git worktree.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentWorkspace
metadata:
  name: jj-workspace
spec:
  type: jj
  repository:
    url: https://github.com/org/repo.git
    branch: main
  storageClass: nfs-csi
  storageSize: 10Gi
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: jj-agent-1
spec:
  harnessRef: claude-harness
  workspaceRef: jj-workspace
  workspace:
    type: jj
  providerRef: gemini-provider
  prompt: "Implement the search feature"
```

### Inline harness (no AgentHarness resource needed)

For one-off runs, specify the harness inline:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: quick-run
spec:
  harness:
    type: claude
    defaultImage: "node:22-slim"
    defaultTimeout: 30m
    defaultRuntimeClassName: gvisor
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
  prompt: "Fix the tests"
```

### Inline provider (no AgentProvider resource needed)

For one-off runs, specify the provider inline:

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: quick-review
spec:
  harnessRef: claude-harness
  provider:
    type: anthropic
    authSecretRef:
      name: anthropic-credentials
      key: api-key
    environment:
      ANTHROPIC_BASE_URL: "https://api.anthropic.com"
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: main
  prompt: "Review PR changes"
```

### Monitoring runs

```bash
# Watch AgentRun status
kubectl get agentruns -w
# NAME           PHASE      AGE
# fix-auth-bug   Running    30s

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

# Integration tests (envtest, real API server, no kubelet)
# Requires: setup-envtest (go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)
KUBEBUILDER_ASSETS=$(setup-envtest use -p path) go test -tags=integration -v -timeout=300s ./test/integration/

# E2E tests (Kind, full cluster with scheduling and garbage collection)
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

- **Unit (68 tests):** Builders (PVC, Job, containers, env vars, workspace PVC/Job/worktree), webhooks (defaulting and validation for all CRDs including workspaceRef rules), resolvers (harness, provider, workspace, toolchain).
- **Integration (20 tests):** Reconciler logic against a real API server. PVC/Job creation, phase transitions (Running, Succeeded, Failed, TimedOut), provider resolution, AgentHarness defaults, terminal phase skipping, AgentWorkspace PVC creation, workspace Ready/Failed transitions, AgentRun with workspaceRef waiting for workspace Ready.
- **E2E (7 tests):** Full cluster behavior. Pod scheduling, PVC binding, init container failure propagation, owner reference garbage collection, Docker image deployment with health probe validation, multi-agent workspace with concurrent runs, full-cycle pipeline (git clone + fake agent binary -> Succeeded).
- **E2E gVisor (5 tests):** Sandbox isolation verification (dmesg gVisor banner), runtimeClassName propagation to Job/Pod, AgentHarness default inheritance, full workflow (Secret + Provider + Harness + git clone + agent -> Succeeded inside gVisor), workspace-based run (init Job has no runtimeClassName, agent Job does).
- **E2E Kata (4 tests):** VM isolation verification (guest kernel differs from host, /dev/pmem0), runtimeClassName propagation to Job/Pod, AgentHarness default inheritance, full workflow (Secret + Provider + Harness + git clone + agent -> Succeeded inside Kata VM, skips in unprivileged kind).

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

**RuntimeClassName** is applied to the Job's PodSpec when set on the AgentRun or inherited from the referenced AgentHarness. Both init and main containers run in the sandboxed runtime. Workspace init Jobs do *not* inherit runtimeClassName; only agent Jobs do.

```
Standalone:                         Workspace:

AgentRun                            AgentWorkspace
    |                                   |
    +-> AgentHarness (via harnessRef)  +-> PVCs (RWX): workspace + shared volumes
    +-> AgentProvider -> Secret        +-> Init Job (git clone) -> Ready
    +-> AgentToolchain (optional)      |
    |                              AgentRun (workspaceRef)
    +-> PVC (RWO)                      |
    +-> Job                            +-> AgentHarness (via harnessRef)
          +-> Init: git clone          +-> AgentProvider -> Secret
          +-> Main: agent binary       +-> AgentToolchain (optional)
          +-> runtimeClassName?        |
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
kubectl delete agentworkspace my-workspace

# Uninstall the operator
kubectl delete -k config/default/

# Remove CRDs (deletes all resources)
kubectl delete -k config/crd/
```
