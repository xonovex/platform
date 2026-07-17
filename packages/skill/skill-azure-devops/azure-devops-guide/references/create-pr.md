# Create a Pull Request

Push the branch, then open the pull request with `az repos pr create`, and capture the returned id and URL. `az repos pr create` opens a PR from an already-pushed source branch against a target; capturing the id lets you script the follow-ups — update, link, reviewers — and hand the user a clickable URL. Because `az repos pr create` does not push, the branch must already be on the remote (`git push -u origin <branch>`, see git-guide); create it beforehand or the command has nothing to open against and fails.

## Deriving the coordinates

Derive `--org`, `--project`, and `--repository` from `git remote get-url origin` (the remote encodes them as `...:v3/<org>/<project>/<repo>`), not from the configured CLI defaults. The default project may belong to a different project than the repo's, so a default-driven create can target the wrong project or fail to find the repo.

## Creating and confirming

Write the title in conventional-commit style and the description per pull-request-guide, then run create with `--org`, `--project`, `--repository`, `--source-branch`, `--target-branch`, `--title`, and `--description`. Capture the id with `--query pullRequestId -o tsv` and build the URL `.../_git/<repo>/pullrequest/<id>`. Confirm the PR is mergeable — `az repos pr show --id <id> --query mergeStatus -o tsv` prints `succeeded` when there are no conflicts against the target — and re-check it after any rebase and force-push. A PR opens as active by default; pass `--draft true` for work in progress.

```bash
id=$(az repos pr create --org https://dev.azure.com/<org> --project <project> \
  --repository <repo> --source-branch feat/x --target-branch main \
  --title "feat: x" --description "..." --query pullRequestId -o tsv)
echo "https://dev.azure.com/<org>/<project>/_git/<repo>/pullrequest/$id"
az repos pr show --id "$id" --query mergeStatus -o tsv   # -> succeeded (no conflicts)
```

Once the PR exists, refresh its body per [update-pr.md](update-pr.md), cross-link related PRs and work items per [linking.md](linking.md), and route it for review per [reviewers-autocomplete.md](reviewers-autocomplete.md).
