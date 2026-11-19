# Workflow Plugin

Use `xonovex-workflow` to take a task from research to validated delivery and closeout. Each step has its own skill, and the harness loads the skill that matches the request. Choose `implementing-plan` for one implementation target or `delegating-plan` for supervised roadmap execution. Use the [Core Plugin](../plugin-core/README.md) for Git operations and the [Agentic Plugin](../plugin-agentic/README.md) to maintain instructions and skills.

![Research and decide, plan and review, implement and validate, review quality and deliver, then close out](diagrams/workflow-lifecycle.svg)

## Follow the workflow

Work through the phases below and return to implementation when checks or review find a gap. Continuous learning can run during any phase. A diagram label names the skill that owns the step, and a `Core:` prefix names a skill of the Core Plugin.

| Phase | Action and result |
| --- | --- |
| Research and decisions | Use `researching-changes` to investigate the task and `settling-questions` to resolve questions. Continue only when the work has a clear purpose and direction. |
| Planning and review | Use `writing-plan`, independently critique the plan with `critiquing-plan`, and apply feedback with `revising-plan`. Record the decision with `approving-plan`. Use `defining-subplans` when the work needs smaller targets. |
| Implementation and validation | Use `implementing-plan` for one target or `delegating-plan` for a supervised roadmap. Load the applicable skills, implement the target, and check success criteria with `validating-plan`. Fix failures and refresh evidence with the progress update of `implementing-plan` before selecting more work. |
| Quality and delivery | Use the Core Plugin's `auditing-code-quality` for an audit. Fix bounded findings or return to research for broader changes. Integrate isolated work when applicable, validate the combined result, and commit within the granted authority. |
| Closeout and follow-up | Use `closing-plan` to return status, evidence, remaining work, and follow-up seeds for completed, paused, or handed-over work. Use `distilling-plan` when a completed implementation should become a replayable skill suite. |

A plan approval records a review decision. Status metadata does not grant authority to implement, create branches, merge, push, or publish. Use the authority already granted by the user and the project rules.

### Open the complete diagram

Read the main phases from top to bottom. The diagram includes acceptance and rejection, revision loops, execution choices, failed checks, worktree delivery or abandonment, pull-request review, closeout, and continuous learning.

[Open the scalable diagram](diagrams/workflow-diagram.svg), [open the PNG](diagrams/workflow-diagram.png), or [edit its Graphviz source](diagrams/workflow-diagram.dot).

![Complete workflow with review outcomes, retries, delivery choices, and learning alongside implementation](diagrams/workflow-diagram.svg)

## Choose an execution mode

Use `implementing-plan` for one selected plan or subplan. Request the apply effect for implementation; the skill defaults to inspect. Reconstruct the target context and load its applicable skills before editing. Stop after the selected target rather than continuing into another child plan.

Use `delegating-plan` for supervised roadmap execution. The supervisor briefs implementation agents, independently verifies their results, and records progress. Sequential ordering is the default. Parallel groups need disjoint file sets and completed dependencies.

![Choose one target or a supervised roadmap, validate criteria, fix failures, and record progress before advancing](diagrams/development-cycle.svg)

Validate explicit success criteria, not just test exit codes. The progress update of `implementing-plan` returns refreshed progress and evidence. Persist the result separately when requested so another session can reconstruct the remaining work.

## Deliver isolated work

Use the worktree operations of the Core Plugin's [`using-git`](../plugin-core/skills/using-git/SKILL.md) skill when isolation is needed and branch creation is within the granted authority.

| Operation | Result |
| --- | --- |
| Create a worktree | Create an isolated worktree before implementation. |
| Merge a worktree | Integrate into the source worktree. Validate the combined result before declaring completion. |
| Abandon a worktree | Record why work stops and retain useful lessons. Worktree removal is a separate option. |
| Clean up worktrees | Remove eligible stale or merged worktrees when cleanup is requested. |
| Commit | Commit validated work. Push only when publication is requested. |

## Carry lessons into future work

Use `reflecting-on-sessions` to carry session lessons forward: it applies project-specific lessons to the relevant `AGENTS.md` files and reusable procedures to their owning skills. Both operations extract and apply directly, and each has a preview. Extract the lessons to files first only when a separate extraction is useful, then integrate the stored reflections.

Reflection captures session lessons. `distilling-plan` reconstructs a completed implementation as a replayable skill suite. Use the Agentic Plugin's `authoring-instructions` and `authoring-skills` when the resulting guidance needs shortening.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-workflow` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-workflow@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [approving-plan](skills/approving-plan/SKILL.md) | Recording the user's approval or rejection of a reviewed plan. |
| [closing-plan](skills/closing-plan/SKILL.md) | Closing out a completed, paused, or handed-over plan with a follow-up record. |
| [critiquing-plan](skills/critiquing-plan/SKILL.md) | A written plan needs an independent adversarial review before anyone approves or implements it. |
| [defining-subplans](skills/defining-subplans/SKILL.md) | Expanding a parent plan into focused child plans with their execution order. |
| [delegating-plan](skills/delegating-plan/SKILL.md) | Supervising the execution of a roadmap, a plan, or a list of subplans by handing each item to an implementation agent and verifying its result. |
| [distilling-plan](skills/distilling-plan/SKILL.md) | Turning a completed plan's implementation (its branch, plan record, and execution logs) into a replayable suite of skills. |
| [implementing-plan](skills/implementing-plan/SKILL.md) | Implementing, resuming, or previewing work from an existing plan or subplan, one target at a time, and when refreshing the plan's progress from implementation evidence. |
| [reflecting-on-sessions](skills/reflecting-on-sessions/SKILL.md) | Reflecting on a session or distilling lessons into reusable form. |
| [researching-changes](skills/researching-changes/SKILL.md) | Researching a codebase and external sources to ground a future implementation plan, without writing the plan or changing code. |
| [revising-plan](skills/revising-plan/SKILL.md) | Applying explicit feedback, annotations, or critique findings to an existing plan as a traceable new revision. |
| [settling-questions](skills/settling-questions/SKILL.md) | Open decisions for a plan need an answer from the user, one at a time, each with a recommendation. |
| [validating-plan](skills/validating-plan/SKILL.md) | Checking a plan's success criteria and Definition of Done against implementation evidence without changing the plan. |
| [writing-plan](skills/writing-plan/SKILL.md) | Writing one high-level implementation plan from requirements, research, or settled decisions, before any detailed subplan exists. |

## Validation

Run the plugin gate from the repository root. It checks the plugin documents, formatting, owned skills, and diagram rendering.

```bash
npx moon run plugin-workflow:ci-check --force
```

Edit diagram sources under [`diagrams/`](diagrams/README.md) using the Xonovex design system in the `drodan-platform` repository (`packages/xonovex/xonovex-design-system`), then run `nix develop --no-update-lock-file --command npx moon run plugin-workflow:graph-build --force` to refresh SVG and PNG outputs with the supplied fonts.
