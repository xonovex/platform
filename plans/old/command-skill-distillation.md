---
type: plan
has_subplans: true
status: pending-approval
feature: command-skill-distillation
dependencies:
  plans: []
  subplans:
    - plans/command-skill-distillation/00-mechanism-pilot.md
    - plans/command-skill-distillation/01-utility-distill.md
    - plans/command-skill-distillation/02-workflow-plan-git-distill.md
    - plans/command-skill-distillation/03-git-host-skill-create.md
    - plans/command-skill-distillation/04-pr-command-distill.md
    - plans/command-skill-distillation/05-docs-validate-release.md
parallel_groups:
  - group: 1
    plans: [mechanism-pilot]
    note: "GATE. Distill ONE family (insights-*, 3 cmds) end-to-end and prove the runtime contract before touching the other 36: (a) plugin.json `dependencies` auto-installs the skill and a missing skill yields `dependency-unsatisfied`; (b) a thin command actually loads its skill via the Skill tool at run time; (c) Codex `.codex-plugin` honours both; (d) local `--plugin-dir` dev loop resolves dependencies. Nothing else starts until this passes."
  - group: 2
    plans: [utility-distill, workflow-plan-git-distill, git-host-skill-create]
    depends_on: [1]
    note: "The three independent build tracks. utility-distill edits command-utility only; workflow-plan-git-distill edits command-workflow (plan-*/git-commit/worktree) only; git-host-skill-create scaffolds the NEW skill-git-host package + adds host-independent review-op references to code-review-guide. Different packages — run in parallel worktrees."
  - group: 3
    plans: [pr-command-distill]
    depends_on: [2]
    note: "De-dup the 5 pr-* commands' craft and repoint their orchestration at the new git-host-guide + code-review-guide ops. Needs git-host-skill-create (3) done. SERIALIZES with workflow-plan-git-distill (2) on command-workflow/.claude-plugin/plugin.json — land after it."
  - group: 4
    plans: [docs-validate-release]
    depends_on: [1, 2, 3]
    note: "READMEs + AGENTS.md, marketplace.json (register skill-git-host), version bumps, fmt/lint, and a final two-harness verification pass."
proposed_subplans:
  - mechanism-pilot
  - utility-distill
  - workflow-plan-git-distill
  - git-host-skill-create
  - pr-command-distill
  - docs-validate-release
skills_to_consult:
  - command-guide          # the thin-command structure (frontmatter, Arguments contract, delegation) and <150-line ceiling
  - skill-guide            # authoring the new skill-git-host: SKILL.md + references/{op}.md, progressive disclosure, descriptions
  - moon-guide             # scaffolding the new skill package (moon.yml, tags:[skill]) consistent with siblings
  - pull-request-guide     # main is protected — sizing/splitting these as reviewable PRs, how-tested
  - code-review-guide      # Conventional Comments on the PRs
  - git-guide              # conventional commits; the worktree-per-track flow for group 2
research_sources:
  documentation:
    - packages/command/command-utility/commands/                       # 18 tier-1 commands (content/instructions/insights/skill-guide/slashcommand)
    - packages/command/command-workflow/commands/                      # 21 commands (plan-* x12, git-commit, worktree x3, pr-* x5)
    - packages/command/command-utility/.claude-plugin/plugin.json      # gains `dependencies` (5 skills); mirror in .codex-plugin/
    - packages/command/command-workflow/.claude-plugin/plugin.json     # gains `dependencies` (6 skills incl. new git-host); mirror in .codex-plugin/
    - packages/skill/skill-insights/insights-guide/                    # pilot twin: SKILL.md Operations -> references/{extract,integrate-*}.md
    - packages/skill/skill-pull-request/pull-request-guide/SKILL.md    # author craft; line 8 disowns host mechanics
    - packages/skill/skill-code-review/code-review-guide/SKILL.md      # platform-independent review craft; line 8 points at a "matching host skill" that does NOT exist
    - packages/skill/skill-git/git-guide/references/                   # precedent: commit.md + worktree-* are tool-driving op references
    - packages/skill/skill-plan/plan-guide/references/                 # precedent: 12 plan-* op references the commands mirror
    - .claude-plugin/marketplace.json                                  # register the new xonovex-skill-git-host plugin
    - packages/skill/skill-pull-request/{package.json,moon.yml}        # scaffold shape for the new skill package
  versions:
    claude_code_min: "2.1.110"          # plugin.json `dependencies` support
    plugin_dependencies_doc: "https://code.claude.com/docs/en/plugin-dependencies"
    plugins_reference_doc: "https://code.claude.com/docs/en/plugins-reference"
    marketplace: "xonovex-marketplace (all command + skill plugins live here; no cross-marketplace deps needed)"
  design:
    - "Audit (39 commands, 2026-06): 34 are 60-90% verbatim copies of a single skill references/*.md (avg 72% overlap); the only command-unique content is frontmatter + the Arguments flag contract. These distill to thin delegators. Source: workflow wcvkhly9q."
    - "The 5 pr-* commands are the exception (8-30% overlap): genuine host orchestration + a stateful 4-stage findings pipeline. Decision (adversarial panel w90axxcpa): de-dup their craft into the existing craft skills, and extract the SHARED host-driving + findings-schema + pipeline contract into a NEW git-host ops skill (the 'matching host skill' code-review-guide already points to) — NOT into the craft skills, which deliberately disown host mechanics and auto-trigger ambiently."
    - "Hard-requirement mechanism (verified vs docs, run wcvkhly9q): plugin.json `dependencies` is the install-time hard requirement — auto-installs the skill, disables the command with `dependency-unsatisfied` if missing. It guarantees INSTALL, not runtime LOAD; the command body must still load the skill via the Skill tool (reliable because the dep guarantees presence). A command cannot rely on implicitly auto-triggering another plugin's skill."
    - "Open mechanism gates (resolve in the pilot before mass rollout): Codex `.codex-plugin` parity for `dependencies` and command->Skill load is UNDOCUMENTED; `--plugin-dir` local-dev dependency resolution is UNDOCUMENTED; both must be tested empirically."
    - "Non-goal: making all 39 commands look identical. The 5 pr-* legitimately stay thicker (host I/O, stateful pipeline, per-stage least-privilege tool scoping skills cannot gate). Target end state: 34 thin + 5 thinner-but-orchestrating."
---

# Command / Skill Distillation

## Overview

The `xonovex-utility` and `xonovex-workflow` command plugins duplicate the guideline
skills they sit beside: 34 of 39 slash commands are 60-90% verbatim copies of a single
skill `references/*.md` file. The same instructions live in two places and drift. This
plan distils each command to the minimum it uniquely owns — frontmatter and its argument
contract — and delegates the procedure to the skill, making the skill the single source
of truth. The skill becomes a **hard requirement** via the `plugin.json` `dependencies`
field, so a command can never ship without the skill it delegates to.

The 5 `pr-*` commands are handled differently (they orchestrate rather than duplicate):
their restated craft is de-duped into the existing craft skills, and the host-driving and
review-pipeline machinery they share is lifted into a new `git-host` operations skill.

## Goals

- Eliminate the command↔skill duplication for the 34 single-operation commands; one owner per concept.
- Enforce skill presence as a hard, install-time requirement (`plugin.json` `dependencies`).
- Establish a single thin-command shape: frontmatter + Arguments contract + a Skill-load delegation.
- De-dup the 5 `pr-*` commands' craft and consolidate their shared orchestration into a new `git-host` skill, resolving `code-review-guide`'s dangling "matching host skill" pointer.
- Keep the change reviewable: prove the runtime + dependency mechanism on one family before touching the rest.

## Current State

- 39 commands across two plugins; both plugins currently declare **no** `dependencies`.
- Every tier-1 command's body restates its skill twin (Goal / Workflow / Output / Gotchas / Error Handling). The skills already expose each operation through `SKILL.md` → `references/{op}.md` progressive disclosure — built to be the single source, with the commands as the anomaly.
- `pr-create` already half-follows the target pattern (`Skill` in `allowed-tools`, "load the pull request skill and the git skill") but still inlines a description template and gotchas.
- The `pr-*` findings JSON schema is restated in `pr-review-analyze.md` **and** `pr-review-post.md` and has already diverged (analyze carries `status: new|recurring`; post dropped it) — live command-to-command drift.
- `code-review-guide` line 8 points at a "matching host skill" for posting/anchoring/auth that does not exist; the `pr-*` commands are that missing layer, stranded as commands.

## Research Findings

- **Hard requirement is natively supported.** `plugin.json` `dependencies` (Claude Code ≥ 2.1.110): installing the command plugin from the marketplace auto-installs the skill plugins at the same scope; a missing/uninstalled dependency disables the command with `dependency-unsatisfied` (visible in `claude plugin list`, `/plugin`, `/doctor`). Bare-string names suffice (same marketplace; `version` optional). Adversarially verified against the official docs.
- **Install ≠ load.** The dependency guarantees the skill is present, not that its text is in context at run time. The thin command must explicitly load the skill via the `Skill` tool (reliable precisely because the dependency guarantees presence). Implicit auto-trigger of another plugin's skill is not a designed contract — do not rely on it.
- **Per-plugin dependencies.** Dependencies are declared once per command plugin, not per command file: `xonovex-utility` → 5 skills, `xonovex-workflow` → 6 skills (incl. the new `git-host`).
- **pr-\* are orchestrators, not duplicates** (8-30% overlap). Moving their orchestration into the craft skills was rejected: it contradicts both craft skills' written charters, would duplicate the host-driving boilerplate across two skills (host detection/auth is shared by author and reviewer sides), and would pollute two ambient-trigger skills with stateful gh-pipeline mechanics. The shared substrate gets one owner: a new `git-host` skill.
- **Open gates (pilot must close before rollout):** Codex `.codex-plugin` parity and `--plugin-dir` local-dev dependency resolution are both undocumented.

## Proposed Approach

### The thin-command shape (tier 1, 34 commands)

Each distilled command keeps only:
1. **Frontmatter** — `description`, `allowed-tools` (with `Skill` added), `argument-hint`.
2. **Arguments** — the flag/default contract (e.g. `--out-dir insights/`, `--max 3`, `--research-only`). This is command-unique and absent from the skills; it stays.
3. **Delegation** — "Load the `<skill>` skill (plugin `<plugin>`) and perform its **`<operation>`** operation with these arguments. The skill is the source of truth for procedure, output format, and gotchas — do not restate them."

Everything else (Goal, Workflow, Output Format, Examples, Error Handling, Gotchas) is **deleted** from the command and lives only in the skill reference. Target: ~15-25 lines/command (from 70-160).

Each command plugin gains a `dependencies` array in **both** `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json`.

### The pr-* shape (tier 2, 5 commands)

- **De-dup craft** into the existing `pull-request-guide` / `code-review-guide` (strip restated templates, label vocabulary, blocking rules).
- **New skill `xonovex-skill-git-host`** (`git-host-guide`) owns the host-specific delivery: host detection, coordinates-from-remote, auth, the Host Mapping tables, and `references/{create,review-post,review-resolve}.md` (the `gh`/GraphQL operations).
- **Host-independent review ops + contract** → `code-review-guide` gains `references/findings-schema.md` (one owner for the JSON contract), `references/review-analyze.md`, `references/review-refine.md`. These are platform-independent review methodology, consistent with that skill's identity.
- Each `pr-*` command keeps its argument parsing, workflow ordering, error handling, and per-stage `allowed-tools`; it loads (craft skill) + (git-host-guide and/or code-review-guide ops) via the `Skill` tool.
- `xonovex-workflow` depends on `pull-request-guide`, `code-review-guide`, and `git-host-guide` (plus `plan`, `git`).

> Open sub-decision for `03`: whether `findings-schema.md` / pipeline contract lives in `code-review-guide` (host-independent — default) or in the new `git-host-guide` (single pipeline package). Default: `code-review-guide`.

## Risk Assessment

- **Runtime load reliability** — if a distilled command fails to load its skill at run time, behavior silently degrades to "args + a pointer." Mitigated by the pilot's runtime test and by always using an explicit `Skill`-tool instruction, never implicit triggering.
- **Codex parity (unknown)** — if `.codex-plugin` ignores `dependencies` or can't load the skill from a command, distilled commands break on Codex. The pilot tests this; if it fails, keep Codex commands fat or gate the rollout to Claude Code.
- **`--plugin-dir` dev loop (unknown)** — local development may not resolve dependencies the way marketplace installs do. Pilot verifies; if broken, document the install-from-marketplace dev workflow.
- **Discoverability loss** — a distilled command read in isolation no longer fully documents itself. Accepted; the delegation line names the exact skill + operation.
- **pr-\* boundary error** — splitting host-specific vs host-independent wrongly could re-introduce coupling or leak `gh` specifics into the platform-independent contract. Contained to subplan `03`/`04`.
- **command-workflow plugin.json contention** — subplans `02` and `04` both edit it; they must serialize (land `02` then `04`), not run in parallel worktrees against that file.

## Proposed Child Plans

1. **mechanism-pilot** — Distill `insights-*` (3 cmds); add `xonovex-skill-insights` to `xonovex-utility` deps (both manifests); verify dependency enforcement, runtime Skill-load, Codex parity, `--plugin-dir`. Gate for everything else.
2. **utility-distill** — Distill the other 15 `command-utility` commands (content ×3, instructions ×5, skill-guide ×4, slashcommand ×3); complete `xonovex-utility` deps (content, instruction, skill, command).
3. **workflow-plan-git-distill** — Distill the 16 tier-1 `command-workflow` commands (plan-* ×12, git-commit, worktree ×3); add `skill-plan`, `skill-git` to `xonovex-workflow` deps.
4. **git-host-skill-create** — Scaffold `packages/skill/skill-git-host` (moon.yml, package.json, plugin.json, `git-host-guide/SKILL.md` + `references/{create,review-post,review-resolve}.md`); add `references/{findings-schema,review-analyze,review-refine}.md` to `code-review-guide`; register in marketplace.json.
5. **pr-command-distill** — De-dup the 5 `pr-*` commands' craft; repoint orchestration at `git-host-guide` + `code-review-guide` ops; add `pull-request`, `code-review`, `git-host` to `xonovex-workflow` deps. Lands after `03` and after `02` (shared manifest).
6. **docs-validate-release** — Update both command READMEs + AGENTS.md, marketplace.json, version bumps; run fmt/lint across touched packages; final verification on Claude Code and Codex.

## Success Criteria

- All 34 tier-1 commands are thin delegators (≤ ~25 lines) with no restated skill body; behavior unchanged when invoked.
- Both command plugins declare correct `dependencies`; uninstalling a depended-on skill disables the dependent commands with `dependency-unsatisfied`.
- A distilled command, when invoked, loads its skill at run time and produces the same output as before (verified on `insights-extract` and at least one command per family).
- The 5 `pr-*` commands carry no restated craft and no duplicated findings schema; the schema has exactly one owner; the new `git-host-guide` exists and is registered.
- `code-review-guide`'s "matching host skill" reference resolves to `git-host-guide`.
- fmt/lint/build green across all touched packages; READMEs and AGENTS.md reflect the new structure.
- Codex and `--plugin-dir` behaviors are documented (supported, or explicitly worked around).

## Estimated Effort

- **mechanism-pilot**: ~half a day (3 small rewrites + the four verification gates — the gates are the real work).
- **utility-distill**: ~half a day (15 near-mechanical rewrites + 1 manifest).
- **workflow-plan-git-distill**: ~half a day (16 near-mechanical rewrites + 1 manifest).
- **git-host-skill-create**: ~1 day (new package + 3 host op references + 3 code-review op references, authored from the existing pr-* bodies).
- **pr-command-distill**: ~1 day (5 careful rewrites preserving pipeline invariants + manifest).
- **docs-validate-release**: ~half a day.
- Total ~4 days, pilot-gated; tracks 1-3 of group 2 parallelizable.

## Appendix — command → skill/operation map (39)

`xonovex-utility` → depends on content, insights, instruction, skill, command:

| Command | Overlap | Skill plugin | Operation (reference) |
|---|---|---|---|
| content-humanize | 90% | skill-content | content-guide / humanize |
| content-news-add | 80% | skill-content | content-guide / news-add |
| content-travelguide-add | 80% | skill-content | content-guide / travelguide-add |
| insights-extract | 80% | skill-insights | insights-guide / extract |
| insights-instructions-integrate | 80% | skill-insights | insights-guide / integrate-instructions |
| insights-skills-integrate | 60% | skill-insights | insights-guide / integrate-skills |
| instructions-init | 85% | skill-instruction | instruction-guide / init |
| instructions-consolidate | 85% | skill-instruction | instruction-guide / consolidate |
| instructions-sync | 82% | skill-instruction | instruction-guide / sync |
| instructions-simplify | 80% | skill-instruction | instruction-guide / simplify |
| instructions-assimilate | 75% | skill-instruction | instruction-guide / merge |
| skill-guide-create | 82% | skill-skill | skill-guide / create |
| skill-guide-extract | 80% | skill-skill | skill-guide / extract-from-codebase |
| skill-guide-assimilate | 80% | skill-skill | skill-guide / merge |
| skill-guide-simplify | 60% | skill-skill | skill-guide / simplify |
| slashcommand-assimilate | 88% | skill-command | command-guide / merge |
| slashcommand-create | 85% | skill-command | command-guide / create |
| slashcommand-simplify | 80% | skill-command | command-guide / simplify |

`xonovex-workflow` → depends on plan, git, pull-request, code-review, git-host:

| Command | Overlap | Skill plugin | Operation (reference) |
|---|---|---|---|
| plan-worktree-create | 90% | skill-git | git-guide / worktree-create |
| plan-refine | 88% | skill-plan | plan-guide / plan-refine |
| plan-research | 85% | skill-plan | plan-guide / plan-research |
| plan-subplans-create | 85% | skill-plan | plan-guide / plan-subplans-create |
| plan-tdd-create | 85% | skill-plan | plan-guide / plan-tdd-create |
| plan-worktree-abandon | 85% | skill-git | git-guide / worktree-abandon |
| plan-continue | 80% | skill-plan | plan-guide / plan-continue |
| plan-update | 80% | skill-plan | plan-guide / plan-update |
| git-commit | 80% | skill-git | git-guide / commit |
| plan-create | 78% | skill-plan | plan-guide / plan-create |
| plan-worktree-merge | 78% | skill-git | git-guide / worktree-merge |
| plan-research-code-harden | 75% | skill-plan | plan-guide / plan-research-code-harden |
| plan-research-code-align | 72% | skill-plan | plan-guide / plan-research-code-align |
| plan-clarify | 70% | skill-plan | plan-guide / plan-clarify |
| plan-research-code-simplify | 70% | skill-plan | plan-guide / plan-research-code-simplify |
| plan-validate | 60% | skill-plan | plan-guide / plan-validate |
| pr-create | 30% | pull-request + **git-host** (+ git) | craft + git-host / create |
| pr-review-analyze | 25% | code-review | craft + code-review / review-analyze + findings-schema |
| pr-review-post | 18% | code-review + **git-host** | craft + git-host / review-post |
| pr-review-refine | 15% | code-review | code-review / review-refine + findings-schema |
| pr-review-resolve | 8% | code-review + **git-host** | judgment + git-host / review-resolve |
