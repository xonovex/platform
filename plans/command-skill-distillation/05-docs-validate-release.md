---
type: plan
has_subplans: false
parent_plan: plans/command-skill-distillation.md
parallel_group: 4
status: complete
feature: command-skill-distillation
dependencies:
  plans:
    - plans/command-skill-distillation/00-mechanism-pilot.md
    - plans/command-skill-distillation/01-utility-distill.md
    - plans/command-skill-distillation/02-workflow-plan-git-distill.md
    - plans/command-skill-distillation/03-git-host-skill-create.md
    - plans/command-skill-distillation/04-pr-command-distill.md
  files:
    - packages/command/command-utility/README.md
    - packages/command/command-workflow/README.md
    - AGENTS.md
    - .claude-plugin/marketplace.json
    - packages/command/command-utility/.claude-plugin/plugin.json
    - packages/command/command-utility/.codex-plugin/plugin.json
    - packages/command/command-utility/package.json
    - packages/command/command-workflow/.claude-plugin/plugin.json
    - packages/command/command-workflow/.codex-plugin/plugin.json
    - packages/command/command-workflow/package.json
    - packages/skill/skill-code-review/.claude-plugin/plugin.json
    - packages/skill/skill-code-review/.codex-plugin/plugin.json
    - packages/skill/skill-code-review/package.json
    - packages/skill/skill-pull-request/.claude-plugin/plugin.json
    - packages/skill/skill-pull-request/.codex-plugin/plugin.json
    - packages/skill/skill-pull-request/package.json
    - packages/skill/skill-git-host/.claude-plugin/plugin.json
    - packages/skill/skill-git-host/.codex-plugin/plugin.json
    - packages/skill/skill-git-host/package.json
skills_to_consult:
  - command-guide
  - instruction-guide
  - git-guide
  - pull-request-guide
  - moon-guide
validation:
  type_check: n/a            # markdown + JSON only; no typed source touched
  lint: pass                 # 5 touched moon projects fmt-check green; AGENTS.md + both READMEs prettier-clean
  build: pass                # 5 touched moon projects build green
  tests: n/a                 # no unit tests for command/skill markdown
  integration: partial       # static delegation + manifest/marketplace/JSON contract verified; live invocation + post-publish marketplace lifecycle deferred to release (see Verification Results)
---

# 05 — Docs, Validate, Release

## Objective

Close out the distillation: update the two command-plugin READMEs and the root
`AGENTS.md` to describe the command↔skill dependency model, confirm the new
`xonovex-skill-git-host` plugin is registered in the marketplace, apply a consistent
release version bump, and run the final two-harness verification (Claude Code per-family
smoke test, the post-publish dependency-lifecycle confirmation on the real marketplace that
subplan `00` deferred, and confirmation that the Codex / `--plugin-dir` result recorded by
subplan `00` still holds). This is a docs/JSON-only subplan — no command/skill bodies change here;
subplans `01`–`04` produced the final command and skill surface, and this subplan documents
and ships it.

## Tasks

1. **Update both command-plugin READMEs.**
   - `packages/command/command-utility/README.md`: under `## Installation` (after the Codex
     block, before `## Commands`) add a short note that each command now hard-depends on its
     delegated skill and the skill is auto-installed with the plugin:

     > Each command delegates its procedure to a guideline skill and declares that skill in
     > `plugin.json` `dependencies`. Installing this plugin auto-installs the skills it
     > delegates to; if a depended-on skill is missing the command is disabled with
     > `dependency-unsatisfied`.

     Then make the command tables match the final command set: the **Content** table is
     currently missing `content-humanize` (the file `commands/content-humanize.md` exists) —
     add the row `| `content-humanize` | Remove AI writing patterns and add human voice |`.
     Verify the remaining utility rows still match `commands/` (18 commands across
     content/instructions/insights/skill/slashcommand).
   - `packages/command/command-workflow/README.md`: add the same dependency note in the
     `## Installation` section, and verify the single command table still matches the 21
     files in `commands/` (no command was renamed or removed by `02`/`04`).
   - Do not restate skill procedures in the READMEs — only the dependency relationship and
     the command inventory.

2. **Update root `AGENTS.md`.**
   - File: `AGENTS.md`. Under `## Integration Points`, extend the dependency line to record
     the command→skill model, e.g. add a bullet:

     > - command plugins depend on guideline skills via `plugin.json` `dependencies`:
     >   `command-utility` → `skill-{content,insights,instruction,skill,command}`;
     >   `command-workflow` → `skill-{plan,git,pull-request,code-review,git-host}`

   - In `### Packages`, the `skill` bullet already reads "coding guidelines and skills"; no
     new top-level package category is needed (`skill-git-host` lives under `packages/skill/`).
     Only add a mention of `git-host` if it improves clarity; keep the edit minimal and in the
     existing terse bullet style.

3. **Register `git-host` skill in marketplace + apply the release version bump.**
   - File: `.claude-plugin/marketplace.json`. Confirm subplan `03` added the
     `xonovex-skill-git-host` entry (alphabetical order, after `xonovex-skill-git`):

     ```json
     { "name": "xonovex-skill-git-host", "source": "./packages/skill/skill-git-host", "description": "Xonovex Git host skills" },
     ```

     If `03` did not add it (or placed it wrong), add/fix it here so the registry is complete.
   - Version bump decision: distillation is **behavior-preserving**, so this is a **minor**
     bump (`3.0.0` → `3.1.0`), not major. All 60 plugin manifests are currently at exactly
     `3.0.0` (lockstep), so bump consistently to keep the single-version invariant: set
     `metadata.version` in `marketplace.json` to `3.1.0`, and bump every
     `packages/*/*/.claude-plugin/plugin.json`, `packages/*/*/.codex-plugin/plugin.json`, and
     `packages/*/*/package.json` whose `"version"` reads `"3.0.0"` to `"3.1.0"`. Do **not**
     touch the repo-root `package.json` (`0.1.0`, unrelated workspace version).
   - The touched-package minimum that MUST land at `3.1.0`: `command-utility`,
     `command-workflow`, `skill-code-review`, `skill-pull-request`, and the new
     `skill-git-host` (across `.claude-plugin`, `.codex-plugin`, and `package.json` each).
   - Suggested mechanical bump (verify the count before/after, expect the same N files):

     ```bash
     grep -rl '"version": "3.0.0"' packages/*/*/.claude-plugin/plugin.json \
       packages/*/*/.codex-plugin/plugin.json packages/*/*/package.json \
       | xargs sed -i 's/"version": "3.0.0"/"version": "3.1.0"/'
     ```

4. **Repo-wide `fmt:check` + `build` across touched projects.**
   - Lint (prettier `fmt:check`) and build the touched moon projects:

     ```bash
     npx moon run command-utility:fmt-check command-workflow:fmt-check skill-git-host:fmt-check \
       skill-code-review:fmt-check skill-pull-request:fmt-check
     npx moon run command-utility:build command-workflow:build skill-git-host:build \
       skill-code-review:build skill-pull-request:build
     ```

   - If the JSON edits in tasks 1–3 reflow under prettier, run `:fmt` (write) on the affected
     projects and re-run `:fmt-check`. The `marketplace.json` and root `AGENTS.md` are covered
     by the repo-root prettier config — run `npx prettier --check .claude-plugin/marketplace.json AGENTS.md`
     and `--write` if needed. fmt:check and build must be green before the smoke test.

5. **Per-family invocation smoke test on Claude Code.**
   - Install/load the two command plugins (marketplace install, or the `--plugin-dir` dev loop
     validated by subplan `00`) and invoke at least one command per family, confirming the
     command loads its skill via the `Skill` tool at run time and the output matches the
     pre-distillation behavior:
     - utility: `content-humanize` (content), `instructions-simplify` (instruction),
       `insights-extract` (insights — already proven in `00`, re-confirm), `skill-guide-simplify`
       (skill), `slashcommand-simplify` (command).
     - workflow: `plan-validate` (plan), `git-commit` (git), `pr-review-analyze` (pr-* /
       code-review + findings-schema).
   - For each: confirm the run shows the delegated skill being loaded and the result is
     unchanged from the fat-command baseline (spot-check output shape, not byte-identical).
     Record pass/fail per command in the plan-update notes.

6. **Confirm Codex / `--plugin-dir` final state and document it.**
   - Re-read the Codex `.codex-plugin` `dependencies` + command→Skill-load result and the
     `--plugin-dir` dependency-resolution result that subplan `00` recorded. Confirm those
     conclusions still hold against the now-complete surface (both command plugins carry their
     full `dependencies`, `skill-git-host` exists and is registered).
   - Run one Codex invocation (e.g. `pr-review-analyze` or `insights-extract`) to confirm the
     full dependency set still resolves and loads on Codex; if `00` concluded Codex is
     unsupported / worked-around, just re-confirm the documented workaround still applies.
   - Capture the final supported/worked-around state in the READMEs' Installation note (if a
     caveat is needed) and in the parent plan's Success Criteria checkoff. No code change if the
     status is unchanged from `00` — this task is verification + documentation only.

7. **Post-publish dependency-lifecycle confirmation on the real marketplace (carried over from
   subplan `00`'s deferred GATE (a)).** Subplan `00` proved the dependency *mechanism* against the
   working tree via `--plugin-dir` (it read the new `dependencies`, failed closed with the exact
   error type `dependency-unsatisfied` when the skill was absent, and loaded the skill cleanly when
   present). It could not exercise the path through the *published* `xonovex-marketplace` (GitHub
   `xonovex/platform`) or the interactive UI surfaces, because that needs the release commit/tag
   pushed — out of scope until release. Run this once the release is published and
   `claude plugin marketplace update xonovex-marketplace` has pulled it:
   - **Auto-install (positive case):** at a clean scope, run
     `claude plugin install xonovex-utility@xonovex-marketplace` and confirm every delegated skill
     it declares (`xonovex-skill-insights` plus the rest added by subplan `01`) auto-installs at the
     **same scope** without being named on the command line. Repeat for `xonovex-workflow` (its
     `plan`/`git`/`pull-request`/`code-review`/`git-host` skills). Verify with `claude plugin list`.
   - **`dependency-unsatisfied` (negative case) across every surface:** run
     `claude plugin uninstall xonovex-skill-insights`, then confirm the three dependent `insights-*`
     commands are disabled with `dependency-unsatisfied` in **all three** surfaces the parent plan
     names — `claude plugin list`, the `/plugin` TUI, and `/doctor`. Record the **exact** status
     string each surface prints verbatim (they are the user-facing wording and differ from the
     debug-log string subplan `00` captured). Afterward, re-install the skill (or run
     `claude plugin install --prune` cleanup) to restore the satisfied state.
   - This discharges the deferred half of `00`'s GATE (a). The deferred live-invocation half of
     `00`'s GATE (b) — actually running `/xonovex-utility:insights-extract` and watching the `Skill`
     tool fire — is covered by the insights entry in task 5 above; while there, also note whether the
     `Skill` load is deterministic across two or three repeated invocations (the determinism question
     `00` flagged).

## Validation Steps

- **type_check**: n/a (markdown + JSON only; no typed source touched).
- **lint**: `npx moon run command-utility:fmt-check command-workflow:fmt-check skill-git-host:fmt-check skill-code-review:fmt-check skill-pull-request:fmt-check` plus `npx prettier --check .claude-plugin/marketplace.json AGENTS.md` — all clean.
- **build**: `npx moon run command-utility:build command-workflow:build skill-git-host:build skill-code-review:build skill-pull-request:build` — all green.
- **tests**: n/a (no unit tests for command/skill markdown).
- **integration**: invoke one command per family on Claude Code (task 5) and confirm each loads its delegated skill via the `Skill` tool and produces unchanged output; re-confirm on Codex that the full dependency set resolves (task 6) and the `00`-recorded `--plugin-dir` behavior still holds; and, post-publish, run the real-marketplace dependency-lifecycle confirmation (task 7) — auto-install on `plugin install` and `dependency-unsatisfied` on skill uninstall across `claude plugin list`, `/plugin`, and `/doctor`. Also confirm `jq empty .claude-plugin/marketplace.json` and `jq empty` over each bumped `plugin.json`/`package.json` parse cleanly and `xonovex-skill-git-host` is present in the marketplace `plugins` array.

## Success Criteria

- [x] Both command READMEs state that each command hard-depends on (and auto-installs) its delegated skill, and their command tables match `commands/` exactly (utility 18 incl. `content-humanize`; workflow 21).
- [x] Root `AGENTS.md` records the command→skill `dependencies` model; edit stays in the existing terse bullet style.
- [x] `.claude-plugin/marketplace.json` contains the `xonovex-skill-git-host` entry in alphabetical position (line 36, after `xonovex-skill-git`) and its `source` path (`./packages/skill/skill-git-host`) is correct — already added by subplan `03`; no repair needed.
- [x] Version bump applied consistently: `marketplace.json` `metadata.version` and all previously-`3.0.0` plugin/`package.json` manifests (183 files) are at `3.1.0`; root `package.json` (`0.1.0`) untouched; no manifest left at `3.0.0`.
- [x] `fmt:check` and `build` green across all touched moon projects; `AGENTS.md` and both READMEs pass prettier. **Deviation:** `marketplace.json` is intentionally compact (one line per plugin) and is *not* effectively prettier-governed — it fails `prettier --check` at `HEAD` too (pre-existing, no prettierignore), and `--write` would explode it into ~250 multi-line entries. Convention preserved; the version-bump and git-host edits match the existing compact style. Not formatted with `--write`.
- [~] **Static half done; live half deferred.** All 39 commands (18 utility + 21 workflow) statically verified to allow the `Skill` tool and delegate to the matching `*-guide` skill named in their plugin's `dependencies` (per-family representatives spot-checked: `content-humanize`→content, `instructions-simplify`→instruction, `insights-extract`→insights, `skill-guide-simplify`→skill, `slashcommand-simplify`→command, `plan-validate`→plan, `git-commit`→git, `pr-review-analyze`→code-review). Live interactive invocation observing the `Skill` tool fire + output parity vs the fat baseline is deferred — it needs an interactive harness session (carries over `00`'s deferred live-invocation half of GATE (b)).
- [x] Codex and `--plugin-dir` final state re-confirmed against the complete surface: both command plugins carry full `dependencies` in `.claude-plugin` and `.codex-plugin`, and `skill-git-host` exists + is registered, so `00`'s conclusions stand unchanged — Codex silently ignores `dependencies` (no auto-install; non-blocking workaround = install the skill alongside the command), `--plugin-dir` enforces but does not auto-install. The README Installation note records the Codex caveat. A live Codex invocation is deferred (no Codex session here).
- [ ] **Post-publish — deferred to release (carried over from `00`'s deferred GATE (a)).** Requires the release commit/tag pushed to `xonovex/platform` and `claude plugin marketplace update`; nothing is published and the project rule forbids pushing unless explicitly asked. Run after release: real `xonovex-marketplace` install of `xonovex-utility`/`xonovex-workflow` auto-installs their delegated skills at the same scope (`claude plugin list`); uninstalling a depended-on skill disables its dependent commands with `dependency-unsatisfied` across `claude plugin list`, the `/plugin` UI, and `/doctor` — capture the exact per-surface strings.

## Verification Results

Environment: file-editing session on `main` (no push). Static + build verification performed;
interactive-harness and published-marketplace steps deferred as noted.

**Docs (tasks 1–2): DONE.**
- `command-utility/README.md`: added a `### Dependencies` note under Installation and a
  `content-humanize` row to the Content table (now 18 rows total matching `commands/`).
- `command-workflow/README.md`: added the same `### Dependencies` note; the command table
  already listed all 21 commands — no row change needed.
- Both notes state the Claude-Code auto-install + `dependency-unsatisfied` behavior and add the
  Codex caveat (deps not auto-installed; install the skill alongside the command) per `00`'s GATE (c).
- `AGENTS.md`: added one terse bullet under Integration Points recording the verified model —
  `command-utility` → `skill-{content,insights,instruction,skill,command}`,
  `command-workflow` → `skill-{plan,git,pull-request,code-review,git-host}`.

**Marketplace + version bump (task 3): DONE.**
- `xonovex-skill-git-host` already registered by `03` at the correct alphabetical position with the
  correct `source`; confirmed via `jq`, no edit needed.
- Bumped `3.0.0` → `3.1.0`: `marketplace.json` `metadata.version` + 183 `packages/*/*` manifests
  (`.claude-plugin/plugin.json`, `.codex-plugin/plugin.json`, `package.json`). 0 manifests left at
  `3.0.0`; root `package.json` (`0.1.0`) untouched. All bumped JSON + marketplace parse clean (`jq empty`).

**fmt + build (task 4): PASS.**
- `moon run …:fmt-check` green for command-utility, command-workflow, skill-git-host,
  skill-code-review, skill-pull-request (per-project `prettier --check .` covers the README edits).
- `moon run …:build` green for the same five.
- `prettier --check` clean for `AGENTS.md` and both READMEs. `marketplace.json` deviation documented
  above (intentionally compact, pre-existing `--check` failure, not `--write`-formatted).

**Delegation contract (tasks 5–6, static half): VERIFIED.**
- All 39 commands allow the `Skill` tool and delegate to a `*-guide` skill; each named skill matches
  its plugin's `dependencies`. Codex/`--plugin-dir` conclusions from `00` re-confirmed against the
  complete surface (unchanged).

**Deferred to an interactive session / release (tasks 5 live, 6 live, 7):**
- Live slash-command invocation per family observing the `Skill` tool fire + output parity, and the
  `Skill`-load determinism question `00` flagged — needs an interactive harness run.
- A live Codex invocation re-confirming the dependency set resolves on Codex.
- Post-publish real-marketplace lifecycle (auto-install positive + `dependency-unsatisfied` negative
  across `claude plugin list`, `/plugin`, `/doctor`, with exact per-surface strings) — needs the
  release pushed/published first.

## Files Modified / Created

- `packages/command/command-utility/README.md` — dependency note + add `content-humanize` row.
- `packages/command/command-workflow/README.md` — dependency note.
- `AGENTS.md` — command→skill dependency-model bullet under Integration Points.
- `.claude-plugin/marketplace.json` — confirm `xonovex-skill-git-host` registration; `metadata.version` → `3.1.0`.
- Version bump to `3.1.0` (from `3.0.0`) in every `packages/*/*/.claude-plugin/plugin.json`, `packages/*/*/.codex-plugin/plugin.json`, and `packages/*/*/package.json` — touched-package minimum: `command-utility`, `command-workflow`, `skill-code-review`, `skill-pull-request`, `skill-git-host`.
- No new files created (the `skill-git-host` package is created by subplan `03`; this subplan only registers and version-aligns it).

## Dependencies

- **Must land after `00`–`04`**, because this subplan documents, registers, and ships the *final* surface:
  - `01` (utility-distill) and `02` (workflow-plan-git-distill) finalize the thin tier-1 commands and the `command-utility` / `command-workflow` `dependencies` — the READMEs and `AGENTS.md` describe that model and the smoke test invokes those commands.
  - `03` (git-host-skill-create) creates and (per its scope) registers `xonovex-skill-git-host`; this subplan confirms/repairs the marketplace entry and aligns its version.
  - `04` (pr-command-distill) finalizes the `pr-*` commands and `code-review`/`pull-request`/`git-host` craft — the per-family smoke test (task 5) and the Codex re-confirm (task 6) exercise `pr-review-analyze`.
  - `00` (mechanism-pilot) recorded the authoritative Codex `.codex-plugin` and `--plugin-dir` results that task 6 re-confirms rather than re-derives.
- Runs alone in parallel group 4 (no sibling runs concurrently), so there is no manifest-contention risk; the version bump intentionally rewrites manifests other subplans authored, which is safe only because they have all landed.

## Estimated Duration

~half a day: ~1.5h docs (READMEs + AGENTS.md), ~0.5h marketplace check + scripted version bump, ~0.5h fmt/build, ~1.5h two-harness smoke test and final documentation of Codex/`--plugin-dir` state.
