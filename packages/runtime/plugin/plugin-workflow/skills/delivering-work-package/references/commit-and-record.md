# Commit and record an item

## Contents

- The tree hash
- The commit
- The log entry
- The report of an item
- The report of the package

## The tree hash

Compute the tree hash inside the item's worktree, at its root, with a temporary index so the real index stays untouched:

```bash
(export GIT_INDEX_FILE="$(mktemp -d)/index"; git add -A && git write-tree)
```

The variable must reach both commands. Written as `GIT_INDEX_FILE=x git add -A && git write-tree`, it reaches only `git add`, and `git write-tree` prints the tree of the real index. Compare the result with the verdict's `worktree_hash` as plain strings. Never adjust a hash and never copy one from an earlier verdict.

## The commit

Commit only when all of these hold:

- The verdict's results are all `pass`, its `source` is `independent`, and its `worktree_hash` equals the hash you just computed.
- `deliveries/<work-id>/commit-message.md` exists, validates, and its `work` and `branch` match the item and the worktree's branch.

Then commit in the worktree through **using-git**:

1. Write the body of `commit-message.md`, without its frontmatter, to a temporary file outside the worktree.
2. Stage the whole worktree with `git add -A` and commit with `git commit -F <file>`.
3. Check the result: `git rev-parse 'HEAD^{tree}'` equals the verdict's `worktree_hash`, and `git status --short` is empty.

A hook that fails, or a commit whose tree differs from the verdict's hash, stops the item. Report the hook output or the two hashes. Never pass `--no-verify`, amend the commit or reset the branch to get past it; the invoker decides. A stopped commit adds nothing to the stack, so the next item builds on the base before it.

The commit holds the item's code only. The files of the change directory that the steps wrote stay uncommitted in the main checkout; name them in the report of the package.

## The log entry

Write one log entry for each item the run reached, a skipped item included, because a later run reads the state of the package from the log. A run that started no step and committed nothing writes no log entry. Use the template and the rules of **recording-handoff**:

- `skill: delivering-work-package`, `action: Update`, and no `touched`: each step agent lists the files it wrote in its own log entry.
- The summary sentence names the item and its result: committed, stopped with the reason, or skipped with the reason.
- After a blank line, a table of the steps: the step, the model and the effort of its agent, and its result. Then the verdict counts, the first 12 characters of the tree hash, the commit ID and the branch.

Then run `index --write`, `log --write` and `validate` with the script of **recording-handoff**.

## The report of an item

Report each item in a few lines, the result first:

```text
Item 2 of 4: work/search-api-k3m7, committed
Base: search-store-q2fa (the branch of item 1)
Steps: implementing-work-item (mid tier, medium) done · validating-work-item (mid tier, medium) 3 pass · preparing-delivery (cheapest tier, low) done
Checked: notes valid · diff inside the item · verdict independent, one result per criterion · tree 4b825dc642cb matches
Commit: 9f31c2a feat(search): add saved-search API, on branch search-api-k3m7
Log: log/2026-11-19-delivering-work-package-r4tb.md
```

A stopped item names the step, the criterion or the question, and the items it skips.

## The report of the package

After the last item, report:

- A table of every item and its result: committed with its commit ID, stopped with its reason, or skipped with its reason.
- The last branch of each repository, which holds every committed item, and any branch that it does not contain.
- The questions that wait for the invoker, and the open decisions that skipped items wait on.
- The files of the change directory that stay uncommitted in the main checkout.
- The next steps, offered and not started: **validating-work-package** for the package, a fix of a stopped item through **implementing-work-item** on the invoker's word, and a push or a pull request through **using-git**, **using-github** or **using-gitlab** when the invoker asks.
