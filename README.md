# Xonovex Platform Monorepo

![License](https://img.shields.io/badge/license-MIT-blue)
![Node](https://img.shields.io/badge/node-22.18%2B-green)
![Go](https://img.shields.io/badge/go-1.26%2B-00ADD8)

> Monorepo for Xonovex AI agent tools, workflows, and skills

AI coding agents handle prompts, tools, and code changes. What they don't manage is the environment around them: sandbox isolation, model provider routing, terminal sessions, reproducible toolchains, and orchestration at scale.

Xonovex fills this gap. It currently supports [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [OpenCode](https://github.com/anomalyco/opencode) as agents, with sandboxing via bubblewrap and Docker, VM-level isolation via gVisor and Kata Containers, confidential computing via [Confidential Containers (CoCo)](https://github.com/confidential-containers) with AMD SEV-SNP and Intel TDX, model routing through providers like Gemini, GLM, and GPT, workspace management with Git and [Jujutsu](https://github.com/jj-vcs/jj), reproducible toolchains via Nix, and Kubernetes orchestration for running agents at scale.

The included skills are token-efficient, harness-neutral, and based on current research and best practices (Agent Skills spec, agentskills.io, agents.md). Skills provide instructions, references, scripts, and setup capabilities; installing one is not proof that a policy executes or blocks an action.

- **[agent-cli-go](packages/agent/agent-cli-go/)** configures sandboxes, providers, and terminal sessions, then launches the agent
- **[agent-operator-go](packages/agent/agent-operator-go/)** orchestrates agents as Kubernetes Jobs with managed workspaces, provider secrets, shared multi-agent workspaces, namespace-level policy enforcement, network isolation, and Nix toolchain provisioning
- **[moon-nix-toolchain](packages/moon/moon-nix-toolchain/)** wraps every Moon task in the repository's Nix flake dev shell, giving reproducible flake-pinned toolchains in local runs, pre-commit hooks, and CI
- **[Skills](packages/skill/)** give agents coding guidelines they follow automatically; plan-driven development with worktrees, project-instruction management, insight extraction, and skill authoring all live here as consolidated skill packages

## Quick Start

### Agent CLI

```bash
npm install -g @xonovex/agent-cli-go
agent-cli run --agent claude --isolation bwrap --provider gemini
```

The sandbox is selected by three orthogonal axes — `--isolation {none,bwrap,docker}` ×
`--provision {none,nix,command}` × `--network {host,none,proxy}` — see
`packages/agent/AGENTS.md`.

![Three Claude Code agents in a tiled tmux session, each in its own git worktree: one under bwrap, one under bwrap with a Nix-provisioned toolchain, and one under Docker, routed to two different model providers](packages/asset/asset-images/multiple-agents.png)

Each pane is a separate worktree with its own axis combination and provider, so
concurrent agents neither share a checkout nor a sandbox.

### Agent Kubernetes Operator

The operator requires a digest-pinned agent image and an installed sandboxed
RuntimeClass such as gVisor or Kata. Set these to values available in your cluster:

```bash
export XONOVEX_AGENT_IMAGE='ghcr.io/your-org/xonovex-agent@sha256:<64-hex-digest>'
export XONOVEX_RUNTIME_CLASS='gvisor'

# Requires cert-manager v1.16+ with its CA injector enabled
# Install CRDs and deploy the operator
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/crd
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/default

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

`network: host` permits unrestricted egress so the quick start can reach the
public model API. Production namespaces should replace it with an enforceable
cluster-level egress proxy or FQDN-aware policy.

### Agent Plugins

Each skill is a separate plugin. A compatible harness can route to an installed skill when its description matches the task; user-invocable commands explicitly load hard skill dependencies. Harness support, loading behavior, permissions, and native enforcement remain product-specific. The skills follow the [Agent Skills spec](https://agentskills.io/specification).

#### Claude Code

```bash
# Add the Xonovex plugin marketplace
claude plugin marketplace add xonovex/platform

# Install only the independent capabilities you need
claude plugin install xonovex-workflow@xonovex-marketplace            # explicit operations, context forwarding, publishing, and workspace commands
claude plugin install xonovex-skill-plan@xonovex-marketplace          # research, create, critique, revise, expand, continue, update, validate
claude plugin install xonovex-skill-git@xonovex-marketplace           # commit, merge-resolve, feature-worktree create/merge/abandon/cleanup
claude plugin install xonovex-skill-github@xonovex-marketplace        # GitHub issues, Projects, PRs/reviews, durable context, enforcement
claude plugin install xonovex-skill-gitlab@xonovex-marketplace        # GitLab issues/work items, boards, MRs/reviews, durable context, enforcement
claude plugin install xonovex-skill-instruction@xonovex-marketplace   # AGENTS.md init / sync / simplify / consolidate / merge
claude plugin install xonovex-skill-reflect@xonovex-marketplace       # session reflection: extract lessons, fold into AGENTS.md or a skill
claude plugin install xonovex-skill-code-review@xonovex-marketplace   # Conventional Comments feedback: blocking vs non-blocking, summary + inline
claude plugin install xonovex-skill-pull-request@xonovex-marketplace  # PR authoring: description, size/atomicity, testing evidence, tradeoffs, self-review
claude plugin install xonovex-skill-command@xonovex-marketplace        # author / merge / simplify reusable prompt files (cross-harness format reference)
claude plugin install xonovex-skill-skill@xonovex-marketplace         # author / extract / merge / simplify / validate Agent Skills
claude plugin install xonovex-skill-content@xonovex-marketplace       # multilingual articles, news, travel guides, prose humanization
claude plugin install xonovex-skill-llmstxt@xonovex-marketplace       # /llms.txt files and per-page markdown mirrors

# Install language / framework guides (apply automatically when editing those files)
claude plugin install xonovex-skill-typescript@xonovex-marketplace
claude plugin install xonovex-skill-react@xonovex-marketplace
claude plugin install xonovex-skill-hono@xonovex-marketplace
claude plugin install xonovex-skill-zod@xonovex-marketplace
claude plugin install xonovex-skill-vitest@xonovex-marketplace
# … see .claude-plugin/marketplace.json for the full list
```

#### Codex

```bash
# Add the Xonovex plugin marketplace
codex plugin marketplace add xonovex/platform

# Install only the independent capabilities you need
codex plugin add xonovex-skill-plan@xonovex-marketplace          # research, create, critique, revise, expand, continue, update, validate
codex plugin add xonovex-skill-git@xonovex-marketplace           # commit, merge-resolve, feature-worktree create/merge/abandon/cleanup
codex plugin add xonovex-skill-github@xonovex-marketplace        # GitHub issues, Projects, PRs/reviews, durable context, enforcement
codex plugin add xonovex-skill-gitlab@xonovex-marketplace        # GitLab issues/work items, boards, MRs/reviews, durable context, enforcement
codex plugin add xonovex-skill-instruction@xonovex-marketplace   # AGENTS.md init / sync / simplify / consolidate / merge
codex plugin add xonovex-skill-reflect@xonovex-marketplace       # session reflection: extract lessons, fold into AGENTS.md or a skill
codex plugin add xonovex-skill-code-review@xonovex-marketplace   # Conventional Comments feedback: blocking vs non-blocking, summary + inline
codex plugin add xonovex-skill-pull-request@xonovex-marketplace  # PR authoring: description, size/atomicity, testing evidence, tradeoffs, self-review
codex plugin add xonovex-skill-command@xonovex-marketplace        # author / merge / simplify reusable prompt files (cross-harness format reference)
codex plugin add xonovex-skill-skill@xonovex-marketplace         # author / extract / merge / simplify / validate Agent Skills
codex plugin add xonovex-skill-content@xonovex-marketplace       # multilingual articles, news, travel guides, prose humanization
codex plugin add xonovex-skill-llmstxt@xonovex-marketplace       # /llms.txt files and per-page markdown mirrors

# Install language / framework guides (apply automatically when editing those files)
codex plugin add xonovex-skill-typescript@xonovex-marketplace
codex plugin add xonovex-skill-react@xonovex-marketplace
codex plugin add xonovex-skill-hono@xonovex-marketplace
codex plugin add xonovex-skill-zod@xonovex-marketplace
codex plugin add xonovex-skill-vitest@xonovex-marketplace
# … see .agents/plugins/marketplace.json for the full list
```

### Upgrades and retired skill plugins

Upgrade the marketplace and selected plugins together; start a new session after changing an installation. Four pre-5.x plugin names are retired and must not remain installed alongside their replacements:

| Retired plugin              | Replacement             |
| --------------------------- | ----------------------- |
| `xonovex-skill-general-fp`  | `xonovex-skill-fp`      |
| `xonovex-skill-general-oop` | `xonovex-skill-oop`     |
| `xonovex-skill-insights`    | `xonovex-skill-reflect` |
| `xonovex-skill-prompt`      | `xonovex-skill-command` |

In Codex CLI, open `codex`, run `/plugins`, select each retired installed entry, and choose **Uninstall plugin**; then install its replacement. For Claude Code user-scope installations:

```bash
claude plugin uninstall xonovex-skill-general-fp@xonovex-marketplace
claude plugin uninstall xonovex-skill-general-oop@xonovex-marketplace
claude plugin uninstall xonovex-skill-insights@xonovex-marketplace
claude plugin uninstall xonovex-skill-prompt@xonovex-marketplace
```

## Development

```bash
git clone https://github.com/xonovex/platform.git
cd platform && npm install
```

Tasks are managed with [Moon](https://moonrepo.dev/):

```bash
npx moon run <project>:<task>    # run a specific task
npx moon run :<task>             # run a single-colon task across matching projects
npm run fmt:check                # run the repository format-check aggregate
npx moon query projects          # list all projects
```

## License

MIT

---

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.
