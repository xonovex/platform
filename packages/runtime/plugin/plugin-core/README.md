# Core Plugin

Apply software design, testing, review, version control, credentials, and accessibility practices.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-core` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-core@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-core@xonovex-marketplace
```

Optional companions: `xonovex-game-engine`, `xonovex-integrations`, `xonovex-workflow`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [analyzing-connascence](skills/analyzing-connascence/SKILL.md) | Grading or loosening the coupling between two pieces of code, or judging how cohesive a module is. |
| [applying-bdd](skills/applying-bdd/SKILL.md) | Driving software from concrete agreed examples of behaviour: running three-amigos discovery and example mapping before coding, formulating Given-When-Then scenarios in feature files, and treating them as living documentation that tests the team their shared understanding. ATDD and BDD are one practice here. |
| [applying-ddd](skills/applying-ddd/SKILL.md) | Finding and naming domain boundaries and modelling inside them: establishing a ubiquitous language, drawing bounded contexts and a context map, protecting a model with an anti-corruption layer, and applying the tactical building blocks (entity, value object, aggregate + root, domain event, repository, domain/application service). |
| [applying-fp](skills/applying-fp/SKILL.md) | Writing functional-style code or reviewing for FP cleanliness. |
| [applying-hexagonal-architecture](skills/applying-hexagonal-architecture/SKILL.md) | Isolating an application or domain core from its I/O and delivery mechanisms behind interfaces: hexagonal / ports-and-adapters / clean / onion architecture. |
| [applying-microkernel-architecture](skills/applying-microkernel-architecture/SKILL.md) | Building an extensible system: a minimal core plus interchangeable plug-ins selected through a registry. The microkernel / plug-in architecture. |
| [applying-oop](skills/applying-oop/SKILL.md) | Designing class hierarchies or applying OOP principles. |
| [applying-orthogonality](skills/applying-orthogonality/SKILL.md) | Deciding how to decompose a system into modules or packages along independent variation axes, and where each concern's boundary belongs. |
| [applying-tdd](skills/applying-tdd/SKILL.md) | Driving code test-first or coaching the red-green-refactor rhythm: writing a failing test before the production code, going green with fake-it / obvious-implementation / triangulation, then refactoring to remove duplication, working a test list one item at a time, and letting the tests grow the design. |
| [assuring-accessibility](skills/assuring-accessibility/SKILL.md) | Planning, assessing, governing, or operating accessibility across web, mobile, desktop, documents, content, and service journeys. |
| [auditing-code-quality](skills/auditing-code-quality/SKILL.md) | Auditing existing code for quality WITHOUT changing it: a read-only pass that finds smells, grades them by severity, and routes each to its owner. |
| [authoring-pull-requests](skills/authoring-pull-requests/SKILL.md) | Authoring a pull or merge request - writing the description, sizing and splitting the change, documenting how it was tested, surfacing tradeoffs, and getting it review-ready before assigning reviewers. |
| [authoring-tests](skills/authoring-tests/SKILL.md) | Writing or reviewing a single good test independent of any framework: structuring it as Arrange-Act-Assert / Four-Phase, meeting FIRST, naming it, choosing and naming the right test double (Dummy / Stub / Spy / Mock / Fake), deciding what to mock and what not to, telling state from behaviour verification, and spotting test smells. |
| [managing-credentials](skills/managing-credentials/SKILL.md) | Choosing, storing, injecting, rotating, revoking, or responding to exposure of machine-to-machine credentials and secrets: the tokens a service, workload, or CI job presents to another system. |
| [managing-versions](skills/managing-versions/SKILL.md) | Bumping a package version, cutting a release, or detecting which packages changed version. |
| [reviewing-code](skills/reviewing-code/SKILL.md) | Writing, structuring, or labelling code-review feedback on a pull or merge request: Conventional Comments labels (praise / nitpick / suggestion / issue / question / thought / chore / todo), the blocking vs non-blocking vs if-minor decorations, pairing one top-level summary with line-anchored inline comments, cross-linking the summary to its details instead of 'see comment 3', and verifying each claim against an authoritative source before asserting it. Platform-independent review craft. |
| [using-git](skills/using-git/SKILL.md) | Running git operations or resolving repo-state issues. |
| [writing-user-stories](skills/writing-user-stories/SKILL.md) | Writing, evaluating, splitting, or refining user stories: applying INVEST, the 3 Cs (Card / Conversation / Confirmation), the 'As a / I want / so that' template, writing acceptance criteria, slicing vertically into a walking skeleton, splitting with SPIDR or the splitting-pattern flowchart, and refining the backlog into ready items. |

## Validation

Run `npx moon run plugin-core:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
