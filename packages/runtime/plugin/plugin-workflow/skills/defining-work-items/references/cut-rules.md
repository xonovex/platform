# The cut rules

## What the cut decides

The cut decides which work items exist, which team owns each, which repository each changes, and whether a work package groups them. It is a proposal first: the operator sees it in the reply and approves it before any file changes, because a cut that is wrong after the files exist, and especially after work has started on them, is costly to undo.

## Group by outcome and discipline

One work item delivers one outcome that one team can verify. Group by the outcome the plan names and by the discipline that delivers it, never by the layer or the platform that happens to hold the code.

| Cut by layer or platform (avoid) | Cut by outcome (use) |
| --- | --- |
| "Database changes", "API changes", "UI changes" | "A customer sees the saved address on the checkout page" |
| "Android work", "iOS work" | "Customers on every client see the saved address" (see below) |
| "Refactor the module" | "The reporting module builds without the legacy export option" |

A layer item cannot be verified alone, because its criteria are only visible through the other layers. An outcome item can.

## Platforms and repositories

Similar work across platforms is one item when one team delivers it in one repository: its objective names the platforms or modules it covers, and the item has one set of criteria that covers each of them.

A work item has one work repository, one worktree and one branch, so the `repository` key names exactly one repository from the project's map. When the same outcome is implemented in several repositories, write one item for each repository under one work package. Each item names its platform in its title and objective, and the package carries the shared outcome. Say in the proposal that the split follows the repositories, so the operator can see why one outcome became several items.

## Keep work apart where there is no counterpart

When one side of an outcome has no counterpart on the other side (a backend change with no client change, a client that needs a server change that already exists), do not force them into one item. Each piece is its own item, and a `depends_on` or a shared package shows how they relate.

## Size of an item

An item is the right size when:

- one team delivers it and a reviewer can check its criteria without reading another team's work;
- it fits one worktree and one branch in one repository;
- it can be finished and verified before the next item starts, or it says what it waits for.

Split an item that two teams would deliver, that spans repositories, or whose first part can start while the second waits. Do not split only to make items small.

Which rows belong to the cut, the rows that wait and the split or merge of outline rows are in [outline-rows.md](outline-rows.md).
