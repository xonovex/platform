# Reviewers, Auto-complete, and Draft

Add reviewers, optionally set auto-complete so the PR merges once policies pass, and use draft for work in progress. Reviewers route the PR to the right people; auto-complete merges it the moment required reviews and branch policies are satisfied so it does not sit waiting; draft signals "not ready" without un-assigning anyone.

## Commands

Add reviewers with `az repos pr reviewer add --id <id> --reviewers <email|group> [...]`. Set auto-complete with `az repos pr update --id <id> --auto-complete true`, optionally with `--squash` and `--delete-source-branch true`. Create a draft with `--draft true` on create, and flip it off with `az repos pr update --id <id> --draft false` when ready.

```bash
az repos pr reviewer add --id 12345 --reviewers you@example.com
az repos pr update --id 12345 --auto-complete true --delete-source-branch true
```

## What auto-complete waits on

Auto-complete does not itself require anything — the required reviewers and merge gates come from branch policies, and auto-complete simply waits on them. Where those gates are defined — minimum and required reviewers, build validation, comment resolution, and status checks — is owned by [repos-and-policies.md](repos-and-policies.md). Setting auto-complete on a draft PR does nothing: the PR stays open until you flip draft off.
