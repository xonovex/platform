# The worktree and the branch

A work item has one worktree and one branch in its work repository. All edits for the item happen there. Nothing is committed, so the worktree holds the whole change until delivery.

## Contents

- Naming
- Location
- Finding or creating the worktree
- State during the run

## Naming

The worktree procedure of **using-git** names the directory and the branch. Use the work file's name as the feature name, for example `capped-total-api-7k2f`, because the people reading the branch list find the work item by it.

## Location

The project's instructions say where worktrees live. The default is the sibling of the clone that the worktree procedure of **using-git** names, for example `../shop-api-feature-capped-total-api-7k2f`. Use the project's location when it gives one.

## Finding or creating the worktree

1. List the repository's worktrees with `git worktree list` and look for the branch.
2. When the worktree exists, use it. The entry step names its branch and its uncommitted files; those files are the work already done, or changes of other origin that this run leaves alone.
3. When none exists, create it with the worktree procedure of **using-git**, from the base branch the invoker confirms. Creating a local worktree is not an outward effect. Say in the report that you created it, with the path and the base.
4. When the branch exists without a worktree, check it out into a new worktree with `git worktree add <path> <branch>` after the invoker confirms.

Never switch the main clone's branch to do the work, and never reuse the worktree of another work item.

When the plan directory and the work repository are the same repository, the base branch is the one branch the invoker confirms for the main checkout, and the worktree's branch, created from it for this work item, needs no second confirmation. A worktree made from the base branch also holds a tracked copy of `plans/`. Read decisions and the work file from the main checkout's bundle, write the notes and the log entry only there, and leave the worktree's copy of `plans/` untouched, because an edit there is lost with the worktree and never reaches the bundle.

## State during the run

- Changes stay uncommitted and unstaged.
- Another process may have left uncommitted files in the worktree. Name them at the start, leave them as they are and keep them out of your summary of what changed.
- The notes file records the worktree path and the branch as they were at the end of the run.
