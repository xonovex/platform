# Workflow Commands

Plan-driven development workflow with worktrees and parallel execution.

## Installation

### Claude Code

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-workflow@xonovex-marketplace
```

### Dependencies

Each command delegates its procedure to a guideline skill, declared in `plugin.json`
`dependencies`. On Claude Code, installing this plugin auto-installs those skills; if a
depended-on skill is missing the command is disabled with `dependency-unsatisfied`. On
Codex, `dependencies` is not auto-installed — install the delegated skill plugins
alongside this one.

The `pr-*` commands additionally load a **host-delivery skill** chosen from the git remote
to open PRs/MRs and post reviews — `xonovex-skill-github` (GitHub), `xonovex-skill-gitlab`
(GitLab), or another `xonovex-skill-<host>`. These are pluggable, not hard dependencies:
install the one matching your host. With none installed, the `pr-*` commands still load and
tell you which host skill to add.

```
+---------------------+     +---------------------+     +---------------------+
|      Research       |     |      Planning       |     |   Worktree Setup    |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-research    |---->| 1. plan-create      |---->| 1. plan-worktree-   |
|    - viability      |     | 2. plan-subplans    |     |      create         |
|    - alternatives   |     | 3. git-commit       |     | 2. cd <worktree>    |
| 2. plan-clarify     |     |                     |     |                     |
+---------------------+     +---------------------+     +---------------------+
                                                                  |
            +-----------------------------------------------------+
            |
            v
+---------------------+     +---------------------+     +---------------------+
|  Development Loop   |     |    Code Quality     |     |        Merge        |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-continue    |---->| 1. plan-research    |---->| 1. plan-worktree-   |
| 2. (implement)      |     |    (code-quality    |     |      merge          |
| 3. plan-validate    |     |     audit)          |     | 2. git-commit       |
| 4. reflect-extract  |     +---------------------+     |      --push         |
| 5. plan-update      |            |                    +---------------------+
+---------------------+            |                              |
            ^                      |                              |
            |                      |                              v
            +--- more subplans? ---+                    +---------------------+
                                                        |        Done         |
                                                        +---------------------+

Parallel: Multiple agents work on parallel subplan groups in separate worktrees
Learning: reflect-to-instructions / reflect-to-skill fold learnings into guidelines for future sessions
```

[View workflow diagram](../../diagram/diagram-agent-workflow/workflow-diagram.png)

## Commands

| Command                 | Description                                                                                |
| ----------------------- | ------------------------------------------------------------------------------------------ |
| `plan-research`         | Research codebase + web, or run a read-only code-quality audit (harden / simplify / align) |
| `plan-interrogate`      | Interrogate to surface unknown decisions before a plan exists                              |
| `plan-clarify`          | Walk known open decisions one by one in plain prose                                        |
| `plan-create`           | Create a high-level plan for user review (test-first plans route to tdd / bdd guides)      |
| `plan-refine`           | Process user annotations and refine iteratively                                            |
| `plan-critique`         | Adversarially stress-test a plan to expose weaknesses                                      |
| `plan-subplans-create`  | Generate detailed subplans with parallel execution detection                               |
| `plan-worktree-create`  | Create a git worktree for a feature branch                                                 |
| `plan-continue`         | Resume work from an existing plan                                                          |
| `plan-validate`         | Verify that a plan or current work has been fully achieved                                 |
| `plan-update`           | Update plan status and test results                                                        |
| `git-commit`            | Commit and push changes                                                                    |
| `pr-create`             | Open a pull request with a drafted description (any host)                                  |
| `pr-review-analyze`     | Review a branch diff into a structured findings file                                       |
| `pr-review-refine`      | Refine review findings one by one before publishing                                        |
| `pr-review-post`        | Publish a structured, labelled code review to a PR (any host)                              |
| `pr-review-resolve`     | Verify fixes and resolve the review's blocking threads                                     |
| `plan-worktree-merge`   | Merge feature worktree back to source                                                      |
| `plan-worktree-abandon` | Document and abandon a feature with reason and learnings                                   |

## Design Decisions

- **Domain-agnostic commands**: the agent figures out what to do based on context
- **No hooks except git hooks**: agents decide when something cannot be fixed
- **Plans committed in git**: continue from another machine, branch off for alternatives
- **`*-simplify` commands**: generalize, compress, remove duplication
