# Workflow Plugin

Research, plan, implement, validate, and close out work, then integrate lessons from the session.

Install `xonovex-workflow@xonovex-marketplace` in Claude Code or Codex.

## Skills

- [plan-guide](skills/plan-guide/SKILL.md)
- [reflect-guide](skills/reflect-guide/SKILL.md)

## Commands

| Command                   | Purpose                                                                         |
| ------------------------- | ------------------------------------------------------------------------------- |
| `plan-accept`             | [Approve a Plan](commands/plan-accept.md)                                       |
| `plan-continue`           | [Continue Progress from Plan](commands/plan-continue.md)                        |
| `plan-create`             | [Create Plan with Research](commands/plan-create.md)                            |
| `plan-critique`           | [Adversarially Critique a Plan](commands/plan-critique.md)                      |
| `plan-decide`             | [Settle Decisions One at a Time](commands/plan-decide.md)                       |
| `plan-delegate`           | [Supervise Roadmap Execution by Delegation](commands/plan-delegate.md)          |
| `plan-distill`            | [Distill Completed Work Into Skills](commands/plan-distill.md)                  |
| `plan-followup`           | [Close Out a Plan](commands/plan-followup.md)                                   |
| `plan-reject`             | [Reject a Plan](commands/plan-reject.md)                                        |
| `plan-research`           | [Research Codebase and Web](commands/plan-research.md)                          |
| `plan-revise`             | [Revise Plan from Feedback](commands/plan-revise.md)                            |
| `plan-subplans-create`    | [Generate Detailed Subplans from Parent Plan](commands/plan-subplans-create.md) |
| `plan-update`             | [Update Plan Progress](commands/plan-update.md)                                 |
| `plan-validate`           | [Validate Plan Achievement](commands/plan-validate.md)                          |
| `reflect-extract`         | [Extract Development Lessons](commands/reflect-extract.md)                      |
| `reflect-to-instructions` | [Convert Insights to AGENTS.md](commands/reflect-to-instructions.md)            |
| `reflect-to-skill`        | [Convert Insights to Skill](commands/reflect-to-skill.md)                       |

## Validation

Run `npx moon run plugin-workflow:ci-check` from the repository root.
