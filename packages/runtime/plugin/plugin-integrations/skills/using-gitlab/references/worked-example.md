# worked-example: Open an MR, post a review with an inline comment, resolve the thread

The push follows **using-git**, the description follows **authoring-pull-requests**, and the comment labels follow **reviewing-code**.

```bash
# 0. confirm identity + host before any write (a stale env token acts as the wrong user)
glab auth status

# 1. push the branch, then open the MR (--yes or it hangs in scripts)
git push -u origin HEAD
iid=$(glab mr create -R group/project \
  --source-branch "$(git branch --show-current)" --target-branch main \
  --title "feat: x" --description "<what/why/how>" \
  --reviewer alice --label backend --yes \
  | grep -oE '!\d+' | tr -d '!')

# 2a. summary note (plain, non-anchored)
glab mr note "$iid" -R group/project -m "## Review summary
Two issues inline; withholding approval until addressed."

# 2b. inline comment on an added line (glab resolves the three SHAs for you)
glab mr note create "$iid" -R group/project \
  --file src/api/users.go --line 42 \
  -m "**issue (blocking):** unchecked error"

# 2c. block by withholding approval (do NOT approve); verify gating
glab api "projects/group%2Fproject/merge_requests/$iid" | jq -r .detailed_merge_status

# 3. later: list threads, match by id, resolve, then approve to clear the gate
glab api "projects/group%2Fproject/merge_requests/$iid/discussions" \
  | jq -r '.[] | select(.notes[0].position.new_path=="src/api/users.go") | .id'
glab api --method PUT \
  "projects/group%2Fproject/merge_requests/$iid/discussions/<discussion_id>?resolved=true"
glab mr approve "$iid" -R group/project
```
