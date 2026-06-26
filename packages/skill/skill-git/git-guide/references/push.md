# Push a branch and keep its diff clean

The local-git steps that precede opening a PR / MR: publish the branch upstream, and
rebase it onto the base so the diff is just this change. Driving the host — opening the PR,
posting reviews — is the host skill's job (`github-guide`, `gitlab-guide`, or another
`skill-<host>`); this file is the git half they defer to.

## Push the branch upstream

```bash
git push -u origin <branch>        # publish and set the upstream in one step
```

- Pushing a branch does **not** open a PR / MR, and opening one does not push — do them in
  order. The host's create step has nothing to open against until the branch is pushed.
- An SSH push may print a "post-quantum key exchange" line — it is informational, not an error.
- Nothing committed ahead of the base → there is nothing to push or open; say so.

## Rebase onto the base so the diff is just this change

If the branch is behind its base, a stale merge-base inflates the PR diff with unrelated
commits. Rebase onto the base before opening, then re-push:

```bash
git fetch origin <base>
git rebase origin/<base>           # replay this branch's commits on top of the base
git push --force-with-lease        # never plain --force on a shared branch
```

- Use `--force-with-lease`, not `--force` — it refuses to overwrite commits you have not seen,
  so a teammate's push is not silently clobbered.
- Conflicts during the rebase → resolve per [merge-resolve.md](merge-resolve.md), then
  `git rebase --continue`.
- Rebasing rewrites history; only rebase a branch that is yours / unshared, and never rebase
  or force-push `main` / `master`.
