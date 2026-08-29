# Workflow Commands

Install this workflow to research, plan, implement, validate, and merge work through explicit plan and worktree commands.

## Installation

Install the workflow plugin for the active harness. Codex also needs the delegated skill plugins because it does not install plugin dependencies automatically.

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

Each command delegates its procedure to a guideline skill declared in `plugin.json` dependencies. Claude Code installs these skills with the plugin and disables a command with `dependency-unsatisfied` when a required skill is missing. Codex does not install `dependencies` automatically, so install the delegated skill plugins with this plugin.

```
+---------------------+     +---------------------+     +---------------------+
|      Research       |     |      Planning       |     |   Worktree Setup    |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-research    |---->| 1. plan-create      |---->| 1. plan-worktree-   |
|    - viability      |     | 2. plan-subplans    |     |      create         |
|    - alternatives   |     | 3. git-commit       |     | 2. cd <worktree>    |
| 2. plan-decide      |     |                     |     |                     |
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

Supervised: plan-delegate walks a roadmap's ordering, briefing one implementation agent per item and reviewing before it accepts
Parallel: Multiple agents work on parallel subplan groups in separate worktrees
Learning: reflect-to-instructions / reflect-to-skill fold learnings into guidelines for future sessions
```

## Commands

Use the command that matches the current stage of the plan-driven workflow.

| Command                 | Description                                                                                |
| ----------------------- | ------------------------------------------------------------------------------------------ |
| `plan-research`         | Research codebase + web, or run a read-only code-quality audit (harden / simplify / align) |
| `plan-decide`           | Settle decisions one at a time: walk known open ones, or discover them by questioning      |
| `plan-create`           | Create a high-level plan for user review (test-first plans route to tdd / bdd guides)      |
| `plan-revise`           | Apply annotations and prompt feedback to the plan                                          |
| `plan-critique`         | Adversarially stress-test a plan to expose weaknesses                                      |
| `plan-accept`           | Approve a plan for execution (sanity-check, set status: approved)                          |
| `plan-reject`           | Reject a plan with a reason (set status: rejected, keep the plan)                          |
| `plan-subplans-create`  | Generate detailed subplans with parallel execution detection                               |
| `plan-worktree-create`  | Create a git worktree for a feature branch                                                 |
| `plan-continue`         | Resume work from an existing plan                                                          |
| `plan-delegate`         | Work a roadmap as supervisor: brief an agent per item, verify it, record and commit        |
| `plan-validate`         | Verify that a plan or current work has been fully achieved                                 |
| `plan-update`           | Update plan status and test results                                                        |
| `plan-followup`         | Close out a completed, paused, or handed-over plan with an inline follow-up record         |
| `plan-distill`          | Distill a completed plan's branch, record, and logs into a replayable skill suite          |
| `git-commit`            | Commit and push changes                                                                    |
| `plan-worktree-merge`   | Merge feature worktree back to source                                                      |
| `plan-worktree-abandon` | Document and abandon a feature with reason and learnings                                   |
| `plan-worktree-cleanup` | Remove stale and merged worktrees, and prune leftover admin metadata                       |

## Design Decisions

The workflow uses these constraints to keep plans portable and operations explicit.

- **Domain-agnostic commands**: the agent figures out what to do based on context
- **No hooks except git hooks**: agents decide when something cannot be fixed
- **Plans committed in git**: continue from another machine, branch off for alternatives
- **`*-simplify` commands**: generalize, compress, remove duplication
