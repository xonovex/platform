# Core Plugin

Apply software design, testing, review, version control, credentials, and accessibility practices.

## Install

Add the [Xonovex marketplace](../../../../README.md#agent-plugins) once, then install this plugin in the selected harness.

```bash
# Claude Code
claude plugin install xonovex-core@xonovex-marketplace

# Codex
codex plugin add xonovex-core@xonovex-marketplace
```

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [accessibility-guide](skills/accessibility-guide/SKILL.md) | Planning, assessing, governing, or operating accessibility across web, mobile, desktop, documents, content, and service journeys. |
| [bdd-guide](skills/bdd-guide/SKILL.md) | Driving software from concrete agreed examples of behaviour: running three-amigos discovery and example mapping before coding, formulating Given-When-Then scenarios in feature files, and treating them as living documentation that tests the team their shared understanding. ATDD and BDD are one practice here. |
| [code-quality-guide](skills/code-quality-guide/SKILL.md) | Auditing existing code for quality WITHOUT changing it: a read-only pass that finds smells, grades them by severity, and routes each to its owner. |
| [code-review-guide](skills/code-review-guide/SKILL.md) | Writing, structuring, or labelling code-review feedback on a pull or merge request: Conventional Comments labels (praise / nitpick / suggestion / issue / question / thought / chore / todo), the blocking vs non-blocking vs if-minor decorations, pairing one top-level summary with line-anchored inline comments, cross-linking the summary to its details instead of 'see comment 3', and verifying each claim against an authoritative source before asserting it. Platform-independent review craft. |
| [connascence-guide](skills/connascence-guide/SKILL.md) | Grading or loosening the coupling between two pieces of code, or judging how cohesive a module is. |
| [credential-management-guide](skills/credential-management-guide/SKILL.md) | Choosing, storing, injecting, rotating, revoking, or responding to exposure of machine-to-machine credentials and secrets: the tokens a service, workload, or CI job presents to another system. |
| [ddd-guide](skills/ddd-guide/SKILL.md) | Finding and naming domain boundaries and modelling inside them: establishing a ubiquitous language, drawing bounded contexts and a context map, protecting a model with an anti-corruption layer, and applying the tactical building blocks (entity, value object, aggregate + root, domain event, repository, domain/application service). |
| [debugging-guide](skills/debugging-guide/SKILL.md) | Chasing a bug in native or low-level software: a crash, access violation, use-after-free, leak, intermittent/heisenbug, or 'works on my machine' failure, and when deciding how to prevent a whole class of bugs by design. |
| [fp-guide](skills/fp-guide/SKILL.md) | Writing functional-style code or reviewing for FP cleanliness. |
| [git-guide](skills/git-guide/SKILL.md) | Running git operations or resolving repo-state issues. |
| [hexagonal-pattern-guide](skills/hexagonal-pattern-guide/SKILL.md) | Isolating an application or domain core from its I/O and delivery mechanisms behind interfaces: hexagonal / ports-and-adapters / clean / onion architecture. |
| [microkernel-pattern-guide](skills/microkernel-pattern-guide/SKILL.md) | Building an extensible system: a minimal core plus interchangeable plug-ins selected through a registry. The microkernel / plug-in architecture. |
| [oop-guide](skills/oop-guide/SKILL.md) | Designing class hierarchies or applying OOP principles. |
| [orthogonal-pattern-guide](skills/orthogonal-pattern-guide/SKILL.md) | Deciding how to decompose a system into modules or packages along independent variation axes, and where each concern's boundary belongs. |
| [pull-request-guide](skills/pull-request-guide/SKILL.md) | Authoring a pull or merge request - writing the description, sizing and splitting the change, documenting how it was tested, surfacing tradeoffs, and getting it review-ready before assigning reviewers. |
| [tdd-guide](skills/tdd-guide/SKILL.md) | Driving code test-first or coaching the red-green-refactor rhythm: writing a failing test before the production code, going green with fake-it / obvious-implementation / triangulation, then refactoring to remove duplication, working a test list one item at a time, and letting the tests grow the design. |
| [testing-guide](skills/testing-guide/SKILL.md) | Writing or reviewing a single good test independent of any framework: structuring it as Arrange-Act-Assert / Four-Phase, meeting FIRST, naming it, choosing and naming the right test double (Dummy / Stub / Spy / Mock / Fake), deciding what to mock and what not to, telling state from behaviour verification, and spotting test smells. |
| [user-stories-guide](skills/user-stories-guide/SKILL.md) | Writing, evaluating, splitting, or refining user stories: applying INVEST, the 3 Cs (Card / Conversation / Confirmation), the 'As a / I want / so that' template, writing acceptance criteria, slicing vertically into a walking skeleton, splitting with SPIDR or the splitting-pattern flowchart, and refining the backlog into ready items. |
| [versioning-guide](skills/versioning-guide/SKILL.md) | Bumping a package version, cutting a release, or detecting which packages changed version. |

## Commands

Use the command name exposed by the harness. Command titles use the `xonovex-core` namespace.

| Command | Purpose |
| --- | --- |
| `git-commit` | [Create a conventional commit from the changes; use `--push` only when publication is requested.](commands/git-commit.md) |
| `plan-worktree-abandon` | [Record why feature work stops and preserve its lessons before abandoning the worktree.](commands/plan-worktree-abandon.md) |
| `plan-worktree-cleanup` | [Remove eligible stale or merged worktrees and prune their remaining Git metadata.](commands/plan-worktree-cleanup.md) |
| `plan-worktree-create` | [Create an isolated worktree for a feature branch.](commands/plan-worktree-create.md) |
| `plan-worktree-merge` | [Integrate a feature worktree into its source worktree and validate the combined result.](commands/plan-worktree-merge.md) |

## Validation

Run `npx moon run plugin-core:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills and commands.
