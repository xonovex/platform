# Project settings

Record the settings once in a `Workflow` section of the project's instructions, so every workflow step reads the same values. A setting with a default needs no entry. A setting marked explicit has no default, and a step that applies something outward stops when it is missing.

## Contents

- The settings
- A complete example
- The skills map
- Default skills for each phase
- Missing settings

## The settings

| Setting | Default | Meaning |
| --- | --- | --- |
| `plans` | `plans/` | The directory that holds the bundles: one bundle for each change, directly or one level down in a family directory. |
| `repositories` | none | Explicit when more than one repository is involved. A map from a name to `{ resource, host }`. |
| `delivery` | none | Explicit. `mode` is `branch`, `trunk` or `pull-request`. |
| `skills` | none | A map from a phase name to the installed skills to load in that phase. The phase name is the name of the workflow step. |

`repository` on a work item and `name` in a plan's `repositories` are keys of the repositories map. `host` is the kind of host that decides which integration skill applies, such as **using-github** or **using-gitlab**.

A project with families keeps the same `plans` setting: a bundle is `<plans>/<change>/` or `<plans>/<family>/<change>/`, and a family directory holds a `roadmap.md` that is not a bundle (see [layout.md](layout.md)). A step that starts a change in a family is told the family by the invoker; it never guesses one from the name of the change.

## A complete example

The example is invented. Put the section in the project's instructions file, for example under `## Workflow`.

````markdown
## Workflow

```yaml
plans: docs/change-plans/
repositories:
  shop-app:
    resource: https://git.example.org/shop/shop-app
    host: github
  shop-api:
    resource: https://git.example.org/shop/shop-api
    host: github
delivery:
  mode: pull-request
skills:
  researching: [analyzing-connascence]
  writing-plan: [applying-bdd]
  defining-work-items: [writing-user-stories]
  implementing-work-item: [applying-tdd, authoring-tests]
  preparing-delivery: [authoring-pull-requests, using-git]
```
````

## The skills map

A step loads skills from up to four sources. From most to least specific:

1. The `skills_to_consult` list of the plan and of the work item the step works on.
2. A custom workflow skill the project provides for the phase.
3. The `skills` map in the project settings.
4. The defaults below, only for a phase that the project's `skills` map does not name.

The sources add up: a step loads the union. When two sources disagree on the same concern, the more specific source wins. A skill named in a `skills_to_consult` list that is not installed is a blocker: report its name and stop, because the plan or the work item asked for that guidance by name. A skill named only by the project's `skills` map or by a default that is not installed is reported and skipped, so a default list never blocks a project that installs fewer skills. When no source names a skill for the phase, the step loads nothing extra and says so in one line.

## Default skills for each phase

The defaults name skills of the xonovex catalog. A phase that is not listed has no default.

| Phase | Default skills |
| --- | --- |
| `researching` | `auditing-code-quality`, `analyzing-connascence` |
| `writing-plan` | `applying-orthogonality`, `applying-hexagonal-architecture` |
| `critiquing-plan` | `analyzing-connascence`, `applying-orthogonality` |
| `defining-work-items` | `writing-user-stories`, `applying-bdd` |
| `implementing-work-item` | `applying-tdd`, `authoring-tests` |
| `validating-work-item` | `reviewing-code`, `authoring-tests` |
| `preparing-delivery` | `authoring-pull-requests`, `using-git`, `managing-versions` |
| `validating-work-package` | `reviewing-code` |
| `distilling-change` | `authoring-skills` |

## Missing settings

- A missing `plans` uses `plans/`.
- A missing `repositories` is acceptable while one repository is involved. With more than one, it is a blocker for any step that writes to a repository.
- A missing `delivery` is a blocker for a step that applies something outward: pushing a branch or opening a pull request. Report the missing setting and stop. Steps that only read and write plan files carry on.
- Never guess a value from the repository's remote. Ask the invoker to record it.
