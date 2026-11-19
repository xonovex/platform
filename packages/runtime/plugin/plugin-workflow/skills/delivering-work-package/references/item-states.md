# The package, the order and the state of each item

## The package

The package is the work package file the invoker names, with the work items whose `parent` is that file. When the invoker names none and the change has one package file, use it; with more than one, ask which. When the work items have no `parent`, the package is every work item of the change that is not `deprecated`, as the layout reference of **recording-handoff** defines.

## The order

Sort the items so that each comes after every item in its `depends_on`, and break a tie by the order of `work/index.md`. A cycle is a validation error of the bundle: report it and stop before the first agent.

Run the items in this order, one at a time. Report the order and every skip with its reason before the first agent starts, and recompute the skips after each item.

A `depends_on` item outside the package counts as done only by the proof that **implementing-work-item** accepts: a verdict whose results are all `pass`, or the invoker's word, recorded in the log entry's `authorization`.

## The state of an item

Read the state from the files and from Git, never from memory. "Current hash" is the tree hash of the item's worktree, from the command in [commit-and-record.md](commit-and-record.md); the implementation notes name the worktree and the branch. Take the first row that holds, and start only the steps it names.

| State | Evidence | Steps the item needs |
| --- | --- | --- |
| Committed | An all-`pass` verdict, and the item's branch holds a commit whose tree (`git rev-parse '<branch>^{tree}'`) equals its `worktree_hash` | None; it counts as done |
| Delivery written | An all-`pass` verdict that matches the current hash, and a `commit-message.md` with matching `work` and `branch` and a later `generated.at` than the verdict | The commit only |
| Validated | An all-`pass` verdict that matches the current hash, and no newer `commit-message.md` | **preparing-delivery**, then the commit |
| Failed on this tree | A verdict with a `fail` or `blocked` result that matches the current hash | None; the item stops again, and its fix goes through **implementing-work-item** on the invoker's word |
| Implemented | Implementation notes and a worktree that holds a change, with no verdict or one that does not match the current hash | **validating-work-item**, **preparing-delivery**, then the commit |
| Not started | No implementation notes, or a worktree that holds no change | All three steps, then the commit |

A verdict with `source: self-check` covers nothing here: treat the item as implemented and run a new validation.

## Skipped and stopped items

- A `readiness: waiting` item is skipped. Name the open decisions in its `waits_on`; **settling-questions** settles them.
- An item whose `depends_on` names a skipped or stopped item, directly or through another item, is skipped. Name the item it waits on.
- An item with no `repository`, or one that the `repositories` map does not hold, is skipped with that reason.
- An item stops when a step agent stops on a precondition or a question, when its verdict is not all `pass`, or when its commit fails.

A skip or a stop never ends the package.

## The base of each item

The base of an item is the branch of the last item committed in the same repository, by this run or an earlier one, or the confirmed branch of the main checkout when there is none. Name the base in the brief of **implementing-work-item**, which creates the worktree from it.

A stopped item adds nothing to the stack. When a later run finishes it, its branch forks from the stack: report the branches that the last branch does not contain, and leave their integration to the invoker through **using-git**.
