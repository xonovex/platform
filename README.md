# Xonovex Platform Monorepo

![License](https://img.shields.io/badge/license-MIT-blue)
![Node](https://img.shields.io/badge/node-20%2B-green)
![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8)

> Monorepo for Xonovex AI agent tools, workflows, and skills

AI coding agents handle prompts, tools, and code changes. What they don't manage is the environment around them: sandbox isolation, model provider routing, terminal sessions, reproducible toolchains, and orchestration at scale.

Xonovex fills this gap. It currently supports [Claude Code](https://docs.anthropic.com/en/docs/claude-code) and [OpenCode](https://github.com/anomalyco/opencode) as agents, with sandboxing via bubblewrap and Docker, VM-level isolation via gVisor and Kata Containers, confidential computing via [Confidential Containers (CoCo)](https://github.com/confidential-containers) with AMD SEV-SNP and Intel TDX, model routing through providers like Gemini, GLM, and GPT, workspace management with Git and [Jujutsu](https://github.com/jj-vcs/jj), reproducible toolchains via Nix, and Kubernetes orchestration for running agents at scale.

The included skills are token-efficient, harness-neutral, and based on current research and best practices (Agent Skills spec, agentskills.io, agents.md).

- **[agent-cli-go](packages/agent/agent-cli-go/)** configures sandboxes, providers, and terminal sessions, then launches the agent
- **[agent-operator-go](packages/agent/agent-operator-go/)** orchestrates agents as Kubernetes Jobs with managed workspaces, provider secrets, shared multi-agent workspaces, namespace-level policy enforcement, network isolation, and Nix toolchain provisioning
- **[Skills](packages/skill/)** give agents coding guidelines they follow automatically; plan-driven development with worktrees, project-instruction management, insight extraction, and skill authoring all live here as consolidated skill packages

## Quick Start

### Agent CLI

```bash
npm install -g @xonovex/agent-cli-go
agent-cli run --agent claude --sandbox bwrap --provider gemini
```

### Agent Kubernetes Operator

```bash
# Install CRDs and deploy the operator
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/crd
kubectl apply -k https://github.com/xonovex/platform//packages/agent/agent-operator-go/config/default

# Create a provider and run an agent
kubectl create secret generic gemini-credentials --from-literal=api-key='your-key'
kubectl apply -f - <<EOF
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentProvider
metadata:
  name: gemini-provider
spec:
  displayName: Google Gemini
  authTokenSecretRef:
    name: gemini-credentials
    key: api-key
---
apiVersion: agent.xonovex.com/v1alpha1
kind: AgentRun
metadata:
  name: review-code
spec:
  harness:
    type: claude
  providerRef: gemini-provider
  workspace:
    type: git
    repository:
      url: https://github.com/org/repo.git
      branch: main
  prompt: "Review the codebase and suggest improvements"
EOF
```

### Agent Plugins

Each skill is a separate plugin. Skills are applied **automatically** when the agent detects a relevant task — no explicit slash-command invocation needed. Skills also work cross-harness (Claude Code, OpenCode, etc.) since they follow the [Agent Skills spec](https://agentskills.io/specification).

#### Claude Code

```bash
# Add the Xonovex plugin marketplace
claude plugin marketplace add xonovex/platform

# Install workflow skills — each covers a full operation lifecycle via consolidated references
claude plugin install xonovex-skill-plan@xonovex-marketplace          # research, plan, refine, subplans, continue, update, validate, code-align/harden/simplify
claude plugin install xonovex-skill-git@xonovex-marketplace           # commit, merge-resolve, feature-worktree create/merge/abandon/cleanup
claude plugin install xonovex-skill-instruction@xonovex-marketplace   # AGENTS.md init / sync / simplify / consolidate / merge
claude plugin install xonovex-skill-insights@xonovex-marketplace      # session retrospective: extract, integrate-instructions, integrate-skills
claude plugin install xonovex-skill-prompt@xonovex-marketplace        # author / merge / simplify reusable prompt files (cross-harness format reference)
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

# Install workflow skills — each covers a full operation lifecycle via consolidated references
codex plugin add xonovex-skill-plan@xonovex-marketplace          # research, plan, refine, subplans, continue, update, validate, code-align/harden/simplify
codex plugin add xonovex-skill-git@xonovex-marketplace           # commit, merge-resolve, feature-worktree create/merge/abandon/cleanup
codex plugin add xonovex-skill-instruction@xonovex-marketplace   # AGENTS.md init / sync / simplify / consolidate / merge
codex plugin add xonovex-skill-insights@xonovex-marketplace      # session retrospective: extract, integrate-instructions, integrate-skills
codex plugin add xonovex-skill-prompt@xonovex-marketplace        # author / merge / simplify reusable prompt files (cross-harness format reference)
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

## Development

```bash
git clone https://github.com/xonovex/platform.git
cd platform && npm install
```

Tasks are managed with [Moon](https://moonrepo.dev/):

```bash
npx moon run <project>:<task>    # run a specific task
npx moon run :<task>             # run task across all projects
moon query projects              # list all projects
```

## License

MIT

---

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.
