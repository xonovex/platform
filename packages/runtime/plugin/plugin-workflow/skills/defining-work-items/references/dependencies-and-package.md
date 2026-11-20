# Dependencies, decisions, skills and the work package

## Contents

- Dependencies
- Decisions
- Skills to consult
- Checking the graph
- No package is a valid result
- A new work package

## Dependencies

`depends_on` lists the work files that must be done before this item can start or be verified. List a blocker only when it is real:

- The item uses something the other item delivers, such as an endpoint, a schema or a library version.
- The item's criteria cannot be checked until the other item is done.

Do not list an ordering preference, work that merely touches the same area, a package, or an open decision. An open decision goes in `waits_on`, with `readiness: waiting`. A ready row can depend on a waiting row, because the dependency carries the waiting and the row has no open decision of its own.

Choose the file names of every new item first, then fill in `depends_on`, so a dependency can name a file that does not exist yet.

## Decisions

`decisions` lists the settled decisions that shape this item, as concept IDs. The decision files own their text, and the item links them. An open decision belongs in `waits_on` and never in `decisions`. Leave out a decision that holds for the whole plan and does not shape this item.

## Skills to consult

`skills_to_consult` lists installed skills this item needs beyond the plan's list, by name. Add one only when the item needs guidance that the plan and the project's settings do not already load. A name that is not installed is a blocker for the later step that loads it, so list only names you know are installed.

## Checking the graph

Before reporting, check the graph from the written files:

1. Every `depends_on` entry resolves to a work file that exists.
2. Following `depends_on` never returns to the file it started from. The script reports a cycle as an error; if it does, remove the dependency that is not a real blocker, and ask the operator when both are real.
3. No file depends on a work package.
4. `work/index.md`, regenerated with `index --write`, shows the items in dependency order under `Ready now` and `Waiting on a decision`.

## No package is a valid result

A work package file is optional. A plan whose items have no `parent` has no package; its package is the set of its work items that are not `deprecated`. Propose no package when a single item, or several items that do not deliver one outcome together, need no shared parent. Say so in the proposal, in the "No work package, and why" line, with the reason (for example "one item, so nothing to group"). Do not create a package to make the cut look complete.

## A new work package

Propose a package when the plan calls for one: several items deliver one outcome together, for example one outcome across repositories. A single item needs no package. A new package file name is `<slug>-<suffix>` with a random four-character suffix from `a-z` and `2-7`.

A package file is a work file with:

- `kind: work-package`, a `team` (the team that owns the outcome), and a `readiness` (`ready`, or `waiting` with `waits_on` when every item in it waits);
- no `criteria`, no `parent`, no `repository`;
- a body of `# Objective` that states the shared outcome in two or three sentences, and no list of its items, because the items point to it through `parent`.

Choose its file name after the outcome, not the activity. The proposal shows the package like any other file; it is written only after approval.
