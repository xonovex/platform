# Update a Pull Request

Change a pull request's title, description, or draft state with `az repos pr update --id`, and refresh a description by reading the current body, appending, and re-setting the whole value. `--description` replaces the entire body, so passing only a new note wipes the rest. New commits update the PR automatically on push, but the description does not — refresh it deliberately.

## Editing fields

Run `az repos pr update --id <id>` with `--title "..."`, `--description "..."`, or `--draft true|false`. To append to the description, read the current body first with `az repos pr show --id <id> --query description -o tsv`, add the note, then re-set the whole value — passing only the new note to `--description` overwrites everything already there. Read a multi-line description back afterward to confirm it rendered.

```bash
cur=$(az repos pr show --id 12345 --query description -o tsv)
az repos pr update --id 12345 --description "$cur

## Note
Rebased onto origin/main."
```

## After new commits

After pushing new commits, no re-creation is needed — the PR tracks the branch, and force-pushing a rebased branch updates the PR in place. Re-check that it still merges cleanly with `az repos pr show --id <id> --query mergeStatus -o tsv` returning `succeeded`, the same mergeability check applied when the PR is first created in [create-pr.md](create-pr.md).
