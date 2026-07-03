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
    - packages/skill/skill-git-host/moon.yml
    - packages/skill/skill-git-host/package.json
    - packages/skill/skill-git-host/prettier.config.ts
    - packages/skill/skill-git-host/.claude-plugin/plugin.json
    - packages/skill/skill-git-host/.codex-plugin/plugin.json
    - packages/skill/skill-git-host/git-host-guide/SKILL.md
    - packages/skill/skill-git-host/git-host-guide/references/create.md
    - packages/skill/skill-git-host/git-host-guide/references/review-post.md
    - packages/skill/skill-git-host/git-host-guide/references/review-resolve.md
    - packages/skill/skill-code-review/code-review-guide/SKILL.md
    - packages/skill/skill-code-review/code-review-guide/references/findings-schema.md
    - packages/skill/skill-code-review/code-review-guide/references/review-analyze.md
    - packages/skill/skill-code-review/code-review-guide/references/review-refine.md
    - .claude-plugin/marketplace.json
skills_to_consult: [skill-guide, command-guide, moon-guide, git-guide, code-review-guide]
validation:
  type_check: n/a
  lint: pass
  build: pass
  tests: n/a
  integration: pass
---

# 03 — git-host-skill-create

## Objective

Create the new `xonovex-skill-git-host` skill package — the "matching host skill" that
`code-review-guide` line 8 already points to but does not yet exist — and give it the
host-specific delivery operations (`create`, `review-post`, `review-resolve`) lifted from
the current `pr-*` command bodies. In parallel, add the host-**independent** review
operations and findings contract to `code-review-guide` (`findings-schema`, `review-analyze`,
`review-refine`) and resolve its dangling "matching host skill" pointer to `git-host-guide`.
Finally, register the new plugin in `marketplace.json`.

This subplan **authors the skills only** — it edits no commands. The reference files here
become the single source of truth; subplan `04-pr-command-distill` deletes the duplicated
bodies from the five `pr-*` commands and wires the `xonovex-workflow` plugin `dependencies`.
No command-plugin `dependencies` wiring happens in this subplan.

Resolved open sub-decision (parent plan, line 132): the findings JSON contract and the
analyze/refine pipeline live in **`code-review-guide`** (platform-independent review
methodology), **not** in `git-host-guide`. `git-host-guide` owns only host I/O.

## Tasks

1. **Scaffold the `skill-git-host` package files.** Mirror the sibling `skill-pull-request`
   package exactly.
   - `packages/skill/skill-git-host/moon.yml` — copy verbatim from
     `packages/skill/skill-pull-request/moon.yml`:
     ```yaml
     $schema: https://moonrepo.dev/schemas/project.json
     language: unknown
     tags: [skill]
     tasks:
       build:
         command: echo 'build complete'
         options:
           cache: false
     ```
   - `packages/skill/skill-git-host/package.json` — name `@xonovex/skill-git-host`,
     `version 3.0.0`, `private: true`, `type: module`, keywords `["xonovex","skill","git-host"]`,
     and the same `fmt` / `fmt:check` scripts as the sibling.
   - `packages/skill/skill-git-host/prettier.config.ts` — `export {default} from "@xonovex/prettier-config";`
   - `packages/skill/skill-git-host/.claude-plugin/plugin.json` — note the array form of
     `skills` and the compact `author` (matches `skill-pull-request/.claude-plugin`):
     ```json
     {
       "name": "xonovex-skill-git-host",
       "version": "3.0.0",
       "description": "Xonovex Git host delivery skills",
       "author": {"name": "Xonovex"},
       "skills": ["./git-host-guide"]
     }
     ```
   - `packages/skill/skill-git-host/.codex-plugin/plugin.json` — note `skills` is a **string**
     (not array) and `author` is expanded, matching `skill-pull-request/.codex-plugin`:
     ```json
     {
       "name": "xonovex-skill-git-host",
       "version": "3.0.0",
       "description": "Xonovex Git host delivery skills",
       "author": {
         "name": "Xonovex"
       },
       "skills": "./git-host-guide"
     }
     ```

2. **Author `git-host-guide/SKILL.md`.** Path
   `packages/skill/skill-git-host/git-host-guide/SKILL.md`. Follow the
   Essentials + Operations + Progressive Disclosure shape of `insights-guide/SKILL.md`.
   Frontmatter `name: git-host-guide` and the description from the scope, completing the
   triggers, e.g.:
   ```
   description: "Use when delivering PRs/reviews to a git host — detecting the host from
   the remote and driving its CLI/API (GitHub via gh today): push/open PR, post review
   comments, promote blocking reviews, resolve threads. Triggers on opening a PR, posting
   a review to a host, gh pr create/edit/review, request-changes, resolving review threads,
   comment anchors, host detection from the remote — even when the user doesn't say 'gh'
   or name a host."
   ```
   Body sections:
   - A one-line identity: this is the host-delivery tier; the *craft* (PR description,
     review labels) lives in `pull-request-guide` / `code-review-guide`, this skill only
     drives the detected host's CLI/API.
   - **Essentials** bullets (host-shared substrate): detect host from
     `git remote get-url origin`, never hardcode; resolve repo coordinates from the remote
     each run; verify auth before any write (`gh auth status`); the workflow speaks neutral
     concepts, the host's CLI realizes each (Host Mapping); preview-and-confirm before any
     outward-facing write, `--yes` to bypass; build bodies/JSON with a serializer or heredoc,
     never hand-escape.
   - **Operations** (one bullet each, pointing at the reference):
     - **Create** — push the branch and open a PR with a drafted description, reviewers,
       work-item and sibling links — see `references/create.md`
     - **Review-post** — publish a structured review (anchored inline comments + summary),
       promote blockers, cross-link, write back comment ids — see `references/review-post.md`
     - **Review-resolve** — verify each finding is fixed, resolve its thread, optionally
       reply — see `references/review-resolve.md`
   - **Progressive Disclosure** — one "Read … - Load when …" line per reference.

3. **Author the three `git-host-guide` references** (source content FROM the current command
   bodies named below; subplan `04` deletes it there). Keep each to the host-driving
   procedure only — leave craft to the craft skills.
   - `packages/skill/skill-git-host/git-host-guide/references/create.md` — lift from
     `command-workflow/commands/pr-create.md`: host detection + "coordinates from the remote
     each run"; the **Host Mapping** table (push branch / create PR / set description / draft
     / reviewers / link work-item / cross-link siblings → `gh` realizations); the push →
     rebase-onto-base → `gh pr create --base --head` → enrich (`--reviewer`, work-item ref,
     sibling URLs) → verify-merge-status sequence; and the **additive** `gh pr edit --body`
     rule (it replaces the whole body, so read-back → append → re-set). Note: an open PR
     already exists → refresh rather than duplicate; `--draft` is a no-op where unsupported;
     the SSH "post-quantum key exchange" line is informational.
   - `packages/skill/skill-git-host/git-host-guide/references/review-post.md` — lift from
     `command-workflow/commands/pr-review-post.md`: the anchored inline-comment path
     (`path` + `line` + `side`, confirm it anchored not orphaned); the `**<label> (<decoration>)**`
     bold lead-in prefix, idempotent (skip if body already starts with the label);
     `--request-changes` promotion of blocking findings (and that it only *gates* merge where
     the host enforces it); the `html_url` deep-link cross-link as the **final** edit (ids do
     not exist until posted); and the `commentId` write-back onto each finding so
     `review-resolve` can match later.
   - `packages/skill/skill-git-host/git-host-guide/references/review-resolve.md` — lift from
     `command-workflow/commands/pr-review-resolve.md`: read the PR's open blocking threads;
     match findings to threads by `commentId` first, fall back to `path` + body similarity
     (never `line`, which shifts); resolve via `resolveReviewThread` (GraphQL); reply with
     `in_reply_to` id when `--reply`; and the verify-before-resolve gate (judge the code at
     the anchor, a moved line is not a fixed line; skip already-resolved threads).

4. **Add the three `code-review-guide` references, the SKILL.md disclosure, and resolve the
   dangling pointer.**
   - `packages/skill/skill-code-review/code-review-guide/references/findings-schema.md` — the
     **single** owner of the findings JSON contract (currently restated in both
     `pr-review-analyze.md` and `pr-review-post.md`, already drifted). One canonical shape
     (`summary` + `findings[]` with `path` / `line` / `lineType` / `label` / `decoration` /
     `blocking` / `body`, plus the optional `status` `new|recurring` and `commentId` that the
     pipeline writes), the hunk-header `@@ -a,b +c,d @@` new-file line-number parsing
     (walk `+`/context lines; `ADDED` vs `CONTEXT`), and the "build/edit with a serializer
     (`python3` + `json`), never hand-escape" rule.
   - `packages/skill/skill-code-review/code-review-guide/references/review-analyze.md` — lift
     the host-independent methodology from `command-workflow/commands/pr-review-analyze.md`:
     compute `git diff <base>...<branch>`, read for correctness first then quality, anchor
     every finding to a real new-file `ADDED`/`CONTEXT` line, the effort dial
     (low/medium = fewer high-confidence, high = wider recall leaning on `question`), and the
     `--since` prior-findings comparison (match by `path` + body similarity, tag
     `new`/`recurring`/`gone`, carry `commentId` — compares findings only, never the PR, so it
     stays platform-independent).
   - `packages/skill/skill-code-review/code-review-guide/references/review-refine.md` — lift
     from `command-workflow/commands/pr-review-refine.md`: the per-finding operations
     (keep / reword / relabel / re-anchor / merge / split / drop), re-derive the summary,
     re-validate the anchor after re-anchor/split, label discipline (every finding keeps a
     known label + explicit decoration; dropping `blocking` flips the decoration), idempotency,
     and **STOP after each pass** for the next review round (`--final` runs full validation).
   - Edit `packages/skill/skill-code-review/code-review-guide/SKILL.md` line ~8 so "load the
     matching host skill" names `git-host-guide` (e.g. "…load the host skill `git-host-guide`
     (plugin `xonovex-skill-git-host`)."), and add three matching Progressive-Disclosure
     bullets (one per new reference) plus an Operations/Essentials cross-link as appropriate.

5. **Register the plugin in `marketplace.json`.** Add to the `plugins` array of
   `/.claude-plugin/marketplace.json` (alphabetical neighbours are `…-git` and
   `…-gpu-rendering`), matching the existing entry shape:
   ```json
   { "name": "xonovex-skill-git-host", "source": "./packages/skill/skill-git-host", "description": "Xonovex Git host delivery skills" },
   ```
   There is exactly one marketplace file in the repo (`/.claude-plugin/marketplace.json`); no
   separate codex marketplace exists, so no second registration is needed.

6. **Validate.** Run prettier `fmt:check` on the touched packages, `moon build` for the
   touched skill projects, and confirm the skill resolves (frontmatter parses, all
   progressive-disclosure pointers point at files that exist, marketplace JSON is valid) and
   that `code-review-guide`'s host-skill pointer now resolves to `git-host-guide`. See
   Validation Steps.

## Validation Steps

- **Lint (prettier fmt:check on touched packages):**
  - `npx moon run skill-git-host:fmt-check` (or `npm --prefix packages/skill/skill-git-host run fmt:check`)
  - `npx moon run skill-code-review:fmt-check` (or the package's `fmt:check`)
  - `npx prettier --check .claude-plugin/marketplace.json`
- **Build (moon build for touched projects):**
  - `npx moon run skill-git-host:build`
  - `npx moon run skill-code-review:build`
  - `npx moon query projects --tags skill` lists `skill-git-host` (new project is discovered).
- **JSON validity:** `python3 -c 'import json,sys; [json.load(open(p)) for p in sys.argv[1:]]'`
  over `packages/skill/skill-git-host/.claude-plugin/plugin.json`,
  `packages/skill/skill-git-host/.codex-plugin/plugin.json`, and
  `.claude-plugin/marketplace.json` exits 0.
- **Integration (the skill loads; no command is invoked in this subplan):**
  - In Claude Code with `--plugin-dir`, confirm `git-host-guide` appears in the skill catalog
    with its description and the three operations, and that loading it via the `Skill` tool
    succeeds.
  - Confirm every reference path named in both SKILL.md files exists on disk (no dangling
    progressive-disclosure link): `git-host-guide` references `create.md` / `review-post.md` /
    `review-resolve.md`; `code-review-guide` references `findings-schema.md` /
    `review-analyze.md` / `review-refine.md`.
  - Confirm `code-review-guide/SKILL.md` no longer says "matching host skill … that does not
    exist" — it now names `git-host-guide` and that target exists.

## Success Criteria

- [x] `packages/skill/skill-git-host/` exists with `moon.yml` (`tags:[skill]`, echo build),
      `package.json` (`@xonovex/skill-git-host`), `prettier.config.ts`, and both
      `.claude-plugin/plugin.json` (`skills` array) + `.codex-plugin/plugin.json` (`skills` string).
- [x] `git-host-guide/SKILL.md` has `name: git-host-guide`, the host-delivery description,
      and Essentials + Operations + Progressive Disclosure sections.
- [x] `git-host-guide/references/{create,review-post,review-resolve}.md` carry the host-driving
      procedures lifted from the three `pr-*` command bodies, craft excluded.
- [x] `code-review-guide` has `references/{findings-schema,review-analyze,review-refine}.md`,
      the findings schema has exactly one owner, and matching disclosure bullets are in SKILL.md.
- [x] `code-review-guide/SKILL.md`'s "matching host skill" reference resolves to `git-host-guide`.
- [x] `xonovex-skill-git-host` is registered in `.claude-plugin/marketplace.json`.
- [x] fmt:check, `moon build`, and JSON validity are green on all touched packages; the skill
      loads and every progressive-disclosure pointer resolves.

## Files Modified / Created

Created:
- `packages/skill/skill-git-host/moon.yml`
- `packages/skill/skill-git-host/package.json`
- `packages/skill/skill-git-host/prettier.config.ts`
- `packages/skill/skill-git-host/.claude-plugin/plugin.json`
- `packages/skill/skill-git-host/.codex-plugin/plugin.json`
- `packages/skill/skill-git-host/git-host-guide/SKILL.md`
- `packages/skill/skill-git-host/git-host-guide/references/create.md`
- `packages/skill/skill-git-host/git-host-guide/references/review-post.md`
- `packages/skill/skill-git-host/git-host-guide/references/review-resolve.md`
- `packages/skill/skill-code-review/code-review-guide/references/findings-schema.md`
- `packages/skill/skill-code-review/code-review-guide/references/review-analyze.md`
- `packages/skill/skill-code-review/code-review-guide/references/review-refine.md`

Modified:
- `packages/skill/skill-code-review/code-review-guide/SKILL.md` (resolve dangling pointer + add disclosure bullets)
- `.claude-plugin/marketplace.json` (register `xonovex-skill-git-host`)

## Dependencies

- **`00-mechanism-pilot`** must land first: it proves the dependency/runtime contract and the
  marketplace + `--plugin-dir` resolution this new skill relies on. No skill or command should
  be added before that gate passes.
- Independent of siblings `01-utility-distill` and `02-workflow-plan-git-distill`: different
  packages (`skill-git-host`, `skill-code-review`, `marketplace.json` vs `command-utility` /
  `command-workflow`), so it runs in parallel in group 2 with no file overlap.
- **`04-pr-command-distill` depends on this**, not the reverse: it consumes `git-host-guide`
  and the new `code-review-guide` references when it thins the five `pr-*` commands and wires
  `xonovex-workflow` `dependencies`. The duplicated command bodies are only removed there.

## Estimated Duration

~1 day (new package scaffold + 3 host-op references + 3 code-review-op references authored from
the existing `pr-*` bodies, plus SKILL.md edits, registration, and validation).
