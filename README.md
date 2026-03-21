# Xonovex Platform Monorepo

![License](https://img.shields.io/badge/license-MIT-blue)
![Node](https://img.shields.io/badge/node-20%2B-green)
![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8)

> Monorepo for Xonovex AI agent tools, workflows, and skills

AI coding agents handle prompts, tools, and code changes. What they don't manage is the environment around them: sandbox isolation, model provider routing, terminal sessions, reproducible toolchains, and orchestration at scale.

Xonovex fills that gap. It configures the execution context before the agent launches, then gets out of the way.

## Packages

### Agent

| Package | Description |
|---------|-------------|
| [agent-cli](packages/agent/agent-cli/) | Unified CLI for running agents with sandbox, provider, and terminal options |
| [agent-cli-go](packages/agent/agent-cli-go/) | Go implementation of agent-cli (cross-platform binaries) |
| [agent-operator-go](packages/agent/agent-operator-go/) | Kubernetes operator for running agents as Jobs with managed workspaces |

### Commands

| Package | Description |
|---------|-------------|
| [command-workflow](packages/command/command-workflow/) | Plan-driven development workflow with worktrees and parallel execution |
| [command-utility](packages/command/command-utility/) | Utility commands for content, instructions, insights, and skills |

### Skills

| Package | Description |
|---------|-------------|
| [skill-general-fp](packages/skill/skill-general-fp/) | Functional programming guidelines |
| [skill-general-oop](packages/skill/skill-general-oop/) | Object-oriented programming guidelines |
| [skill-typescript](packages/skill/skill-typescript/) | TypeScript guidelines |
| [skill-python](packages/skill/skill-python/) | Python guidelines |
| [skill-c99](packages/skill/skill-c99/) | C99 guidelines |
| [skill-lua](packages/skill/skill-lua/) | Lua guidelines |
| [skill-react](packages/skill/skill-react/) | React patterns |
| [skill-docker](packages/skill/skill-docker/) | Docker best practices |
| [skill-kubernetes](packages/skill/skill-kubernetes/) | Kubernetes patterns |
| [skill-terraform](packages/skill/skill-terraform/) | Terraform patterns |
| [skill-git](packages/skill/skill-git/) | Git workflow guidelines |
| [skill-shell-scripting](packages/skill/skill-shell-scripting/) | Shell scripting guidelines |
| [skill-vitest](packages/skill/skill-vitest/) | Vitest testing patterns |
| [skill-zod](packages/skill/skill-zod/) | Zod schema validation patterns |
| [skill-sql-postgresql](packages/skill/skill-sql-postgresql/) | PostgreSQL and SQL guidelines |
| [skill-threejs](packages/skill/skill-threejs/) | Three.js 3D graphics patterns |

## Quick Start

### Agent CLI

```bash
# TypeScript
npm install -g @xonovex/agent-cli
agent-cli run --agent claude --sandbox bwrap

# Go (cross-platform binary)
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

### Claude Code Plugins

Add the marketplace, then install workflow commands and skills:

```bash
# Add the Xonovex plugin marketplace
claude plugin marketplace add xonovex/platform

# Install workflow commands (plan, code quality, git)
claude plugin install xonovex-workflow@platform

# Install utility commands (instructions, insights, skills)
claude plugin install xonovex-utility@platform

# Install skills (each skill is a separate plugin)
claude plugin install xonovex-typescript@platform
claude plugin install xonovex-react@platform
claude plugin install xonovex-general-fp@platform
```

Once installed, workflow commands are available as slash commands in Claude Code:

```
/plan-research          Research codebase and web for requirements
/plan-create            Create a high-level plan for user review
/plan-subplans-create   Generate detailed subplans with parallel execution detection
/plan-worktree-create   Create a git worktree for a feature branch
/plan-continue          Resume work from an existing plan
/plan-validate          Verify that a plan or current work has been fully achieved
/plan-update            Update plan status and test results
/plan-refine            Process user annotations and refine iteratively
/plan-worktree-merge    Merge feature worktree back to source
/code-simplify          Consolidate duplicates, remove dead code, flatten abstractions
/code-harden            Improve type safety, validation, and error handling
/code-align             Align two similar implementations and suggest improvements
/git-commit             Commit and push changes
```

Skills are applied automatically when relevant to the task.

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
