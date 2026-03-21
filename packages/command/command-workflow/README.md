# Workflow Commands

Plan-driven development workflow with worktrees and parallel execution.

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

| Command                | Description                                           |
| ---------------------- | ----------------------------------------------------- |
| `plan-research`        | Research viability, suggest alternatives              |
| `plan-create`          | Create plan with frontmatter and parallelization info |
| `plan-subplans-create` | Create subplans for parallel execution                |
| `plan-worktree-create` | Create worktree at `../<repo>-<feature>`              |
| `plan-continue`        | Auto-detect plan and resume work                      |
| `plan-validate`        | Validate against guidelines and tests                 |
| `plan-update`          | Update plan status                                    |
| `plan-worktree-merge`  | Merge with intelligent conflict resolution            |
| `code-simplify`        | Find code smells                                      |
| `code-harden`          | Improve type safety and error handling                |
| `code-align`           | Align two similar implementations                     |
| `git-commit`           | Commit changes (use `--push` to push)                 |

## Design Decisions

- **Domain-agnostic commands**: the agent figures out what to do based on context
- **No hooks except git hooks**: agents decide when something cannot be fixed
- **Plans committed in git**: continue from another machine, branch off for alternatives
- **`*-simplify` commands**: generalize, compress, remove duplication
