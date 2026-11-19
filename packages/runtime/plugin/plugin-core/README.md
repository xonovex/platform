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
| [analyzing-connascence](skills/analyzing-connascence/SKILL.md) | Code feels too tightly coupled and the user wants to know how strong the coupling is, how to loosen it, or whether a module is cohesive. |
| [applying-bdd](skills/applying-bdd/SKILL.md) | Writing Gherkin feature files or Given-When-Then scenarios, or agreeing with the team on concrete examples of behaviour before coding. |
| [applying-ddd](skills/applying-ddd/SKILL.md) | Modelling a business domain in code or drawing the boundaries between parts of a system around that domain. |
| [applying-fp](skills/applying-fp/SKILL.md) | Writing functional-style code or reviewing for FP cleanliness. |
| [applying-hexagonal-architecture](skills/applying-hexagonal-architecture/SKILL.md) | Separating business logic from the database, UI, framework or other I/O behind interfaces, or structuring an app as hexagonal, ports-and-adapters, clean or onion architecture. |
| [applying-microkernel-architecture](skills/applying-microkernel-architecture/SKILL.md) | Making a system extensible through plug-ins: a small core that loads interchangeable plug-ins from a registry. |
| [applying-oop](skills/applying-oop/SKILL.md) | Designing classes or class hierarchies or applying OOP principles. |
| [applying-orthogonality](skills/applying-orthogonality/SKILL.md) | Deciding how to split a system into modules or packages, or where the boundary of each concern belongs. |
| [applying-tdd](skills/applying-tdd/SKILL.md) | Writing code test-first or coaching the red-green-refactor cycle. |
| [assuring-accessibility](skills/assuring-accessibility/SKILL.md) | Making a website, app, document or service accessible, or checking, gating or claiming its accessibility conformance. |
| [auditing-code-quality](skills/auditing-code-quality/SKILL.md) | Auditing existing code for quality and reporting graded smells without changing anything. |
| [authoring-pull-requests](skills/authoring-pull-requests/SKILL.md) | An author writes or prepares their own pull or merge request for review: its title and description, its size and split, how it was tested, and a self-review before reviewers are assigned. |
| [authoring-tests](skills/authoring-tests/SKILL.md) | Writing or reviewing a single test in any framework, or choosing what to mock in it. |
| [managing-credentials](skills/managing-credentials/SKILL.md) | Storing, injecting, rotating or revoking a secret or token that a service, workload or CI job uses to reach another system, or when one has leaked. |
| [managing-versions](skills/managing-versions/SKILL.md) | Bumping a package version, cutting a release, or finding which packages changed version. |
| [reviewing-code](skills/reviewing-code/SKILL.md) | Reviewing the code of a pull request, merge request or diff for defects, or reviewing an existing module or file that is not a change, and writing the review feedback. |
| [using-git](skills/using-git/SKILL.md) | Running git commands or fixing repository state. |
| [writing-user-stories](skills/writing-user-stories/SKILL.md) | Writing, splitting or refining user stories and their acceptance criteria for a backlog. |

## Validation

Run `npx moon run plugin-core:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
