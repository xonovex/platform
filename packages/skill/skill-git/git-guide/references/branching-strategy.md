# branching-strategy: Trunk-Based Development vs Long-Lived Branches

**Guideline:** Prefer trunk-based development — everyone commits small changes to a single shared trunk at least daily, each push is squashed and rebased onto trunk head for a linear history, review happens asynchronously _after_ the code lands, and incomplete or risky work hides behind a feature flag — instead of a heavyweight branching model (e.g. GitFlow) where long-lived feature/develop/release branches accumulate.

**Rationale:** Long-lived branches make every problem worse the longer they live: merge conflicts grow until people avoid large refactors, developers batch ever-larger PRs to dodge the wait between dependent changes (making review harder), fundamental design flaws surface only after months, and because each team sits on its own branch there is no shared, runnable view of the product. A gate that blocks landing on review also couples your throughput to reviewer latency. Trunk-based development inverts this: tiny, frequently-integrated changes keep conflicts trivial and the product continuously runnable; squash-rebase-per-push keeps `git log`/`git bisect` linear and searchable; async post-merge review still catches issues (fix forward in the next commit, or revert if serious) without throttling progress; and feature flags let an unfinished replacement live next to the old code so it ships dark and is enabled only when ready.

**How to Apply:**

1. Make trunk (`main`) the integration point; push small, independently-shippable chunks to it frequently — aim for at least daily.
2. Squash each unit of work into one commit rebased onto trunk head, so history stays linear (see [references/commit.md](./commit.md) for message conventions).
3. Use local/short-lived branches only as a personal undo/checkpoint; they exist until you push, not for weeks.
4. Develop a risky or large change behind a feature flag, in parallel with the code it will replace, so trunk stays shippable and the new path is testable before it's complete.
5. Open the PR for _post-push_ asynchronous review; apply trivial fixes immediately as follow-up commits, and revert outright if a landed change is seriously broken.
6. Periodically do a system-level review of a completed area to catch architecture issues (which per-PR review misses) and to spread knowledge.

**Example:**

```sh
# Land a small change on trunk with linear history (squash + rebase, no merge commit).
git switch -c wip            # ephemeral local branch = undo buffer, not a long-lived branch
# ... commit freely while working ...
git switch main && git pull --rebase
git merge --squash wip       # collapse to one logical change
git commit                   # conventional message
git push                     # integrate now; review happens on the PR afterwards
git branch -D wip            # branch's job is done once pushed

# Ship an unfinished subsystem dark, behind a flag, instead of on a long-lived branch:
#   if (feature_enabled("new_renderer")) new_render(); else old_render();
```

**Gotchas:**

- Trunk-based only stays sane if changes are genuinely small and trunk is always green — without CI gating the push (or fast post-push CI + quick revert), frequent integration just spreads breakage faster.
- "Rebase each push onto trunk head" rewrites your local commits; never rebase/force-push a branch others have based work on (the same history-rewrite caution in the skill's Gotchas applies).
- Feature flags are debt: every flag is a branch in the code and a test-matrix dimension — remove the flag and the dead path once the new system wins.
- Post-merge review loses its value if nobody actually reviews; without discipline, "review after" becomes "no review." Keep a real review queue.
- This is a strategy, not a worktree mechanic — for the repo's worktree-based isolation workflow see the worktree references; the two compose (short-lived worktrees feeding a trunk).

**Related:** [references/commit.md](./commit.md), [references/merge-resolve.md](./merge-resolve.md), [references/worktree-create.md](./worktree-create.md)
