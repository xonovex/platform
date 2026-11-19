# Matching a verdict to the worktree

## Contents

- The tree hash
- Why the hash
- The rule
- What to do with each outcome
- Self-checks

## The tree hash

The tree hash of a worktree identifies the exact content that a check ran against. Compute it inside the worktree, at its root, with a temporary index so the real index stays untouched:

```bash
(export GIT_INDEX_FILE="$(mktemp -d)/index"; git add -A && git write-tree)
```

The index file does not exist yet, so `git add -A` fills the temporary index from the working tree, honors `.gitignore` and includes untracked files and deletions. `git write-tree` prints the tree object ID. The variable must be set for both commands: written as `GIT_INDEX_FILE=x git add -A && git write-tree`, it reaches only `git add`, and `git write-tree` then reads the real index and prints the wrong hash. Use an absolute path for the index file.

The hash is stable. Running it twice on an unchanged worktree gives one value, any edit to a tracked or unignored file changes it, and a clean worktree gives the tree of `HEAD`, so `git rev-parse 'HEAD^{tree}'` returns the same value. After the delivery commit, that equality shows the commit holds exactly the content that was checked. [handoff.md](handoff.md) uses it to decide whether the report gives the commit command.

To show the change that the hash stands for, diff `HEAD` against it:

```bash
git diff --stat HEAD <hash>
git diff HEAD <hash>
```

An empty diff means there is nothing to deliver: report it and stop.

## Why the hash

A verdict records the hash of the content it checked. A verdict about earlier content says nothing about the content now, so only the hash equality ties a result to the code the pull request will carry. A branch name or a commit ID cannot do that because the changes stay uncommitted until delivery.

## The rule

1. Compute the current hash.
2. Read `verdicts/<work-id>.md`. When it is missing, `verdict` is `none` and `verdict_matches` is `false`.
3. Compare the verdict's `worktree_hash` with the current hash as plain strings. Equal: `verdict_matches` is `true`. Different: `false`.
4. Only when `verdict_matches` is `true` may the description mark a criterion. A criterion is `met` when the verdict has a result `pass` for its `id`, and `missing` when the result is `fail` or `blocked` or when none is present.

Never copy marks from a verdict that does not match, never adjust a hash to make it match, and never mark a criterion from the implementation notes.

## What to do with each outcome

- **Match**: write the marked criteria, name the verdict and the first 12 characters of the hash under how it was tested.
- **Mismatch**: write the criteria unmarked with the line saying no verdict matches this tree. Say in the reply that the worktree changed after the verdict, and that a fresh **validating-work-item** run produces a verdict for this tree. Do not start it, and give no commit command.
- **No verdict**: the same as a mismatch, with `verdict: none`.

A hash computed after the files are written is the one that counts. When the invoker edits the worktree between your run and delivery, run the skill again.

## Self-checks

A verdict with `source: self-check`, and every check in the implementation notes, comes from the implementer, not from an independent reader. It may appear in the description only with the label `self-check`. A matching self-check verdict marks criteria as `(met, self-check)`. The description never uses the word validated for it.
