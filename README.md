# Xonovex Platform Monorepo

![License](https://img.shields.io/badge/license-MIT-blue) ![Node](https://img.shields.io/badge/node-22.18%2B-green) ![Go](https://img.shields.io/badge/go-1.26%2B-00ADD8)

> Run AI coding agents with explicit sandbox, provider, workspace, toolchain, and orchestration controls.

Use the Agent CLI for local runs, the Agent Operator for policy-governed Kubernetes Jobs, and the plugin catalog for reusable commands and skills.

Xonovex supports [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [OpenCode](https://github.com/anomalyco/opencode). It provides bubblewrap and Docker sandboxes, gVisor and Kata Containers isolation, [Confidential Containers (CoCo)](https://github.com/confidential-containers) with AMD SEV-SNP and Intel TDX, model routing through providers such as Gemini, GLM, and GPT, Git and [Jujutsu](https://github.com/jj-vcs/jj) workspaces, Nix toolchains, and Kubernetes orchestration.

The included skills are token-efficient, harness-neutral, and based on current research and best practices (Agent Skills spec, agentskills.io, agents.md). Skills provide instructions, references, scripts, and setup capabilities; installing one is not proof that a policy executes or blocks an action.

- **[agent-cli-go](packages/tooling/cli/agent-cli-go/)** configures sandboxes, providers, and terminal sessions, then launches the agent
- **[agent-operator-go](packages/sandbox/operator/agent-operator-go/)** orchestrates agents as Kubernetes Jobs with managed workspaces, provider secrets, shared multi-agent workspaces, namespace-level policy enforcement, network isolation, and Nix toolchain provisioning
- **[moon-nix-toolchain](packages/tooling/moon/moon-nix-toolchain/)** wraps every Moon task in the repository's Nix flake dev shell, giving reproducible flake-pinned toolchains in local runs, pre-commit hooks, and CI
- **[Plugins](packages/runtime/plugin/)** give agents coding guidelines they follow automatically; plan-driven development with worktrees, project-instruction management, insight extraction, and skill authoring live in grouped plugins

## Quick Start

Choose the local Agent CLI, Kubernetes operator, or plugin installation path for the required use case.

### Agent CLI

Install the CLI and start a Claude Code run in a bubblewrap sandbox with the Gemini provider.

```bash
npm install -g @xonovex/agent-cli-go
agent-cli run --agent claude --isolation bwrap --provider gemini
```

Select the sandbox with three independent axes: `--isolation {none,bwrap,docker}`, `--provision {none,nix,command}`, and `--network {host,none,proxy}`. See `packages/tooling/cli/AGENTS.md` for the complete model.

![Three Claude Code agents in a tiled tmux session, each in its own git worktree: one under bwrap, one under bwrap with a Nix-provisioned toolchain, and one under Docker, routed to two different model providers](packages/asset/asset-images/multiple-agents.png)

Each pane is a separate worktree with its own axis combination and provider, so concurrent agents do not share a checkout or sandbox.

### Agent Kubernetes Operator

Install the operator only after the cluster has a digest-pinned agent image and a sandboxed RuntimeClass such as gVisor or Kata. Set these variables to values available in the cluster.

```bash
export XONOVEX_AGENT_IMAGE='ghcr.io/your-org/xonovex-agent@sha256:<64-hex-digest>'
export XONOVEX_RUNTIME_CLASS='gvisor'

# Requires cert-manager v1.16+ with its CA injector enabled
# Install CRDs and deploy the operator
kubectl apply -k https://github.com/xonovex/platform//packages/sandbox/operator/agent-operator-go/config/crd
kubectl apply -k https://github.com/xonovex/platform//packages/sandbox/operator/agent-operator-go/config/default

# Create one policy-governed namespace and provider credential
kubectl create namespace ai-agents --dry-run=client -o yaml | kubectl apply -f -
kubectl -n ai-agents create secret generic anthropic-credentials \
  --from-literal=api-key='your-key' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f - <<EOF
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentPolicy
metadata:
  name: sandbox-policy
  namespace: ai-agents
spec:
  enforced:
    runtimeClassName: ${XONOVEX_RUNTIME_CLASS}
    requireSecurityContext: true
    requireNetworkPolicy: true
    maxTimeout: 1h0m0s
    maxResources:
      cpu: "2"
      memory: 4Gi
    allowedImages:
      - ${XONOVEX_AGENT_IMAGE}
    allowedRuntimeClassNames:
      - ${XONOVEX_RUNTIME_CLASS}
    allowedSecretNames:
      - anthropic-credentials
  defaults:
    image: ${XONOVEX_AGENT_IMAGE}
    runtimeClassName: ${XONOVEX_RUNTIME_CLASS}
    timeout: 30m0s
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentProvider
metadata:
  name: anthropic-provider
  namespace: ai-agents
spec:
  displayName: Anthropic Claude
  authTokenSecretRef:
    name: anthropic-credentials
    key: api-key
  authTokenEnv: ANTHROPIC_API_KEY
  environment:
    ANTHROPIC_BASE_URL: https://api.anthropic.com
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: review-code
  namespace: ai-agents
spec:
  harness:
    type: claude
  providerRef: anthropic-provider
  workspace:
    type: git
    repository:
      url: https://github.com/xonovex/platform.git
      branch: main
  prompt: "Review the codebase and suggest improvements"
  network: host
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: "2"
      memory: 2Gi
EOF
```

The `network: host` value permits unrestricted egress so this example can reach the public model API. Production namespaces should use an enforceable cluster-level egress proxy or a fully qualified domain name aware policy.

### Agent Plugins

Install the grouped plugins your tasks need. Each plugin owns related skills and commands. A skill keeps its existing name, and a command uses the namespace of the plugin that owns it.

| Plugin                                                                          | Purpose                                                                                                       |
| ------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| [`xonovex-agentic`](packages/runtime/plugin/plugin-agentic/README.md)           | Configure coding-agent harnesses and author their skills, commands, and project instructions.                 |
| [`xonovex-core`](packages/runtime/plugin/plugin-core/README.md)                 | Apply software design, testing, review, version control, credentials, and accessibility practices.            |
| [`xonovex-game-engine`](packages/runtime/plugin/plugin-game-engine/README.md)   | Build game engines, renderers, audio systems, editors, asset pipelines, and multiplayer networking.           |
| [`xonovex-integrations`](packages/runtime/plugin/plugin-integrations/README.md) | Use GitHub and GitLab for issues, pull requests, delivery, and continuous integration.                        |
| [`xonovex-languages`](packages/runtime/plugin/plugin-languages/README.md)       | Write Python, shell, SQL, and Lua, including TypeScript compiled to Lua.                                      |
| [`xonovex-native`](packages/runtime/plugin/plugin-native/README.md)             | Write C99 and build portable native systems with explicit memory, concurrency, and data layouts.              |
| [`xonovex-platform`](packages/runtime/plugin/plugin-platform/README.md)         | Configure Moon tasks and build and operate Docker images, Kubernetes workloads, and Terraform infrastructure. |
| [`xonovex-typescript`](packages/runtime/plugin/plugin-typescript/README.md)     | Develop the TypeScript stack, its frameworks, package tooling, validation, and tests.                         |
| [`xonovex-workflow`](packages/runtime/plugin/plugin-workflow/README.md)         | Research, plan, implement, validate, and close out work, then integrate lessons from the session.             |
| [`xonovex-writing`](packages/runtime/plugin/plugin-writing/README.md)           | Revise prose and write technical documents, articles, news, and travel guides.                                |

#### Claude Code

Add the marketplace, then install the plugins needed for the task:

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-core@xonovex-marketplace
claude plugin install xonovex-workflow@xonovex-marketplace
claude plugin install xonovex-typescript@xonovex-marketplace
```

#### Codex

Add the marketplace, then install the plugins needed for the task:

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-core@xonovex-marketplace
codex plugin add xonovex-workflow@xonovex-marketplace
codex plugin add xonovex-typescript@xonovex-marketplace
```

#### Upgrade from individual plugins

The grouped plugins replace the individual `xonovex-skill-*` plugins and `xonovex-utility`. Install the group that owns each skill, then uninstall its old individual plugin to prevent duplicate skill registration. The `xonovex-workflow` plugin now contains the planning and reflection skills and their commands. Git commands and the existing `plan-worktree-*` commands belong to `xonovex-core`; authoring commands belong to `xonovex-agentic`; content commands belong to `xonovex-writing`.

## Development

Install workspace dependencies before running Moon tasks and repository gates.

```bash
git clone https://github.com/xonovex/platform.git
cd platform && npm install
```

[Moon](https://moonrepo.dev/) manages project tasks.

```bash
npx moon run <project>:<task>    # run a specific task
npx moon run :<task>             # run a single-colon task across matching projects
npm run fmt:check                # run the repository format-check aggregate
npx moon query projects          # list all projects
```

## License

The repository uses the MIT License.

---

See [CONTRIBUTING.md](CONTRIBUTING.md) for the complete development setup and contribution guidelines.
