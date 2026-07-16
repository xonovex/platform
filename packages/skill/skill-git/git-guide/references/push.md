# push: Publish a Branch and Keep Its Diff Clean

The local-git steps before opening a PR / MR. Driving the host (opening the PR, posting reviews) is the host skill's job (`github-guide` / `gitlab-guide`); this is the git half they defer to.

```bash
git push -u origin <branch>        # publish and set upstream in one step
```

- Pushing does **not** open a PR / MR, and opening one does not push — do both, in order.
- Nothing committed ahead of the base → nothing to push or open; say so.
- An SSH push may print a "post-quantum key exchange" line — informational, not an error.

## Rebase onto the base so the diff is just this change

If the branch is behind its base, a stale merge-base inflates the PR diff with unrelated commits. Rebase, then re-push:

```bash
git fetch origin <base>
git rebase origin/<base>
git push --force-with-lease
```

- Conflicts during rebase → resolve per [merge-resolve.md](merge-resolve.md), then `git rebase --continue`.
- Only rebase a branch that is yours / unshared; never rebase or force-push `main` / `master`.
