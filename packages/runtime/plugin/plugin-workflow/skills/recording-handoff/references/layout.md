# Bundle layout

Every change is one directory of markdown files with YAML frontmatter, and that directory is an Open Knowledge Format v0.2 bundle. The bundle root is `plans/<change>/`; the `plans/` name is a project setting.

```
plans/<change>/
  index.md                     generated; frontmatter holds only okf_version: "0.2"
  log.md                       generated from log/ entries
  plan.md                      type: Workflow Plan
  research/<slug>-<suffix>.md  type: Workflow Research
  decisions/<slug>-<suffix>.md type: Workflow Decision
  critiques/<slug>-<suffix>.md type: Workflow Critique
  diagrams/                    non-markdown sources and images
  work/<slug>-<suffix>.md      type: Workflow Work Item
  implementation/<work-id>/notes.md   type: Workflow Implementation Notes
  verdicts/<work-id>.md        type: Workflow Verdict
  deliveries/<work-id>/        commit-message.md, pull-request.md, comments/<suffix>.md
  log/<date>-<skill>-<suffix>.md   type: Workflow Log Entry
  close-out.md                 type: Workflow Close-out Record
```

`index.md` and `log.md` are reserved names and are never concept documents. Every directory except `diagrams/` gets a generated `index.md`, including `deliveries/`, `implementation/` and their subdirectories; a skill that writes in such a directory regenerates the indexes with the script. A file's concept ID is its path inside the bundle without `.md`, for example `decisions/target-identifier-7k2f`.

## Families

A family groups related changes of one initiative under one directory:

```
plans/<family>/
  roadmap.md                   the order and the state of the changes; not a bundle
  <change>/                    one bundle, laid out as above
```

A bundle is `plans/<change>/` or `plans/<family>/<change>/`, never deeper. The family directory is not a bundle: it holds no `index.md`, `log.md` or `plan.md` of its own, and the script ignores its `roadmap.md` and every entry that is not a bundle. A bundle knows nothing of its family: concept IDs stay relative to the bundle root, and a link from one bundle to another is a plain relative link in the body, never a relation key. The roadmap is a document of the project, outside the contract of this skill; the steps read it only when the invoker names it.

## Types, owners and schemas

A schema is shared, and lives in `recording-handoff/assets/schemas/`, only when more than one skill writes that kind of document. Any other schema lives in `assets/schemas/` of its owner skill. The validator loads the `assets/schemas` directory of every skill next to it, so an owner's schema needs no registration.

| Type | Written by | Schema lives in |
| --- | --- | --- |
| Workflow Plan | writing-plan, revising-plan | recording-handoff (shared) |
| Workflow Decision | settling-questions, writing-plan, revising-plan | recording-handoff (shared) |
| Workflow Work Item | defining-work-items, revising-plan | recording-handoff (shared) |
| Workflow Log Entry | every skill | recording-handoff (shared) |
| Workflow Research | researching-change | researching-change |
| Workflow Critique | critiquing-plan | critiquing-plan |
| Workflow Implementation Notes | implementing-work-item | implementing-work-item |
| Workflow Verdict | validating-work-item | validating-work-item |
| Workflow Commit Message, Workflow Pull Request Description, Workflow Inline Comment | preparing-delivery | preparing-delivery |
| Workflow Close-out Record | validating-work-package | validating-work-package |

A shared type's schema file is named for the type in kebab case: `plan.schema.json`, `decision.schema.json`, `work-item.schema.json`, `log-entry.schema.json`, next to `common.schema.json` with the pieces they share. A document names its schema in the `schema` key (`workflow/decision/1`), and a change that breaks existing documents raises the integer.

## Who creates and who edits

| File kind | Created by | Edited afterwards |
| --- | --- | --- |
| `plan.md` | writing-plan | In place by writing-plan and revising-plan, outside the generated block of links |
| `decisions/` | The skill that settles the question | Never. When a new decision supersedes it, the old file gets `status: deprecated` and `superseded_by`, nothing else |
| `work/` | defining-work-items | In place by defining-work-items and revising-plan |
| `research/`, `critiques/` | Their owner skill | Never; a change is a new file |
| `log/` | Every skill, once for each run | Never |
| `index.md`, `log.md` | The generator in `scripts/plan-bundle.py` | Only by the generator |
| Other owned files (notes, verdicts, deliveries, `close-out.md`) | Their owner skill | As the owner skill defines |

Only the owner skill of a type creates files of that type, and a skill that is not the owner never rewrites an owned file. `recording-handoff` itself only holds the shared contract: it creates no bundle file but the generated ones and the Initialization log entry that starts a change.

## The work package is optional

A work package file (`kind: work-package`) is optional. A plan whose work items have no `parent` has no package: its package is the set of its work items that are not `deprecated`. When a package file exists, its package is that file and the work items whose `parent` is that file.

## Shared versus owned

Ask who writes the type. If one skill writes it, that skill defines the schema and the template, and other skills read the file without a schema of their own. If several skills write it, the contract moves into `recording-handoff` so no writer can drift. A skill that needs a new key on a shared type adds it here, because the key must be valid for every writer.
