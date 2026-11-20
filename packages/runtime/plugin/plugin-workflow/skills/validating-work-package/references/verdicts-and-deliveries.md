# Verdicts and deliveries

## Contents

- Matching a verdict to the worktree
- A delivery that is not applied

## Matching a verdict to the worktree

A verdict is bound to the tree hash it checked. Compute the worktree's hash now, at the worktree root:

```bash
(export GIT_INDEX_FILE="$(mktemp -d)/index"; git add -A && git write-tree)
```

The `export` must stay inside the subshell. The command respects `.gitignore`, counts untracked files and deletions, leaves the real index alone and is stable on an unchanged worktree. On a clean worktree it equals `git rev-parse 'HEAD^{tree}'`.

- Equal hash and `source: independent`: the verdict names which checks passed at exactly this tree. Re-run the cheap ones and the ones the criterion depends on most, and cite them yourself.
- Different hash: the verdict describes other code. Say so, and re-derive the evidence from scratch.
- `source: self-check`: a lead only, never evidence.
- No verdict: the criterion has no prior check to start from. Derive the evidence yourself, or leave it open.
- A worktree that no longer exists, because the branch landed: read the commit the invoker names, and say which revision you checked.

## A delivery that is not applied

**preparing-delivery** writes the commit message, the pull request description and the inline comments under `deliveries/<work-id>/` and applies none of them. When the change of a work item is not committed or not published, report its delivery as prepared and not applied: name the delivery directory and what is missing, a commit, a push or a pull request. The criteria that the change supports stay `gated` on the change landing, and the gate's owner is the user, who commits and publishes the change through **using-git** and **using-github** or **using-gitlab**. The close-out applies nothing itself.
