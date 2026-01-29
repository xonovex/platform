# @xonovex/skills

Xonovex coding guidelines and workflow skills for Claude Code.

## Installation

```bash
# Install as npm package
npm install @xonovex/skills

# Or add the Xonovex marketplace (from within Claude Code)
/plugin marketplace add xonovex/platform

# Install the skills plugin
/plugin install xonovex@xonovex-platform
```

Or test locally during development:

```bash
claude --plugin-dir ./packages/plugins/skills
```

Skills are namespaced as `/xonovex:<skill-name>` (e.g., `/xonovex:typescript-guidelines`).

## Coding Guidelines

Auto-triggered guidelines that activate based on file type and context.

### Languages

| Skill                             | Description                                                                  |
| --------------------------------- | ---------------------------------------------------------------------------- |
| `typescript-guidelines`           | TypeScript Node.js ESM with strict mode, async/await, and Zod validation     |
| `python-guidelines`               | Python 3.12+ with async/await, dataclasses, type hints, and pytest           |
| `c99-guidelines`                  | Modern C99 with memory management, type safety, and error handling           |
| `c99-opinionated-guidelines`      | Opinionated C99 for systems/game/embedded with data-oriented design and SIMD |
| `c99-game-opinionated-guidelines` | Game engine patterns: caller-owns-memory, inverse mass, SoA optimization     |
| `lua-guidelines`                  | Lua 5.4+ with module patterns, tables, coroutines, and error handling        |
| `lua-opinionated-guidelines`      | LuaJIT performance patterns with JIT-friendly tables and pre-allocation      |
| `typescript-to-lua-guidelines`    | TypeScript-to-Lua transpilation with TSTL 1.24+                              |
| `shell-scripting-guidelines`      | POSIX shell scripts with strict mode, quoting, and idempotency               |
| `sql-postgresql-guidelines`       | PostgreSQL 15+ with CTEs, indexing, JSONB, and row-level security            |

### Frameworks

| Skill                         | Description                                                                |
| ----------------------------- | -------------------------------------------------------------------------- |
| `react-guidelines`            | React 19+ with Server Components, Form Actions, and React Compiler         |
| `astro-guidelines`            | Astro static sites with islands architecture and content collections       |
| `hono-guidelines`             | Hono 4.0+ API servers with validation, middleware, and WebSocket           |
| `hono-opinionated-guidelines` | Opinionated Hono with inline OpenAPI handlers and router selection         |
| `express.js-guidelines`       | Express 5+ API servers with route organization and Zod validation          |
| `motion-react-guidelines`     | React animations with Motion: gestures, scroll effects, layout transitions |
| `remotion-guidelines`         | Programmatic video with React using frame-driven animations                |
| `threejs-guidelines`          | Vanilla Three.js with scene setup, materials, shaders, and post-processing |

### Infrastructure & Build

| Skill                   | Description                                                               |
| ----------------------- | ------------------------------------------------------------------------- |
| `docker-guidelines`     | Docker multi-stage builds, security hardening, and optimization           |
| `kubernetes-guidelines` | Kubernetes manifests with GitOps, Kustomize, and multi-environment config |
| `terraform-guidelines`  | Terraform 1.12+ with module design and environment isolation              |
| `cmake-guidelines`      | CMake 3.20+ target-based builds and dependency management                 |
| `moon-guidelines`       | Moonrepo monorepo task management with inheritance and project tagging    |

### Testing & Validation

| Skill               | Description                                                            |
| ------------------- | ---------------------------------------------------------------------- |
| `vitest-guidelines` | Vitest 3+ with type safety, HTTP testing, and mock patterns            |
| `zod-guidelines`    | Zod 4.0+ runtime validation with schema composition and type inference |

### Paradigms

| Skill                    | Description                                                               |
| ------------------------ | ------------------------------------------------------------------------- |
| `general-fp-guidelines`  | Functional programming with pure functions, immutability, and composition |
| `general-oop-guidelines` | Object-oriented programming with SOLID principles and design patterns     |

### Meta & Content

| Skill                     | Description                                                                  |
| ------------------------- | ---------------------------------------------------------------------------- |
| `git-guidelines`          | Git conventional commits, merge conflict resolution, and worktree management |
| `skill-guidelines`        | Creating and managing guideline skills with progressive disclosure           |
| `plan-guidelines`         | Planning research and analysis tasks with code alignment and simplification  |
| `insights-guidelines`     | Extracting insights from development sessions and integrating into skills    |
| `instruction-guidelines`  | Working with AGENTS.md/CLAUDE.md project instruction files                   |
| `content-guidelines`      | Bilingual content creation with structured formatting and CEFR levels        |
| `presentation-guidelines` | Creating presentations from codebase analysis with diagrams and Motion       |
| `strudel-guidelines`      | Algorithmic music generation with Strudel.cc and mini-notation               |

## Workflow

Plan-driven development workflow with worktree isolation and parallel execution.

```
+---------------------+     +---------------------+     +---------------------+
|      Research       |     |      Planning       |     |   Worktree Setup    |
+---------------------+     +---------------------+     +---------------------+
| /plan-research      |---->| /plan-create        |---->| /plan-worktree-     |
|   - viability       |     | /plan-subplans-     |     |       create        |
|   - alternatives    |     |       create        |     |                     |
+---------------------+     | /plan-tdd-create    |     +---------------------+
                            +---------------------+               |
            +------------------------------------------------------+
            |
            v
+---------------------+     +---------------------+     +---------------------+
|  Development Loop   |     |    Code Quality     |     |        Merge        |
+---------------------+     +---------------------+     +---------------------+
| /plan-continue      |---->| /code-simplify      |---->| /plan-worktree-     |
| (implement)         |     | /code-harden        |     |       merge         |
| /plan-validate      |     | /code-align         |     | /git-commit --push  |
| /insights-extract   |     |                     |     +---------------------+
| /plan-update        |     +---------------------+               |
+---------------------+            |                              |
            ^                      |                              v
            |                      |                    +---------------------+
            +--- more subplans? ---+                    |        Done         |
                                                        +---------------------+

Parallel: Multiple agents work on parallel subplan groups in separate worktrees
Learning: /insights-integrate merges learnings into guidelines for future sessions
```

### Workflow Commands

| Command                  | Description                                                    |
| ------------------------ | -------------------------------------------------------------- |
| `/plan-research`         | Research codebase and web for viability and alternatives       |
| `/plan-create`           | Create a high-level plan with research for user review         |
| `/plan-tdd-create`       | Create a TDD plan with atomic test step proposals              |
| `/plan-subplans-create`  | Generate detailed subplans with parallel execution detection   |
| `/plan-continue`         | Resume work from an existing plan document                     |
| `/plan-validate`         | Verify that plan objectives have been fully achieved           |
| `/plan-update`           | Update plan status, phase, and timestamps                      |
| `/plan-worktree-create`  | Create a git worktree for isolated feature development         |
| `/plan-worktree-merge`   | Merge feature worktree back to source with conflict resolution |
| `/plan-worktree-abandon` | Document and abandon a feature with reason and learnings       |
| `/code-simplify`         | Analyze code for consolidation and dead code removal           |
| `/code-harden`           | Analyze code for type safety, validation, and error handling   |
| `/code-align`            | Compare two implementations for structural differences         |
| `/git-commit`            | Auto-generate conventional commit messages with optional push  |
| `/insights-extract`      | Extract development mistakes and lessons from the session      |
| `/insights-integrate`    | Convert extracted insights into guideline skills               |

## License

MIT
