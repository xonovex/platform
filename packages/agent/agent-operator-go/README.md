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

#### Submission boundary

`AgentRun` is the operator's execution request API. Manual tools, harness hooks,
CI/CD systems, webhook handlers, Kubernetes CronJobs, and other external
integrations create `AgentRun` resources through the Kubernetes API. The
operator has no scheduling or event-ingress listener and does not interpret
caller-specific trigger metadata.

The operator admits and reconciles the submitted run specification. Any
additional approval, policy, evidence, or escalation requirements belong to
the caller or the native platform that submits the run.

Every execution namespace must contain exactly one `AgentPolicy`. Admission rejects AgentRuns and AgentWorkspaces when the policy is absent or ambiguous, because the operator cannot safely determine runtime-class or Secret authority otherwise.

#### Full spec reference

| Field              | Type     | Description                                                                                       |
| ------------------ | -------- | ------------------------------------------------------------------------------------------------- |
| `harnessRef`       | string   | Name of an AgentHarness in the same namespace                                                     |
| `harness`          | object   | Inline harness config (mutually exclusive with `harnessRef`)                                      |
| `providerRef`      | string   | Name of an AgentProvider in the same namespace                                                    |
| `provider`         | object   | Inline provider config (mutually exclusive with `providerRef`)                                    |
| `workspaceRef`     | string   | Name of an AgentWorkspace for shared workspace support                                            |
| `workspace`        | object   | Inline workspace config (mutually exclusive with `workspaceRef`)                                  |
| `toolchainRef`     | string   | Name of an AgentToolchain in the same namespace                                                   |
| `toolchain`        | object   | Inline toolchain config (mutually exclusive with `toolchainRef`)                                  |
| `prompt`           | string   | Task prompt for headless execution                                                                |
| `resources`        | object   | K8s resource requirements applied to the agent and its init containers                            |
| `timeout`          | duration | Positive max run duration (default: `1h`)                                                         |
| `env`              | list     | Additional environment variables; Secret refs require policy allowlisting                         |
| `image`            | string   | Digest-pinned agent image override; required unless resolved from a harness, toolchain, or policy |
| `runtimeClassName` | string   | Sandboxed pod runtime class; required unless resolved from a harness or policy                    |
| `nodeSelector`     | map      | Node selector for pod scheduling                                                                  |
| `tolerations`      | list     | Tolerations for pod scheduling                                                                    |

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
  defaultImage: "ghcr.io/example/agents/claude@sha256:<digest>"
  defaultRuntimeClassName: gvisor
  defaultTimeout: 1h
  env:
    - name: LANG
      value: "en_US.UTF-8"
```

#### Full spec reference

| Field                     | Type     | Description                                       |
| ------------------------- | -------- | ------------------------------------------------- |
| `type`                    | string   | Agent type (`claude`, `opencode`)                 |
| `defaultProvider`         | string   | Default provider name                             |
| `defaultImage`            | string   | Default digest-pinned agent image                 |
| `defaultResources`        | object   | Default resource requirements                     |
| `defaultTimeout`          | duration | Default timeout for agent runs                    |
| `defaultRuntimeClassName` | string   | Default pod runtime class (e.g. `gvisor`, `kata`) |
| `env`                     | list     | Default environment variables                     |

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

The controller validates that the referenced Secret exists and contains the specified key, reporting readiness via `.status.ready`. Using the provider from an AgentRun also requires the Secret name in `AgentPolicy.spec.enforced.allowedSecretNames`.

#### Full spec reference

| Field                | Type   | Description                                |
| -------------------- | ------ | ------------------------------------------ |
| `type`               | string | Provider type (e.g. `anthropic`, `openai`) |
| `displayName`        | string | Human-readable name                        |
| `authTokenSecretRef` | object | Secret reference for auth token            |
| `environment`        | map    | Environment variables to set               |
| `cliArgs`            | list   | Additional CLI arguments                   |

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
    # credentialsSecretRef:
    #   name: repository-credentials
    #   key: credentials
  storageClass: nfs-csi
  storageSize: 10Gi
  runtimeClassName: gvisor
  sharedVolumes:
    - name: claude-config
      mountPath: /home/agent/.claude
      storageSize: 1Gi
    - name: opencode-config
      mountPath: /home/agent/.config/opencode
      storageSize: 512Mi
```

**Lifecycle phases:** `Pending` -> `Initializing` -> `Ready` | `Failed`

#### Full spec reference

| Field                         | Type   | Description                                        |
| ----------------------------- | ------ | -------------------------------------------------- |
| `type`                        | string | Workspace type (`git` or `jj`)                     |
| `repository.url`              | string | Git repository URL (required)                      |
| `repository.branch`           | string | Branch to checkout                                 |
| `repository.credentialsSecretRef` | object | Allowlisted Secret key containing one git-credential-store entry for private HTTPS clones |
| `storageClass`                | string | Storage class for workspace PVC (must support RWX) |
| `storageSize`                 | string | Storage size for workspace PVC (default: `10Gi`)   |
| `runtimeClassName`            | string | Sandboxed runtime for the clone Job; defaults from the namespace policy |
| `sharedVolumes[].name`        | string | Volume name (used as PVC suffix)                   |
| `sharedVolumes[].mountPath`   | string | Mount path in agent containers                     |
| `sharedVolumes[].storageSize` | string | PVC size for this volume (default: `1Gi`)          |
| `git.worktree`                | object | Git worktree configuration                         |
| `jj.revision`                 | string | Jujutsu revision                                   |

#### Volume layout

```
workspace PVC (RWX):
  /workspace/              <- main git checkout (from init Job)
  /workspace/.git/         <- shared .git dir
  /workspace-wt/agent-1/   <- worktree for agent-1
  /workspace-wt/agent-2/   <- worktree for agent-2

shared volume PVCs (RWX, one per sharedVolumes entry):
  /home/agent/.claude/           <- claude-config PVC
  /home/agent/.config/opencode/  <- opencode-config PVC
```

For a private HTTPS repository, the referenced Secret key contains a standard git credential-store line such as `https://username:token@github.com`. The key is mounted read-only into the clone container only; it is not mounted into the agent container. Keep the Secret encrypted at rest and out of source control.

### AgentToolchain

Reusable toolchain configuration. The `nix` toolchain provisions via a **pre-built, digest-pinned OCI image** (no per-pod install).

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentToolchain
metadata:
  name: nix-tools
spec:
  type: nix
  nix:
    nixpkgsRev: 49a4bd0573c376468dd7996ddb6f9fa31d8c4d97
    packages:
      - nodejs_24
      - ripgrep
    image: ghcr.io/xonovex/agent@sha256:<digest>
```

#### Full spec reference

| Field                        | Type   | Description                                                                                          |
| ---------------------------- | ------ | ---------------------------------------------------------------------------------------------------- |
| `type`                       | string | Toolchain type (`nix`)                                                                               |
| `nix.nixpkgsRev`             | string | Pinned nixpkgs rev the image was built from (required reproducibility pin)                           |
| `nix.packages`               | list   | Nixpkgs attribute names baked into the image (packages source; mutually exclusive with `flakeRef`)   |
| `nix.flakeRef` / `nix.shell` | string | Project flake + devShell (project-flake source; mutually exclusive with `packages`)                  |
| `nix.image`                  | string | Pre-built, digest-pinned agent OCI image the pod runs (required; satisfies `RequirePinnedProvision`) |

The `nix` toolchain selects the pre-built image as the pod image — the **same content-addressed store-path closure** the CLI resolves (built from the same `flake.lock` + `nix/agent-env.nix`, verified with `nix path-info -r`). The pod starts by image pull: **no `nix-env` emptyDir, no `nixos/nix` init container, no per-pod `nix profile install`**. The AgentRun and AgentToolchain webhooks reject a `NixSpec` without `nixpkgsRev`, exactly one packages/flake source, and an `@sha256:` image digest. Build/push the image with `npx moon run agent-operator-go:agent-image-build` (→ `nix build .#legacyPackages.<sys>.agentImage` + skopeo push).

### AgentPolicy

Namespace-scoped admission policy for AgentRuns and AgentWorkspaces. Exactly one AgentPolicy is required per execution namespace; missing or multiple policies reject admission because the effective authority would be unknown.

```yaml
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentPolicy
metadata:
  name: agent-policy
  namespace: ai-agents
spec:
  enforced:
    runtimeClassName: kata
    allowedRuntimeClassNames: [kata, gvisor]
    allowedSecretNames:
      - gemini-credentials
      - repository-credentials
    requireSecurityContext: true
    requireNetworkPolicy: true
    requireEgressRestricted: true
    maxTimeout: 1h
    maxResources:
      cpu: "2"
      memory: 4Gi
    allowedImages:
      - ghcr.io/example/agents/
  defaults:
    image: ghcr.io/example/agents/runtime@sha256:<digest>
    timeout: 30m
    runtimeClassName: kata
```

| Host policy intent  | Native admission behavior                                                                                           | Independent verification                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| Runtime isolation   | Requires or allowlists `runtimeClassName`                                                                           | Verify the cluster RuntimeClass and runtime implementation                |
| Container hardening | Rejects explicit privilege escalation and root weakening                                                            | Inspect the generated Pod security context and cluster admission policy   |
| Network restriction | Rejects disabled policies, unrestricted host networking, proxy mode without a backend, and unprovable custom egress | Verify generated NetworkPolicy behavior with the installed network plugin |
| Duration bound      | Requires an explicit/policy-defaulted timeout at or below `maxTimeout`                                              | Observe Job timeout and terminal status                                   |
| Resource bound      | Requires a limit for each `maxResources` entry; rejects requests/limits above it                                    | Keep namespace LimitRange and ResourceQuota as an independent control     |
| Image restriction   | Requires a digest-pinned image resolved from the run, harness, toolchain, or policy, then applies `allowedImages`   | Add signature/provenance admission when digest pinning is insufficient    |
| Secret authority    | Rejects env, provider, and repository Secret references whose names are not explicitly allowlisted                | Keep Kubernetes RBAC from granting direct Secret reads to run submitters  |
| Toolchain pinning   | AgentToolchain/inline Nix validation requires revision, source, and image digest                                    | Verify registry digest and the built closure provenance                   |

Policy defaults are applied before harness, provider, and toolchain references are resolved at admission. Referenced execution inputs are snapshotted inline, and the admitted AgentRun stores the exact image, runtime, resources, environment, and Secret references that policy approved.

Admission configuration and the webhook endpoint must be reachable for these controls to enforce. Verify installation with negative probes for a wrong runtime class, disabled or custom-open network policy, host or proxy network mode, invalid workspace storage quantity, excessive timeout/resource limit, disallowed or missing image, moving Nix image tag, policy API outage, and duplicate namespace policies. An accepted object is admission evidence only; it does not prove the runtime, network, registry, or quota layer behaved correctly.

## Installation

### Prerequisites

- Kubernetes cluster (v1.28+)
- `kubectl` configured to access the cluster
- `kustomize` (or `kubectl` with built-in kustomize)
- cert-manager v1.16+ with its CA injector enabled

### Install CRDs

```bash
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/crd
```

### Deploy the operator

```bash
# Deploy with kustomize (pulls from GHCR)
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/default
```

The default overlay deploys the admission service and uses cert-manager to issue
its serving certificate and inject the CA bundle into the webhook configurations.
The manager deployment uses the digest-pinned image declared in
`config/manager/manager.yaml`.

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

### Standalone agent run (direct submission)

Create a Secret, an AgentProvider, optionally an AgentHarness for defaults, then
submit an AgentRun directly.

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

Create a workspace with shared volumes, then launch concurrent agents. Each agent gets an isolated git worktree named from its AgentRun while sharing the same checkout and config directories.

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
  runtimeClassName: gvisor
  sharedVolumes:
    - name: claude-config
      mountPath: /home/agent/.claude
      storageSize: 1Gi
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: agent-1
spec:
  harnessRef: claude-harness
  workspaceRef: my-workspace
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
  runtimeClassName: gvisor
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: jj-agent-1
spec:
  harnessRef: claude-harness
  workspaceRef: jj-workspace
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
    defaultImage: "ghcr.io/example/agents/claude@sha256:<digest>"
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
# Requires: setup-envtest, provided by the nix devshell (nix/k8s.nix)
# The moon task wires KUBEBUILDER_ASSETS itself — no manual export needed:
npx moon run agent-operator-go:go-test-integration
# Direct invocation outside moon:
KUBEBUILDER_ASSETS="$(setup-envtest use -i -p path)" go test -tags=integration -v -timeout=300s ./test/integration/

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

# E2E CoCo tests (creates Kind cluster with simulated kata-cc/kata-tdx RuntimeClasses)
# Validates runtimeClassName propagation, harness defaults, full pipeline, workspace jobs
# Requires: kind, kubectl, Docker
go test -tags=e2e_coco -v -timeout=600s ./test/e2e-coco/
```

### What the tests cover

- **Unit:** Builders (PVC, Job, containers, env vars, workspace PVC/Job/worktree), policy/admission bypass cases, webhooks, resolvers, providers, and toolchains.
- **Integration:** Reconciler logic against a real API server, including resource creation, phase transitions, provider resolution, harness defaults, terminal phases, and shared workspaces.
- **E2E:** Full cluster scheduling, storage, initialization, owner cleanup, image deployment, multi-agent workspaces, and the full agent pipeline.
- **E2E gVisor/Kata/CoCo:** RuntimeClass propagation, harness defaults, complete runs, workspace paths, and runtime-specific isolation evidence where the environment supports it.

### Known host issue: Kind node exec fails (`setns process: exec: already started`)

On affected hosts every Kind-based e2e suite fails at cluster creation ("Writing
configuration") before running a single test, with:

```
OCI runtime exec failed: exec failed: unable to start container process: error starting setns process: exec: already started
```

Observed 2026-07-16 with docker 29.1.3 / runc 1.4.0 on kernel 7.0.6-gentoo
(cgroup v2, nsdelegate). Findings from the investigation:

- `docker exec` into a plain container (alpine) works, with or without
  `--privileged`.
- `docker exec` into a Kind node container fails identically with or without
  `--privileged`, on both `kindest/node:v1.35.0` and `v1.31.9` — the node-image
  version is irrelevant.
- The node container runs systemd in a private cgroup namespace: its PID 1
  sits in the child cgroup `init.scope` while controllers are enabled on the
  container's root cgroup, so attaching an exec'd process to that root cgroup
  violates cgroup v2's internal-process constraint. The `exec: already started`
  text is Go's `os/exec` error for a second `Start` on one `Cmd`, which places
  the fault in runc's exec/cgroup-attach retry path, not in the kernel setns
  call itself.
- Best hypothesis: a runc 1.4.0 exec bug triggered by systemd containers on
  this kernel/cgroup layout. Portage carries runc 1.4.2 (stable) and 1.4.3
  (~testing); upgrading and restarting docker is the untested candidate fix —
  it needs root, so it was not applied.
- Separately, the kata suite extracts a multi-GB Kata release into `$TMPDIR`;
  on a small tmpfs `/tmp` it fails at `tar` (ENOSPC) before reaching Kind. Set
  `TMPDIR` to a disk-backed path to get past that and reach the same Kind
  failure as the other suites.

Working alternative: `.github/workflows/e2e.yml` runs the integration suite
and all four e2e suites on a hosted runner. It is dormant until it reaches the
remote default branch; trigger it with `workflow_dispatch` or wait for its
nightly schedule.

## Architecture

Each AgentRun reconciles along one of two paths:

**Standalone path** (no `workspaceRef`):

1. **Workspace PVC** (RWO) is created for persistent git storage
2. **Job** is created with init container (git clone) and main container (agent binary)
3. Controller watches Job status and updates AgentRun phase

**Workspace path** (with `workspaceRef`):

1. **AgentWorkspace** must be in `Ready` phase (requeue if not)
2. **Job** is created using the workspace's shared PVC (RWX) with init container (git worktree add) and main container (agent binary working in the worktree)
3. Shared volume PVCs are mounted at configured paths (e.g. `~/.claude/`)
4. Controller watches Job status and updates AgentRun phase

**RuntimeClassName** is required for every execution pod. AgentRun Jobs resolve it from the run, harness, or namespace policy. AgentWorkspace clone Jobs resolve it from the workspace or namespace policy. Init and main containers therefore run inside the approved sandboxed runtime.

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
          +-> runtimeClassName         |
                                       +-> Job (uses workspace PVC)
                                             +-> Init: git worktree add
                                             +-> Main: agent binary
                                                   workingDir: /workspace-wt/{run}
                                                   mounts: shared volumes
                                             +-> runtimeClassName
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
