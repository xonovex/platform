---
type: plan
has_subplans: false
parent_plan: plans/release-next-versions.md
parallel_group: group-2
status: complete
dependencies:
  plans:
    - plans/release-next-versions/subplan-01-catalog-downgrade-5.1.0.md
  files:
    - packages/config/*/package.json
    - packages/shared/shared-core/package.json
    - packages/agent/agent-cli-go*/package.json
    - "**/CHANGELOG.md"
    - package-lock.json
skills_to_consult:
  - versioning-guide
  - npm-guide
  - git-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 02: npm Line Bumps to 0.2.0

## Objective

Bump the config npm line (10 `packages/config/*` packages plus `shared-core`) from 0.1.22 to 0.2.0 and the agent npm line (`agent-cli-go` plus its 5 platform packages) from 0.1.31 to 0.2.0, with generated changelog entries, as two commits on local `main`.

## Tasks

1. **Dry-run first** (parent-plan risk mitigation): for one config package and one agent package, run
   `npx moon run eslint-config-base:version-bump -- --type minor --dry-run` and
   `npx moon run agent-cli-go:version-bump -- --type minor --dry-run`,
   and inspect the previewed version, dependent updates, and changelog section. If the script mishandles the lockstep or dependents, fall back to manual edits per `packages/agent/AGENTS.md` and the script README.
2. **Config line**: run for each of the 11 projects (no internal deps among them, order-free):

   ```bash
   for p in eslint-config-base eslint-config-cli prettier-config ts-config-base \
            ts-config-build ts-config-cli ts-config-node vite-config-base \
            vitest-config-base vitest-config-node shared-core; do
     npx moon run "$p":version-bump -- --type minor
   done
   ```

3. Verify the config line: all 11 `package.json` files read 0.2.0; every workspace consumer's exact `@xonovex/*` ref moved 0.1.22 to 0.2.0 (`grep -rn '"0\.1\.22"' packages/ --include=package.json` returns nothing); each package gained a `## 0.2.0` `CHANGELOG.md` entry from conventional commits.
4. Refresh the lockfile (`npm install`), then commit the config line: `chore(config): version packages (0.2.0)` (include `shared-core` in the body, it versions with this line).
5. **Agent line**: bump the five platform packages first so `agent-cli-go`'s `optionalDependencies` refs update, then `agent-cli-go` itself:

   ```bash
   for p in agent-cli-go-darwin-arm64 agent-cli-go-darwin-x64 agent-cli-go-linux-arm64 \
            agent-cli-go-linux-x64 agent-cli-go-win32-x64 agent-cli-go; do
     npx moon run "$p":version-bump -- --type minor
   done
   ```

6. Verify the agent line: all 6 packages read 0.2.0; `agent-cli-go`'s five `optionalDependencies` refs read exactly `0.2.0`; `agent-cli-go-github`'s platform refs moved in lockstep (`grep -rn '"0\.1\.31"' packages/agent --include=package.json` returns nothing); `packages/agent/agent-cli-go/CHANGELOG.md` gained the `## 0.2.0` section that `github-publish` requires, generated from commits since `agent-cli-go-v0.1.31`.
7. Refresh the lockfile (`npm install`), then commit the agent line: `chore(agent-cli-go): version packages (0.2.0)` (matches the convention of `2715e6fb`).

## Validation Steps

- After each commit: `npx moon run :ci-check` exits 0 and `npx moon run :ci-publish-dry-run` exits 0.
- The dry-run publish output shows the new versions as publishable (no "already exists" skip for 0.2.0 packages).

## Success Criteria

- [ ] 11 config-line packages at 0.2.0 with changelog entries; no `0.1.22` ref remains in the workspace.
- [ ] 6 agent-line packages at 0.2.0; `optionalDependencies` and `agent-cli-go-github` refs match; `## 0.2.0` changelog section present in `agent-cli-go`.
- [ ] Internal `@xonovex/*` refs remain exact (no caret introduced).
- [ ] Both gates exit 0 after each of the two commits.
- [ ] Exactly two new commits on local `main`; nothing pushed or published.

## Files Modified

- 11 config-line and 6 agent-line `package.json` files plus their `CHANGELOG.md` files
- Consumer `package.json` files across the workspace (exact ref updates)
- `package-lock.json`

## Dependencies

Runs after subplan-01: both touch catalog `package.json` files (consumer refs) and `package-lock.json`.

## Estimated Duration

Short to medium: scripted bumps, two verification passes, two gate runs.
