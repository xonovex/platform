---
type: plan
has_subplans: false
parent_plan: composable-workflow-implementations-merge
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
  - packages/skill/skill-{aws,azure-devops,bitbucket,bitrise,datadog,accessibility,ai-governance,reliability,security-assurance}/**
  - packages/command/command-workflow/docs/platform-onboarding.md
  - packages/command/command-workflow/scripts/validate-documentation.mjs
  - plans/composable-workflow-implementations-merge/port-manifest.md
skills_to_consult:
- skill-guide
- git-guide
- command-guide
- moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 01: Baseline Gate and Layout Normalization

## Objective

Record a green pre-merge baseline in the worktree, build the port manifest from the pinned fable commit, and move the 9 nested skill packages to the repo's flat `<topic>-guide/` layout with all manifests and cross-references updated.

## Tasks

1. **Verify worktree preconditions.** `git worktree list` shows `../xonovex-platform-merge` on branch `composable-workflow-implementations-merge`; `git config branch.composable-workflow-implementations-merge.mergeBackTo` returns `main`; dependencies installed. Record the starting HEAD in `plans/composable-workflow-implementations-merge/BASELINE.txt`.
2. **Record the green baseline.** In the worktree run and capture into `BASELINE.txt`: full moon validation (`npx moon check --all` or the repo's format/lint/typecheck/build/test tasks), `node packages/command/command-workflow/scripts/validate-documentation.mjs` (expect 8,862 checks), and `go test ./internal/webhook/` from `packages/agent/agent-operator-go`. Any failure stops the subplan — the merge must start green.
3. **Build the port manifest.** Enumerate fable's port-scope files at the pin:
   ```bash
   git -C ../xonovex-platform-fable ls-tree -r --name-only 82303137 -- \
     packages/skill/skill-agent-governance/agent-governance-guide/assets \
     packages/skill/skill-workflow/workflow-guide/assets/fixtures
   ```
   plus the harness/enterprise fixture dirs, `two-plane-architecture.{dot,png}`, and `platform-skill-convention.md`. Write `plans/composable-workflow-implementations-merge/port-manifest.md` with columns: source path @82303137 | destination | disposition (port / adapt / skip-with-reason) | status.
4. **Move the 9 packages to flat layout.** For each of skill-aws, skill-azure-devops, skill-bitbucket, skill-bitrise, skill-datadog, skill-accessibility, skill-ai-governance, skill-reliability, skill-security-assurance:
   ```bash
   git mv packages/skill/skill-aws/skills/aws-guide packages/skill/skill-aws/aws-guide
   rmdir packages/skill/skill-aws/skills
   ```
   (The flat target dirs that exist as empty leftovers are reused by the move.) Update per package: `.claude-plugin/plugin.json` `"skills": ["./<topic>-guide"]`, `.codex-plugin/plugin.json` `"skills": "./<topic>-guide"`, `moon.yml` `moon-skill-validate <topic>-guide`.
5. **Update cross-references to the nested paths.** `packages/command/command-workflow/docs/platform-onboarding.md` (5 links) and `packages/command/command-workflow/scripts/validate-documentation.mjs` (hardcoded `skills/<topic>-guide/evals.json` paths). Confirm no other reference: `grep -rn 'skills/.*-guide' packages/ --include='*.json' --include='*.md' --include='*.yml' --include='*.mjs'`.
6. **Validate and commit.** Run skill validators for all 9 packages, the documentation validator, and repo lint/format; commit as `fix(skill): flatten nested guide layouts to repo convention`.

## Validation Steps

- `find packages/skill -path '*/skills/*' -name SKILL.md` → empty.
- `grep -rn '"skills": "./skills"' packages/` → empty.
- Marketplace resolution: all 93 plugin sources in `.claude-plugin/marketplace.json` and `.agents/plugins/marketplace.json` point at existing package dirs whose manifests resolve to existing guide dirs.
- `moon-skill-validate` passes for the 9 moved packages; `validate-documentation.mjs` passes; lint/format green.

## Success Criteria

- [ ] `BASELINE.txt` records all baseline commands exiting 0 at the starting HEAD
- [ ] Port manifest enumerates 100% of fable's walking-skeleton and fixture files with planned destinations
- [ ] All 9 packages use flat `<topic>-guide/` layout; no `skills/` dirs remain
- [ ] All plugin manifests, moon tasks, docs links, and validator paths updated; grep sweeps empty
- [ ] Full validation green; one conventional commit on the worktree branch

## Files Modified/Created

- Created: `plans/composable-workflow-implementations-merge/{BASELINE.txt,port-manifest.md}`
- Moved: 9 × `packages/skill/skill-*/skills/<topic>-guide/` → `packages/skill/skill-*/<topic>-guide/`
- Modified: 9 × `.claude-plugin/plugin.json`, `.codex-plugin/plugin.json`, `moon.yml`; `packages/command/command-workflow/docs/platform-onboarding.md`; `packages/command/command-workflow/scripts/validate-documentation.mjs`

## Dependencies

None (group 1 — runs first; all later subplans depend on the baseline and manifest).

## Estimated Duration

0.5–1 day.
