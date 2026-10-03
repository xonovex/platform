# Workflow Plugin

Use `xonovex-workflow` to take a task from research to validated delivery and closeout. Choose `plan-continue` for one implementation target or `plan-delegate` for supervised roadmap execution. Use the [Core Plugin](../plugin-core/README.md) for Git operations, worktrees, and code-quality review.

![Research and decide, plan and review, implement and validate, review quality and deliver, then close out](diagrams/workflow-lifecycle.svg)

## Follow the workflow

Work through the phases below and return to implementation when checks or review find a gap. Continuous learning can run during any phase. Diagram labels omit the plugin namespace; command titles use `xonovex-workflow`, `xonovex-core`, or `xonovex-agentic` as shown in their owning catalog.

| Phase | Action and result |
| --- | --- |
| Research and decisions | Use `plan-research` to investigate the task and `plan-decide` to resolve questions. Continue only when the work has a clear purpose and direction. |
| Planning and review | Use `plan-create`, independently critique the plan with `plan-critique`, and apply feedback with `plan-revise`. Record the decision with `plan-accept` or `plan-reject`. Use `plan-subplans-create` when the work needs smaller targets. |
| Implementation and validation | Use `plan-continue` for one target or `plan-delegate` for a supervised roadmap. Load the applicable skills, implement the target, and check success criteria with `plan-validate`. Fix failures and refresh evidence with `plan-update`. |
| Quality and delivery | Use the Core Plugin's `code-quality-guide` for an audit. Fix bounded findings or return to research for broader changes. Integrate isolated work when applicable, validate the combined result, and commit within the granted authority. For requested publication, use the Integrations Plugin for pull or merge requests and the Core Plugin's `code-review-guide` for review findings. |
| Closeout and follow-up | Use `plan-followup` to return status, evidence, remaining work, and follow-up seeds for completed, paused, or handed-over work. Use `plan-distill` when a completed implementation should become a replayable skill suite. |

A plan approval records a review decision. Status metadata does not grant authority to implement, create branches, merge, push, or publish. Use the authority already granted by the user and the project rules.

### Open the complete diagram

Read the main phases from top to bottom. The diagram includes acceptance and rejection, revision loops, execution choices, failed checks, worktree delivery or abandonment, pull-request review, closeout, and continuous learning.

[Open the scalable diagram](diagrams/workflow-diagram.svg), [open the PNG](diagrams/workflow-diagram.png), or [edit its Graphviz source](diagrams/workflow-diagram.dot).

![Complete workflow with review outcomes, retries, delivery choices, and learning alongside implementation](diagrams/workflow-diagram.svg)

## Choose an execution mode

Use `plan-continue` for one selected plan or subplan. Request the `apply` effect for implementation; the planning procedure defaults to `inspect`. Reconstruct the target context and load its applicable skills before editing. Stop after the selected target rather than advancing silently to another subplan.

Use `plan-delegate` for supervised roadmap execution. The supervisor briefs implementation agents, independently verifies their results, and records progress. Sequential ordering is the default. Parallel groups need disjoint file sets and completed dependencies; use isolated worktrees where needed and allowed.

![Choose one target or a supervised roadmap, validate criteria, fix failures, and record progress before advancing](diagrams/development-cycle.svg)

Validate explicit success criteria, not just test exit codes. `plan-update` returns refreshed progress and evidence. Persist the result separately when requested so another session can reconstruct the remaining work.

## Deliver isolated work

Use the Core Plugin's worktree commands when isolation is needed and branch creation is within the granted authority.

| Command | Result |
| --- | --- |
| [`plan-worktree-create`](../plugin-core/commands/plan-worktree-create.md) | Create an isolated worktree before implementation. |
| [`plan-worktree-merge`](../plugin-core/commands/plan-worktree-merge.md) | Integrate into the source worktree. Validate the combined result before declaring completion. |
| [`plan-worktree-abandon`](../plugin-core/commands/plan-worktree-abandon.md) | Record why work stops and retain useful lessons. Worktree removal is a separate option. |
| [`plan-worktree-cleanup`](../plugin-core/commands/plan-worktree-cleanup.md) | Remove eligible stale or merged worktrees when cleanup is requested. |
| [`git-commit`](../plugin-core/commands/git-commit.md) | Commit validated work. Add `--push` only when publication is requested. |

## Carry lessons into future work

Use `reflect-to-instructions` for project-specific lessons and `reflect-to-skill` for reusable procedures. Both commands extract and apply directly. Use `--dry-run` to preview changes. Run `reflect-extract` first only when separate extraction is useful, then integrate with `--from-reflections`.

Reflection captures session lessons. `plan-distill` reconstructs a completed implementation as a replayable skill suite. Use the Agentic Plugin's instruction, skill, and slash-command simplification commands when the resulting guidance needs shortening.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-workflow` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-workflow@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-workflow@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [plan-guide](skills/plan-guide/SKILL.md) | Researching, creating, critiquing, revising, expanding, continuing, updating, validating, closing out, or distilling an implementation plan. |
| [reflect-guide](skills/reflect-guide/SKILL.md) | Reflecting on a session or distilling lessons into reusable form. |

## Commands

Use the command name exposed by the harness. Command titles use the `xonovex-workflow` namespace.

| Command | Purpose |
| --- | --- |
| `plan-accept` | [Record plan approval after a final review; implementation remains a separate operation.](commands/plan-accept.md) |
| `plan-continue` | [Resume one plan or subplan with its context and implementation skills, then stop after that target.](commands/plan-continue.md) |
| `plan-create` | [Create one high-level plan from research for review before expanding it into subplans.](commands/plan-create.md) |
| `plan-critique` | [Stress-test a plan independently and report findings for a separate revision.](commands/plan-critique.md) |
| `plan-decide` | [Resolve open decisions one at a time using known questions or targeted discovery.](commands/plan-decide.md) |
| `plan-delegate` | [Supervise roadmap implementation, verify each agent result independently, and record progress.](commands/plan-delegate.md) |
| `plan-distill` | [Turn a completed implementation into a replayable skill suite with traceable sources.](commands/plan-distill.md) |
| `plan-followup` | [Return a closeout record with status, evidence, remaining work, and follow-up seeds.](commands/plan-followup.md) |
| `plan-reject` | [Record plan rejection and its reason while preserving the plan for revision or review.](commands/plan-reject.md) |
| `plan-research` | [Research code and external evidence for requirements without creating a plan.](commands/plan-research.md) |
| `plan-revise` | [Revise a plan from explicit annotations and feedback; approval remains a separate operation.](commands/plan-revise.md) |
| `plan-subplans-create` | [Expand a parent plan into focused subplans with dependencies and parallel groups.](commands/plan-subplans-create.md) |
| `plan-update` | [Refresh the plan progress and validation evidence from the current implementation.](commands/plan-update.md) |
| `plan-validate` | [Check the plan success criteria and report evidence without changing the plan.](commands/plan-validate.md) |
| `reflect-extract` | [Extract reusable lessons from the session mistakes, discoveries, and corrections.](commands/reflect-extract.md) |
| `reflect-to-instructions` | [Apply session lessons to the relevant AGENTS.md files, or preview them with `--dry-run`.](commands/reflect-to-instructions.md) |
| `reflect-to-skill` | [Apply reusable session lessons to their owning skills, or preview them with `--dry-run`.](commands/reflect-to-skill.md) |

## Validation

Run the plugin gate from the repository root. It checks commands, formatting, owned skills, and diagram rendering.

```bash
npx moon run plugin-workflow:ci-check --force
```

Edit diagram sources under [`diagrams/`](diagrams/README.md) using the [Xonovex design](../../../../DESIGN.md), then run `nix develop --no-update-lock-file --command npx moon run plugin-workflow:graph-build --force` to refresh SVG and PNG outputs with the supplied fonts.
