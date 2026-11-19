# The tree hash

## Contents

- Computing the hash
- Reading the change
- Taking it twice
- When the checks write into the worktree
- How the hash is used later

## Computing the hash

The tree hash of a worktree identifies the exact content that a check ran against. Compute it inside the worktree, at its root, with a temporary index so the real index stays untouched:

```bash
(export GIT_INDEX_FILE="$(mktemp -d)/index"; git add -A && git write-tree)
```

The index file does not exist yet, so `git add -A` fills the temporary index from the working tree, honors `.gitignore` and includes untracked files and deletions. `git write-tree` prints the tree object ID. The variable must be set for both commands: written as `GIT_INDEX_FILE=x git add -A && git write-tree`, it reaches only `git add`, and `git write-tree` then reads the real index and prints the wrong hash. Use an absolute path for the index file.

The hash is stable. Running it twice on an unchanged worktree gives one value, any edit to a tracked or unignored file changes it, and a clean worktree gives the tree of `HEAD`, so `git rev-parse 'HEAD^{tree}'` returns the same value. After the delivery commit, that equality shows the commit holds exactly the content that was checked.

## Reading the change

To show the change that the hash stands for, diff `HEAD` against it:

```bash
git diff --stat HEAD <hash>
git diff HEAD <hash>
```

Read the diff against the criteria: which files and behaviors each criterion should reach, and which edits belong to no criterion. An edit that belongs to no criterion is not a failure of a criterion; mention it under `# Could not check` or in the reply when it looks like scope the plan did not ask for.

An empty diff means the worktree holds no change. Report it and stop, unless the invoker says the criteria are met by content already in the base. Another process may have left uncommitted files in the worktree: name them, and check the criteria against the whole tree, because the hash covers all of it.

## Taking it twice

Compute the hash before the first check and again after the last one, and record the first. The two values are equal when the worktree stayed as it was.

- **Equal**: record the hash.
- **Different and you ran nothing that writes**: someone edited the worktree during the run. Discard the results, report the change and run again from the new hash. A verdict about content that has since moved says nothing about the content now.
- **Different because a check wrote files**: see the next section.

## When the checks write into the worktree

A build or a test run may leave files that `.gitignore` does not cover, and those files change the hash. The verdict records the hash taken before the checks, the content that was judged. After the run, the worktree no longer has that hash, so a later comparison fails until the files are gone.

Do not delete files without the invoker's word, not even ones your checks created. Name the files, say that the verdict will not match the worktree until they are ignored or removed, and tell the invoker which of the two fits the project. Then offer to run the validation again on the cleaned worktree.

## How the hash is used later

**preparing-delivery** compares `worktree_hash` with a hash it computes with the same command, as plain strings. Equal marks the criteria in the pull request description; different, or no verdict, leaves them unmarked. A branch name or a commit ID cannot serve here, because the changes stay uncommitted until delivery.

Never adjust a hash to make it match, and never copy a hash from an earlier verdict.
