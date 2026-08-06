---
type: plan
has_subplans: false
parent_plan: plans/release-next-versions.md
parallel_group: group-1
status: complete
dependencies:
  plans: []
  files:
    - packages/skill/*/package.json
    - packages/command/*/package.json
    - .claude-plugin/marketplace.json
    - package-lock.json
skills_to_consult:
  - versioning-guide
  - git-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 01: Catalog Downgrade to 5.1.0

## Objective

Set the plugin catalog lockstep version from the never-published 6.0.0 to 5.1.0 across all 73 skill and command packages and `.claude-plugin/marketplace.json`, in one commit on local `main`.

## Tasks

1. Verify the starting state: `grep -rl '"version": "6.0.0"' packages/skill packages/command --include=package.json | wc -l` returns 73, and `.claude-plugin/marketplace.json` line 8 reads `"version": "6.0.0"`. Working tree must be clean.
2. Set every catalog package to 5.1.0:

   ```bash
   for f in packages/skill/*/package.json packages/command/*/package.json; do
     (cd "$(dirname "$f")" && npm pkg set version=5.1.0)
   done
   ```

3. Set `"version": "5.1.0"` in `.claude-plugin/marketplace.json` (single occurrence at line 8) and format it: `npx prettier --write .claude-plugin/marketplace.json`.
4. Refresh the lockfile so it matches the workspace versions: `npm install`, then confirm `git diff package-lock.json` only shows 6.0.0 to 5.1.0 version strings.
5. Confirm no 6.0.0 remains: `grep -rn '"6\.0\.0"' packages/skill packages/command .claude-plugin package-lock.json` returns nothing.
6. Commit with a body that records the supersession, for example:

   ```
   chore: set the plugin catalog to 5.1.0 in lockstep

   Supersedes the unpublished 6.0.0 bump (38b499fd). Ships the removal of
   the caveman and fable skills, skill-ablate, and the acceptance-*, pr-*,
   story-refine, and version-bump commands under a minor by decision.
   ```

   The pre-commit hook runs `validate-lockfile.sh` and the staged `:ci-check` closure; do not bypass it.

## Validation Steps

- `npx moon run :ci-check` exits 0.
- `npx moon run :ci-publish-dry-run` exits 0.

## Success Criteria

- [ ] 73 catalog `package.json` files and `marketplace.json` read 5.1.0.
- [ ] `package-lock.json` regenerated; no `6.0.0` string remains in catalog scope.
- [ ] Both gate commands exit 0.
- [ ] Exactly one new commit on local `main`; nothing pushed.

## Files Modified

- `packages/skill/*/package.json` (71), `packages/command/*/package.json` (2)
- `.claude-plugin/marketplace.json`
- `package-lock.json`

## Dependencies

None. Runs before subplan-02 because subplan-02's dependent-ref updates and lockfile regeneration touch overlapping files.

## Estimated Duration

Short: mechanical edits plus one gate run.
