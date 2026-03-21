# Xonovex Platform Monorepo

![License](https://img.shields.io/badge/license-MIT-blue)
![Node](https://img.shields.io/badge/node-20%2B-green)
![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8)

> Monorepo for Xonovex AI agent tools, workflows, and skills

AI coding agents handle prompts, tools, and code changes. What they don't manage is the environment around them: sandbox isolation, model provider routing, terminal sessions, reproducible toolchains, and orchestration at scale. Xonovex fills this gap. What is in this monorepo:

- **[agent-cli-go](packages/agent/agent-cli-go/)** configures sandboxes, providers, and terminal sessions, then launches the agent
- **[agent-operator-go](packages/agent/agent-operator-go/)** orchestrates agents as Kubernetes Jobs with managed workspaces and provider secrets
- **[Workflow commands](packages/command/command-workflow/)** provide plan-driven development with worktrees and parallel execution
- **[Utility commands](packages/command/command-utility/)** manage project instructions, extract insights, and create skills
- **[Skills](packages/skill/)** give agents coding guidelines they follow automatically

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
/xonovex-workflow:plan-research          Research codebase and web for requirements
/xonovex-workflow:plan-create            Create a high-level plan for user review
/xonovex-workflow:plan-subplans-create   Generate detailed subplans with parallel execution detection
/xonovex-workflow:plan-worktree-create   Create a git worktree for a feature branch
/xonovex-workflow:plan-continue          Resume work from an existing plan
/xonovex-workflow:plan-validate          Verify that a plan or current work has been fully achieved
/xonovex-workflow:plan-update            Update plan status and test results
/xonovex-workflow:plan-refine            Process user annotations and refine iteratively
/xonovex-workflow:plan-worktree-merge    Merge feature worktree back to source
/xonovex-workflow:code-simplify          Consolidate duplicates, remove dead code, flatten abstractions
/xonovex-workflow:code-harden            Improve type safety, validation, and error handling
/xonovex-workflow:code-align             Align two similar implementations and suggest improvements
/xonovex-workflow:git-commit             Commit and push changes
```

Utility commands for project instructions, insights, and skill management:

```
/xonovex-utility:instructions-init          Create an AGENTS.md by analyzing directory structure
/xonovex-utility:instructions-simplify      Reduce verbosity in instruction files
/xonovex-utility:instructions-sync          Sync AGENTS.md with current directory state
/xonovex-utility:instructions-consolidate   Remove redundant files and standardize format
/xonovex-utility:insights-extract           Analyze session for mistakes and lessons learned
/xonovex-utility:insights-instructions-integrate  Convert insights into AGENTS.md bullet points
/xonovex-utility:insights-skills-integrate  Convert insights into a skill
/xonovex-utility:skill-guidelines-create    Create a skill from a document or URL
/xonovex-utility:skill-guidelines-extract   Extract patterns from codebase into a skill
/xonovex-utility:slashcommand-create        Create a slash command from a completed task
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
