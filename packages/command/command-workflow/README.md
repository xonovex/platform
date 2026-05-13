# Workflow Commands

Plan-driven development workflow with worktrees and parallel execution.

## Installation

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

```
+---------------------+     +---------------------+     +---------------------+
|      Research       |     |      Planning       |     |   Worktree Setup    |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-research    |---->| 1. plan-create      |---->| 1. plan-worktree-   |
|    - viability      |     | 2. plan-subplans    |     |      create         |
|    - alternatives   |     | 3. git-commit       |     | 2. cd <worktree>    |
+---------------------+     +---------------------+     +---------------------+
                                                                  |
            +-----------------------------------------------------+
            |
            v
+---------------------+     +---------------------+     +---------------------+
|  Development Loop   |     |    Code Quality     |     |        Merge        |
+---------------------+     +---------------------+     +---------------------+
| 1. plan-continue    |---->| 1. code-simplify    |---->| 1. plan-worktree-   |
| 2. (implement)      |     | 2. code-harden      |     |      merge          |
| 3. plan-validate    |     |                     |     | 2. git-commit       |
| 4. insights-extract |     +---------------------+     |      --push         |
| 5. plan-update      |            |                    +---------------------+
+---------------------+            |                              |
            ^                      |                              |
            |                      |                              v
            +--- more subplans? ---+                    +---------------------+
                                                        |        Done         |
                                                        +---------------------+

Parallel: Multiple agents work on parallel subplan groups in separate worktrees
Learning: insights-integrate merges learnings into guidelines for future sessions
```

[View workflow diagram](../../diagram/diagram-agent-workflow/workflow-diagram.png)

## Commands

| Command                 | Description                                                    |
| ----------------------- | -------------------------------------------------------------- |
| `plan-research`         | Research codebase and web for requirements                     |
| `plan-create`           | Create a high-level plan for user review                       |
| `plan-tdd-create`       | Create a TDD plan with research for user review                |
| `plan-subplans-create`  | Generate detailed subplans with parallel execution detection   |
| `plan-worktree-create`  | Create a git worktree for a feature branch                     |
| `plan-continue`         | Resume work from an existing plan                              |
| `plan-validate`         | Verify that a plan or current work has been fully achieved     |
| `plan-update`           | Update plan status and test results                            |
| `plan-refine`           | Process user annotations and refine iteratively                |
| `plan-worktree-merge`   | Merge feature worktree back to source                          |
| `plan-worktree-abandon` | Document and abandon a feature with reason and learnings       |
| `code-simplify`         | Consolidate duplicates, remove dead code, flatten abstractions |
| `code-harden`           | Improve type safety, validation, and error handling            |
| `code-align`            | Align two similar implementations and suggest improvements     |
| `git-commit`            | Commit and push changes                                        |

## Design Decisions

- **Domain-agnostic commands**: the agent figures out what to do based on context
- **No hooks except git hooks**: agents decide when something cannot be fixed
- **Plans committed in git**: continue from another machine, branch off for alternatives
- **`*-simplify` commands**: generalize, compress, remove duplication
