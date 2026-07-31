---
type: report
parent_plan: plans/skill-catalog-eval-sweeps.md
status: final
---

# Sweep results, 2026-07-27

Per-skill outcomes for the three gates, distilled from the sweep runs that back
this plan's success criteria.

The raw per-run artifacts (`.skill-eval-results/`) are deliberately gitignored:
they are bulky, regenerable, and rewritten by every run. That leaves the aggregate
counts in the execution records with nothing reviewable underneath them, so the
per-skill rows are recorded here. Reproduce with
`npx moon run '#skill:ci-skill-eval-trigger'`,
`script-moon-skill-eval-routing:ci-routing`, and
`npx moon run '#skill:ci-skill-eval-output'` — roughly two hours and, at the
per-call caps, under $76 for triggers and $255 for outputs.

## Trigger gate

72 skills, 317 queries scored, 199 deferred to the routing gate,
zero invalid runs, zero transient retries. 10 failures.

| skill                         | scored | passed        | deferred |
| ----------------------------- | ------ | ------------- | -------- |
| `skill-accessibility`         | 7      | 7             | 0        |
| `skill-asset-pipeline`        | 4      | 4             | 4        |
| `skill-astro`                 | 5      | 5             | 1        |
| `skill-audio`                 | 5      | 5             | 3        |
| `skill-bdd`                   | 5      | 5             | 3        |
| `skill-c99-game-opinionated`  | 3      | 3             | 3        |
| `skill-c99-opinionated`       | 9      | 8 (1 failing) | 0        |
| `skill-c99`                   | 3      | 3             | 3        |
| `skill-claude-code`           | 3      | 2 (1 failing) | 4        |
| `skill-cmake`                 | 3      | 3             | 3        |
| `skill-code-quality`          | 5      | 5             | 4        |
| `skill-code-review`           | 4      | 4             | 2        |
| `skill-codex`                 | 3      | 3             | 4        |
| `skill-command`               | 3      | 3             | 3        |
| `skill-connascence`           | 3      | 3             | 3        |
| `skill-content`               | 3      | 3             | 3        |
| `skill-copilot`               | 3      | 3             | 3        |
| `skill-credential-management` | 7      | 7             | 0        |
| `skill-cross-platform`        | 5      | 5             | 3        |
| `skill-data-model`            | 4      | 4             | 4        |
| `skill-data-oriented-design`  | 6      | 6             | 2        |
| `skill-ddd`                   | 4      | 4             | 2        |
| `skill-debugging`             | 4      | 4             | 3        |
| `skill-docker`                | 4      | 4             | 2        |
| `skill-ecs`                   | 3      | 3             | 4        |
| `skill-editor-viewport`       | 5      | 5             | 3        |
| `skill-fp`                    | 3      | 3             | 3        |
| `skill-game-networking`       | 5      | 5             | 4        |
| `skill-git`                   | 8      | 8             | 3        |
| `skill-github`                | 6      | 6             | 5        |
| `skill-gitlab`                | 6      | 6             | 3        |
| `skill-gpu-rendering-vulkan`  | 4      | 4             | 3        |
| `skill-gpu-rendering`         | 6      | 6             | 4        |
| `skill-hexagonal-pattern`     | 4      | 4             | 3        |
| `skill-hono-opinionated`      | 5      | 4 (1 failing) | 2        |
| `skill-hono`                  | 4      | 4             | 2        |
| `skill-imgui`                 | 4      | 4             | 3        |
| `skill-instruction`           | 7      | 6 (1 failing) | 0        |
| `skill-kiro`                  | 3      | 3             | 3        |
| `skill-kubernetes`            | 4      | 4             | 2        |
| `skill-llmstxt`               | 3      | 3             | 3        |
| `skill-lock-free`             | 4      | 4             | 3        |
| `skill-lua-opinionated`       | 3      | 2 (1 failing) | 3        |
| `skill-lua`                   | 3      | 3             | 3        |
| `skill-memory-management`     | 3      | 3             | 5        |
| `skill-microkernel-pattern`   | 4      | 4             | 3        |
| `skill-moon`                  | 3      | 3             | 3        |
| `skill-node-graph`            | 5      | 5             | 4        |
| `skill-npm`                   | 3      | 3             | 3        |
| `skill-oop`                   | 3      | 3             | 3        |
| `skill-opencode`              | 3      | 3             | 3        |
| `skill-orthogonal-pattern`    | 8      | 8             | 2        |
| `skill-pi`                    | 3      | 3             | 3        |
| `skill-plan`                  | 9      | 9             | 1        |
| `skill-pull-request`          | 4      | 4             | 2        |
| `skill-python`                | 4      | 4             | 2        |
| `skill-react`                 | 7      | 7             | 2        |
| `skill-reflect`               | 3      | 3             | 3        |
| `skill-shell-scripting`       | 3      | 3             | 3        |
| `skill-skill`                 | 5      | 4 (1 failing) | 2        |
| `skill-sql-postgresql`        | 3      | 3             | 3        |
| `skill-tdd`                   | 4      | 4             | 5        |
| `skill-terraform`             | 5      | 5             | 2        |
| `skill-testing`               | 3      | 2 (1 failing) | 4        |
| `skill-threejs`               | 3      | 3             | 3        |
| `skill-typescript-to-lua`     | 3      | 3             | 3        |
| `skill-typescript`            | 5      | 5             | 2        |
| `skill-user-stories`          | 4      | 4             | 3        |
| `skill-versioning`            | 3      | 3             | 3        |
| `skill-vitest`                | 4      | 4             | 3        |
| `skill-workflow`              | 12     | 9 (3 failing) | 0        |
| `skill-zod`                   | 3      | 3             | 3        |

### Trigger failures

| skill                    | rate  | expected | query                                                                                      |
| ------------------------ | ----- | -------- | ------------------------------------------------------------------------------------------ |
| `skill-c99-opinionated`  | 0.333 | fire     | editing one header in our C engine triggers a near-full rebuild and compiles are crawling  |
| `skill-claude-code`      | 0     | fire     | Configure Claude Code permissions so repository reads and the test command are allowed, de |
| `skill-hono-opinionated` | 0.333 | fire     | Select the fastest valid Hono router for static API routes plus one wildcard asset route,  |
| `skill-instruction`      | 0     | fire     | no claude.md in services/payments — analyze the code and write one                         |
| `skill-lua-opinionated`  | 0.333 | fire     | Review this Lua Opinionated change for essentials and fix the concrete issues you find.    |
| `skill-skill`            | 0     | fire     | before we install this third-party skill, audit its scripts and the URLs it fetches and lo |
| `skill-testing`          | 0.333 | fire     | this test is flaky and obscure — what smell is it and how do I fix it?                     |
| `skill-workflow`         | 0.333 | fire     | Independently validate this implementation handoff without inheriting the implementer's co |
| `skill-workflow`         | 0.333 | fire     | Carry this decision and its rationale forward as context: 'we chose per-route limits over  |
| `skill-workflow`         | 0     | fire     | Independently review revision 91c2 of packages/api in two passes so prior decision context |

## Routing gate

First run of this gate. 80 scenarios, 77 pass, 3 fail, zero invalid runs.

| expected owner          | rate  | competitors                        | query                                                                            |
| ----------------------- | ----- | ---------------------------------- | -------------------------------------------------------------------------------- |
| `debugging-guide`       | 0     | memory-management-guide            | I want to tag every allocation with file and line and check the counter is zero  |
| `github-guide`          | 0.333 | git-guide                          | reply in the review thread on PR #482 and mark it resolved so the PR can merge   |
| `lua-opinionated-guide` | 0.333 | lua-guide, typescript-to-lua-guide | Review this Lua Opinionated change for essentials and fix the concrete issues yo |

## Output gate

72 skills, zero invalid runs. 45 PASS, 27 FAIL.
Golden eval (`pull-request-guide` eval 5) 0.916 against its 0.833 anchor.

| skill                         | tier         | gate | with | without | delta |
| ----------------------------- | ------------ | ---- | ---- | ------- | ----- |
| `accessibility-guide`         | moderate     | PASS | 0.83 | 0.17    | +0.67 |
| `asset-pipeline-guide`        | conservative | FAIL | 0.83 | 0.35    | +0.48 |
| `astro-guide`                 | moderate     | PASS | 1.00 | 0.92    | +0.08 |
| `audio-guide`                 | conservative | PASS | 0.92 | 0.67    | +0.25 |
| `bdd-guide`                   | aggressive   | PASS | 1.00 | 0.54    | +0.46 |
| `c99-game-opinionated-guide`  | moderate     | PASS | 0.88 | 0.71    | +0.17 |
| `c99-guide`                   | aggressive   | FAIL | 0.75 | 0.79    | -0.04 |
| `c99-opinionated-guide`       | moderate     | PASS | 1.00 | 0.62    | +0.38 |
| `claude-code-guide`           | moderate     | FAIL | 0.33 | 0.00    | +0.33 |
| `cmake-guide`                 | moderate     | PASS | 0.92 | 0.85    | +0.06 |
| `code-quality-guide`          | aggressive   | FAIL | 1.00 | 1.00    | +0.00 |
| `code-review-guide`           | aggressive   | PASS | 0.88 | 0.50    | +0.38 |
| `codex-guide`                 | moderate     | FAIL | 0.42 | 0.00    | +0.42 |
| `command-guide`               | conservative | FAIL | 0.33 | 0.00    | +0.33 |
| `connascence-guide`           | aggressive   | PASS | 1.00 | 0.71    | +0.29 |
| `content-guide`               | conservative | FAIL | 0.83 | 0.42    | +0.42 |
| `copilot-guide`               | moderate     | FAIL | 0.25 | 0.00    | +0.25 |
| `credential-management-guide` | aggressive   | PASS | 0.80 | 0.50    | +0.30 |
| `cross-platform-guide`        | conservative | PASS | 1.00 | 0.33    | +0.67 |
| `data-model-guide`            | conservative | PASS | 0.96 | 0.71    | +0.25 |
| `data-oriented-design-guide`  | conservative | PASS | 1.00 | 0.65    | +0.35 |
| `ddd-guide`                   | aggressive   | PASS | 1.00 | 0.75    | +0.25 |
| `debugging-guide`             | aggressive   | PASS | 1.00 | 0.50    | +0.50 |
| `docker-guide`                | moderate     | FAIL | 0.94 | 1.00    | -0.06 |
| `ecs-guide`                   | conservative | FAIL | 0.75 | 0.25    | +0.50 |
| `editor-viewport-guide`       | conservative | PASS | 0.94 | 0.21    | +0.73 |
| `fp-guide`                    | aggressive   | PASS | 1.00 | 0.75    | +0.25 |
| `game-networking-guide`       | conservative | PASS | 1.00 | 0.75    | +0.25 |
| `git-guide`                   | moderate     | FAIL | 0.63 | 0.39    | +0.24 |
| `github-guide`                | moderate     | FAIL | 0.54 | 0.04    | +0.50 |
| `gitlab-guide`                | moderate     | FAIL | 0.71 | 0.16    | +0.55 |
| `gpu-rendering-guide`         | conservative | PASS | 0.94 | 0.71    | +0.23 |
| `gpu-rendering-vulkan-guide`  | conservative | FAIL | 0.83 | 0.46    | +0.38 |
| `hexagonal-pattern-guide`     | aggressive   | PASS | 1.00 | 0.62    | +0.38 |
| `hono-guide`                  | moderate     | PASS | 0.92 | 0.38    | +0.54 |
| `hono-opinionated-guide`      | moderate     | FAIL | 0.75 | 0.38    | +0.38 |
| `imgui-guide`                 | conservative | PASS | 1.00 | 0.19    | +0.81 |
| `instruction-guide`           | conservative | FAIL | 0.44 | 0.25    | +0.19 |
| `kiro-guide`                  | moderate     | FAIL | 0.33 | 0.00    | +0.33 |
| `kubernetes-guide`            | moderate     | PASS | 1.00 | 0.65    | +0.35 |
| `llmstxt-guide`               | conservative | PASS | 1.00 | 0.12    | +0.88 |
| `lock-free-guide`             | conservative | PASS | 0.96 | 0.67    | +0.29 |
| `lua-guide`                   | aggressive   | PASS | 1.00 | 0.75    | +0.25 |
| `lua-opinionated-guide`       | moderate     | PASS | 1.00 | 0.94    | +0.06 |
| `memory-management-guide`     | conservative | FAIL | 0.83 | 0.50    | +0.33 |
| `microkernel-pattern-guide`   | aggressive   | PASS | 0.92 | 0.46    | +0.46 |
| `moon-guide`                  | moderate     | FAIL | 0.75 | 0.19    | +0.56 |
| `node-graph-guide`            | conservative | PASS | 1.00 | 0.67    | +0.33 |
| `npm-guide`                   | moderate     | PASS | 1.00 | 0.71    | +0.29 |
| `oop-guide`                   | aggressive   | FAIL | 0.75 | 0.75    | +0.00 |
| `opencode-guide`              | moderate     | FAIL | 0.25 | 0.00    | +0.25 |
| `orthogonal-pattern-guide`    | aggressive   | PASS | 0.92 | 0.54    | +0.38 |
| `pi-guide`                    | moderate     | FAIL | 0.50 | 0.00    | +0.50 |
| `plan-guide`                  | moderate     | FAIL | 0.48 | 0.50    | -0.02 |
| `pull-request-guide`          | aggressive   | PASS | 0.78 | 0.30    | +0.48 |
| `python-guide`                | aggressive   | PASS | 1.00 | 0.58    | +0.42 |
| `react-guide`                 | moderate     | PASS | 0.88 | 0.25    | +0.62 |
| `reflect-guide`               | aggressive   | PASS | 0.80 | 0.15    | +0.65 |
| `shell-scripting-guide`       | aggressive   | FAIL | 0.92 | 1.00    | -0.08 |
| `skill-guide`                 | conservative | FAIL | 0.77 | 0.20    | +0.57 |
| `sql-postgresql-guide`        | aggressive   | PASS | 1.00 | 0.93    | +0.07 |
| `tdd-guide`                   | aggressive   | PASS | 0.85 | 0.54    | +0.31 |
| `terraform-guide`             | moderate     | PASS | 0.88 | 0.69    | +0.19 |
| `testing-guide`               | aggressive   | PASS | 1.00 | 0.79    | +0.21 |
| `threejs-guide`               | moderate     | PASS | 0.90 | 0.46    | +0.44 |
| `typescript-guide`            | aggressive   | PASS | 0.88 | 0.38    | +0.50 |
| `typescript-to-lua-guide`     | moderate     | PASS | 1.00 | 0.44    | +0.56 |
| `user-stories-guide`          | aggressive   | PASS | 1.00 | 0.71    | +0.29 |
| `versioning-guide`            | conservative | PASS | 1.00 | 0.69    | +0.31 |
| `vitest-guide`                | moderate     | FAIL | 0.75 | 0.88    | -0.12 |
| `workflow-guide`              | moderate     | FAIL | 0.51 | 0.11    | +0.40 |
| `zod-guide`                   | moderate     | PASS | 0.96 | 0.69    | +0.27 |
