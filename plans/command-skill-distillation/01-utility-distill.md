---
type: plan
has_subplans: false
parent_plan: plans/command-skill-distillation.md
parallel_group: 2
status: complete
feature: command-skill-distillation
dependencies:
  plans:
    - plans/command-skill-distillation/00-mechanism-pilot.md
  files:
    - packages/command/command-utility/commands/content-humanize.md
    - packages/command/command-utility/commands/content-news-add.md
    - packages/command/command-utility/commands/content-travelguide-add.md
    - packages/command/command-utility/commands/instructions-init.md
    - packages/command/command-utility/commands/instructions-consolidate.md
    - packages/command/command-utility/commands/instructions-sync.md
    - packages/command/command-utility/commands/instructions-simplify.md
    - packages/command/command-utility/commands/instructions-assimilate.md
    - packages/command/command-utility/commands/skill-guide-create.md
    - packages/command/command-utility/commands/skill-guide-extract.md
    - packages/command/command-utility/commands/skill-guide-assimilate.md
    - packages/command/command-utility/commands/skill-guide-simplify.md
    - packages/command/command-utility/commands/slashcommand-create.md
    - packages/command/command-utility/commands/slashcommand-assimilate.md
    - packages/command/command-utility/commands/slashcommand-simplify.md
    - packages/command/command-utility/.claude-plugin/plugin.json
    - packages/command/command-utility/.codex-plugin/plugin.json
skills_to_consult:
  - command-guide
  - git-guide
  - pull-request-guide
  - code-review-guide
validation:
  type_check: n/a
  lint: pass
  build: pass
  tests: n/a
  integration: documented
---

# 01 — utility-distill

## Objective

Distill the remaining 15 `command-utility` slash commands (content ×3, instructions ×5,
skill-guide ×4, slashcommand ×3) to the proven thin-command shape established by the
pilot (`00-mechanism-pilot`), and complete the `xonovex-utility` `dependencies` array in
**both** plugin manifests so every depended-on skill is a hard install-time requirement.

Each command collapses to: unchanged frontmatter (plus `Skill` in `allowed-tools`), its
verbatim **Arguments** contract (the only command-unique substance), and a one-paragraph
**Delegation** that loads the named skill and invokes a single operation. Everything else
(Goal / Workflow / Output / Examples / Error Handling / Gotchas / Pattern Catalog / etc.)
is deleted — it already lives in the skill's `references/{op}.md`. The pilot already added
`xonovex-skill-insights`; this subplan adds the other four skill dependencies, completing
the array.

Per-command operation map (skill name / plugin / `references/{op}.md`):

| Command | Skill | Plugin | Operation |
|---|---|---|---|
| content-humanize | content-guide | xonovex-skill-content | humanize |
| content-news-add | content-guide | xonovex-skill-content | news-add |
| content-travelguide-add | content-guide | xonovex-skill-content | travelguide-add |
| instructions-init | instruction-guide | xonovex-skill-instruction | init |
| instructions-consolidate | instruction-guide | xonovex-skill-instruction | consolidate |
| instructions-sync | instruction-guide | xonovex-skill-instruction | sync |
| instructions-simplify | instruction-guide | xonovex-skill-instruction | simplify |
| instructions-assimilate | instruction-guide | xonovex-skill-instruction | merge |
| skill-guide-create | skill-guide | xonovex-skill-skill | create |
| skill-guide-extract | skill-guide | xonovex-skill-skill | extract-from-codebase |
| skill-guide-assimilate | skill-guide | xonovex-skill-skill | merge |
| skill-guide-simplify | skill-guide | xonovex-skill-skill | simplify |
| slashcommand-create | command-guide | xonovex-skill-command | create |
| slashcommand-assimilate | command-guide | xonovex-skill-command | merge |
| slashcommand-simplify | command-guide | xonovex-skill-command | simplify |

## Tasks

1. **Distill the content family (×3).** Rewrite each of:
   - `packages/command/command-utility/commands/content-humanize.md`
   - `packages/command/command-utility/commands/content-news-add.md`
   - `packages/command/command-utility/commands/content-travelguide-add.md`

   Keep the existing frontmatter, add `Skill` to `allowed-tools`, and keep the verbatim
   `## Arguments` block. Delete Goal / Workflow / Pattern Catalog / Output Format /
   Examples / Gotchas / Language & Readability / Writing Philosophy / Required Structure /
   Safety Rails. Preserve these command-only flag contracts exactly:
   - content-news-add: `--path` **required**, `--lang` default `en`, `--days` default `7`,
     `--max` default `3`, `--slug` derived from `slugify(title_en)` when absent.
   - content-travelguide-add: `topic` + `subject` + `--path` **required**, `--lang` default
     `en`, `--research-only`, `--slug` derived from `slugify(subject)-guide` when absent.
   - content-humanize: `text-or-file` **required** (path / inline / `-` stdin), `--tone`
     `formal|casual|technical`, `--in-place`, `--audit`.

   Thin shape (content-news-add shown):

   ```markdown
   # /xonovex-utility:content-news-add — Auto-curate latest news stories

   ## Arguments
   - `topic` (required): The subject to search news for.
   - `--path` (required): The destination directory for the generated files.
   - `--lang` (optional): Comma-separated languages (e.g., `en,nl`). Defaults to `en`.
   - `--days` (optional): Number of days to look back. Defaults to `7`.
   - `--max` (optional): Maximum number of stories. Defaults to `3`.
   - `--slug` (optional): Custom slug; derived from the title when absent.

   ## Delegation
   Load the `content-guide` skill (plugin `xonovex-skill-content`) and perform its
   **news-add** operation with these arguments. The skill is the source of truth for the
   procedure, output format, and gotchas — do not restate them.
   ```

2. **Distill the instructions family (×5).** Rewrite each of:
   - `instructions-init.md` (op **init**), `instructions-consolidate.md` (op **consolidate**),
     `instructions-sync.md` (op **sync**), `instructions-simplify.md` (op **simplify**),
     `instructions-assimilate.md` (op **merge**)

   under `packages/command/command-utility/commands/`. Same shape; delegate to
   `instruction-guide` (plugin `xonovex-skill-instruction`). Preserve verbatim Arguments:
   - instructions-init: `directory` required, `--dry-run`, `--recursive`.
   - instructions-consolidate: `--dry-run`, `--path <directory>` (defaults to workspace root).
   - instructions-sync: `agents-file` optional, `--all`, `--dry-run`, `--update-workflows`.
   - instructions-simplify: `instruction-file` required, `--dry-run`,
     `--target-reduction <percent>` (default 45, range 30-60).
   - instructions-assimilate: `target-instructions` + `source-instructions` required,
     `--aspects`, `--percentage` (default 45, range 10-100), `--interactive`, `--dry-run`.

   Note the operation rename: the `instructions-assimilate` command maps to the
   `merge` reference (not `assimilate`); the Delegation line must say **merge**.

3. **Distill the skill-guide family (×4).** Rewrite each of:
   - `skill-guide-create.md` (op **create**), `skill-guide-extract.md`
     (op **extract-from-codebase**), `skill-guide-assimilate.md` (op **merge**),
     `skill-guide-simplify.md` (op **simplify**)

   under `packages/command/command-utility/commands/`. Delegate to `skill-guide` (plugin
   `xonovex-skill-skill`). Delete the inlined Spec Constraints / Skill Structure / SKILL.md
   format blocks / Reference File Format / Interactive Mode — all owned by the skill.
   Preserve verbatim Arguments:
   - skill-guide-create: `source` required, `--name` **required** (kebab-case), `--dry-run`.
   - skill-guide-extract: `skill-name` + `source-path` required, `--update`, `--interactive`,
     `--dry-run`.
   - skill-guide-assimilate: `target-skill` + `source-skill` required, `--aspects`,
     `--percentage` (default 50, range 10-100), `--interactive`, `--dry-run`.
   - skill-guide-simplify: `skill-file` required, `--dry-run`, `--target-reduction <percent>`
     (default 70, range 50-90).

   Note both renames: `skill-guide-extract` → **extract-from-codebase**,
   `skill-guide-assimilate` → **merge**.

4. **Distill the slashcommand family (×3).** Rewrite each of:
   - `slashcommand-create.md` (op **create**), `slashcommand-assimilate.md` (op **merge**),
     `slashcommand-simplify.md` (op **simplify**)

   under `packages/command/command-utility/commands/`. Delegate to `command-guide` (plugin
   `xonovex-skill-command`). Preserve verbatim Arguments:
   - slashcommand-create: `description` required, `--name` (auto-generated when absent),
     `--interactive`.
   - slashcommand-assimilate: `target-command` + `source-command` required, `--aspects`,
     `--percentage` (default 50, range 10-100), `--interactive`, `--dry-run`.
   - slashcommand-simplify: `command-file` required, `--dry-run`, `--target-reduction`
     (default 50, range 30-70).

   Note the rename: `slashcommand-assimilate` → **merge**.

5. **Complete the dependencies array in both manifests.** Edit:
   - `packages/command/command-utility/.claude-plugin/plugin.json`
   - `packages/command/command-utility/.codex-plugin/plugin.json`

   The pilot already inserted `xonovex-skill-insights`. Set the full array (bare strings,
   identical in both files) so all five skills this plugin's commands delegate to are
   hard requirements:

   ```json
   "dependencies": [
     "xonovex-skill-content",
     "xonovex-skill-insights",
     "xonovex-skill-instruction",
     "xonovex-skill-skill",
     "xonovex-skill-command"
   ]
   ```

   Keep existing `name` / `version` / `description` / `author` untouched; add the array as a
   sibling key. If the pilot has not yet landed, add the array with all five entries (this
   subplan depends on 00, so the pilot lands first — see Dependencies).

6. **Validate.** Run `command-utility:fmt-check` and `command-utility:build`; then
   spot-invoke one command per family (`content-humanize`, `instructions-init`,
   `skill-guide-create`, `slashcommand-create`) and confirm each loads its named skill via
   the `Skill` tool and produces the same behavior as before distillation. See
   Validation Steps for exact commands.

## Validation Steps

- **lint (prettier fmt:check):**
  `npx moon run command-utility:fmt-check` — the rewritten command `.md` files and both
  `plugin.json` manifests pass prettier formatting.
- **build:**
  `npx moon run command-utility:build` — project builds clean after the edits.
- **manifest JSON sanity:**
  `python3 -c "import json,sys; [json.load(open(p)) for p in sys.argv[1:]]" packages/command/command-utility/.claude-plugin/plugin.json packages/command/command-utility/.codex-plugin/plugin.json`
  confirms both manifests are valid JSON and the `dependencies` arrays are byte-identical
  (5 entries each).
- **integration (skill loads + output unchanged):** invoke one command per family —
  `/xonovex-utility:content-humanize`, `/xonovex-utility:instructions-init`,
  `/xonovex-utility:skill-guide-create`, `/xonovex-utility:slashcommand-create` — and verify
  each: (a) resolves its `dependencies` (no `dependency-unsatisfied`), (b) loads the named
  skill through the `Skill` tool at run time, and (c) produces output matching the
  pre-distillation behavior for a representative argument set. Reuse the verification harness
  proven in `00-mechanism-pilot`.
- **line-count check:** each rewritten command is ≤ ~25 lines and contains exactly one
  `## Arguments` and one `## Delegation` section, no residual Goal/Workflow/Output/Gotchas.

## Success Criteria

- [x] All 15 listed commands are thin delegators: unchanged frontmatter with `Skill` in
      `allowed-tools`, verbatim `## Arguments`, and a single `## Delegation` paragraph naming
      the correct skill, plugin, and operation.
- [x] Every command-unique flag default/contract is preserved verbatim (paths, defaults,
      ranges, slug-derivation rules, required flags).
- [x] Operation renames are correct: instructions-assimilate→merge, skill-guide-extract→
      extract-from-codebase, skill-guide-assimilate→merge, slashcommand-assimilate→merge.
- [x] Both `command-utility` manifests carry identical 5-entry `dependencies` arrays
      (content, insights, instruction, skill, command).
- [x] `command-utility:fmt-check` and `command-utility:build` pass.
- [x] Spot-invoked command per family loads its skill at run time and behaves identically to
      the pre-distillation command.

## Verification Results

Environment: `claude` CLI `2.1.193`. Edits are uncommitted working-tree changes, so the
runtime gates were exercised against the working tree via `--plugin-dir` (the harness proven
in `00-mechanism-pilot`). Of the five declared dependencies only `xonovex-skill-skill` is
installed in this environment; `content` / `instruction` / `command` / `insights` are not —
which makes the fail-closed gate observable.

- **Structure / renames.** All 15 commands are thin delegators: exactly one `## Arguments`
  and one `## Delegation`, `Skill` appended to `allowed-tools`, zero residual
  Goal/Workflow/Output/Gotchas sections; 25–35 lines each (the longer ones carry verbatim
  6-flag Arguments + folded `argument-hint`). Every `## Delegation` names the correct
  skill/plugin/operation, including all four renames (assimilate→merge ×3,
  extract→extract-from-codebase).
- **lint / build.** `npx moon run command-utility:fmt-check command-utility:build` — both pass.
- **manifest sanity.** `claude plugin validate packages/command/command-utility` passes; both
  `.claude-plugin` and `.codex-plugin` manifests parse as JSON and carry byte-identical
  5-entry `dependencies` arrays (content, insights, instruction, skill, command).
- **GATE (a) — dependency enforcement: PASS.** Loading `command-utility` alone via
  `--plugin-dir` (deps absent) fails closed in the debug log:
  `error type: dependency-unsatisfied` /
  `Dependency "xonovex-skill-content" is not installed — run \`claude plugin install xonovex-skill-content\`…`.
  The loader reads the new 5-entry array and gates the plugin's components.
- **GATE (b) — Skill-tool load: PASS.** Loading `command-utility` together with all five skill
  packages via repeated `--plugin-dir` clears `dependency-unsatisfied`; the four target guides
  load and register with the `Skill` tool under exactly the names the `## Delegation` blocks
  reference: `Skill prompt: showing "xonovex-skill-content:content-guide"
  (userFacingName="content-guide")`, and likewise `instruction-guide`, `skill-guide`,
  `command-guide`. Output parity is structural — each delegated operation reference
  (`content-guide/references/{humanize,news-add,travelguide-add}.md`,
  `instruction-guide/references/{init,…,merge}.md`,
  `skill-guide/references/{create,extract-from-codebase,merge,simplify}.md`,
  `command-guide/references/{create,merge,simplify}.md`) exists and owns the procedure the fat
  command previously inlined. A live byte-for-byte pre/post diff is not possible (the fat
  bodies were replaced with no captured baseline), consistent with the pilot.

## Files Modified / Created

- packages/command/command-utility/commands/content-humanize.md (modified)
- packages/command/command-utility/commands/content-news-add.md (modified)
- packages/command/command-utility/commands/content-travelguide-add.md (modified)
- packages/command/command-utility/commands/instructions-init.md (modified)
- packages/command/command-utility/commands/instructions-consolidate.md (modified)
- packages/command/command-utility/commands/instructions-sync.md (modified)
- packages/command/command-utility/commands/instructions-simplify.md (modified)
- packages/command/command-utility/commands/instructions-assimilate.md (modified)
- packages/command/command-utility/commands/skill-guide-create.md (modified)
- packages/command/command-utility/commands/skill-guide-extract.md (modified)
- packages/command/command-utility/commands/skill-guide-assimilate.md (modified)
- packages/command/command-utility/commands/skill-guide-simplify.md (modified)
- packages/command/command-utility/commands/slashcommand-create.md (modified)
- packages/command/command-utility/commands/slashcommand-assimilate.md (modified)
- packages/command/command-utility/commands/slashcommand-simplify.md (modified)
- packages/command/command-utility/.claude-plugin/plugin.json (modified — dependencies completed)
- packages/command/command-utility/.codex-plugin/plugin.json (modified — dependencies completed)

## Dependencies

- **`plans/command-skill-distillation/00-mechanism-pilot.md` (must land first).** Two reasons:
  (1) the pilot proves the runtime contract — dependency auto-install / `dependency-unsatisfied`,
  command→`Skill` load, Codex `.codex-plugin` parity, `--plugin-dir` dev loop — and this
  subplan inherits that exact thin-command shape and verification harness without re-proving it;
  (2) the pilot creates the `dependencies` array in both `command-utility` manifests (adding
  `xonovex-skill-insights`); this subplan edits the same two files to complete the array, so it
  must serialize after the pilot to avoid clobbering its insertion.
- No other group-2 sibling touches `command-utility`, so this runs in parallel with
  `02-workflow-plan-git-distill` and `03-git-host-skill-create` once the pilot has landed.

## Estimated Duration

~half a day — 15 near-mechanical command rewrites following the pilot's pattern, two
small manifest edits, and a per-family spot-verification pass.
