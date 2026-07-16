# optimize: Trim a Skill to Its Delta Over Baseline Model Knowledge

Cut everything the target model already produces unprompted; keep only what it lacks. Never touch `description`/triggers — they route (see [writing-descriptions.md](writing-descriptions.md)). Condensing mechanics live in [simplify.md](simplify.md); this adds the knowledge-delta discipline, tier depth, and the ablation that proves nothing essential was lost.

## Principle

A guideline skill mostly encodes conventions and non-obvious facts, not novel capability — the model can already write the code; the skill's value is the rule (a nudge) plus the facts it does not reliably know. So cutting rationale and generic procedure is safe: the model never needed them. The only risk is removing a non-obvious fact the model would get wrong without. Keep those, and verify.

## Classify the depth (tier)

- **aggressive** — generic language / general SE concept the model already owns (typescript, python, oop, tdd, patterns): strip rationale + generic procedure; keep the rule, exact identifiers, exceptions, one example.
- **moderate** — a specific framework/tool or an opinionated overlay: keep the API/opinion delta; cut filler + duplicate examples.
- **conservative** — version-pinned / exact-fact / spec-heavy (a specific API, spec numbers, this skill): light touch, dedupe only — the specifics are the value.

When unsure whether a detail is essential, keep it.

## Cut vs keep

- **Cut**: rationale that restates the obvious ("improves readability"), generic procedure a model does unprompted ("run tests", "commit clearly", "run the linter"), duplicate/near-duplicate examples, restated headings, general-knowledge filler.
- **Keep**: the rule/stance; exact identifiers (flags, lint-rule ids, signatures, versions, counts, spec limits); exception/edge cases; opinionated deltas; non-obvious gotchas; one tight example where it helps.

## Dedupe within the skill

Merge near-duplicate reference files into one, then update every SKILL.md link and the Progressive Disclosure entry (keep one Load-when trigger) and delete the unused file. Remove SKILL.md index sections that duplicate the Progressive Disclosure list. Only dedupe within one skill — cross-skill de-duplication is a [catalog-audit.md](catalog-audit.md) operation.

## Fix, don't just cut

If a trim surfaces a genuine defect (a rule that as written would mislead), fix it and note it. Do not add new guidance beyond fixing clear defects.

## Gate, then verify

1. `scripts/validate.py <skill-dir>` must PASS (frontmatter, refs resolve, Gotchas present, Progressive Disclosure Load-when triggers).
2. Verify the trim kept its value with an **ablation against the weakest model you deploy** — a stronger model hides value loss:
   - Diff the removed content (`git diff <pre-trim-ref> HEAD -- <skill>`); from the `-` lines list genuine knowledge-at-risk items (non-obvious facts, exact identifiers) — ignore filler.
   - For each, an eval whose correct answer needs that fact — the `evals.json` seed + `eval-outputs.py` runner in [evaluating-outputs.md](evaluating-outputs.md).
   - Measure the **with-skill** arm against the _repo_ content, not the installed plugin: inject the trimmed `SKILL.md` + `references/` into the weakest model (e.g. `--append-system-prompt-file`) so uncommitted edits are what gets tested — `eval-outputs.py` resolves the installed plugin, so it verifies only released content.
   - Run with-skill vs without-skill on the weakest model: `without-skill` correct → already known, cut was safe; `with-skill` correct only → essential and kept; `with-skill` **wrong** → the trim removed a fact the skill no longer conveys → restore the exact text.
   - Restore only the flagged facts, then re-validate.

## At catalog scale

Run the full loop per skill, in parallel (independent dirs, no conflicts): **baseline** — run the skill's `evals.json` prompts on the weakest model with the skill absent; **grade** — an eval it already passes marks that content known (cut), one it fails marks the fact essential (keep); **trim** — cut the known, keep the delta and everything no eval covers; **gate** — `scripts/validate.py`; **ablate** — re-run the essential (baseline-failed) evals with the trimmed skill injected and restore any it no longer conveys. Persist the per-skill `evals.json` so re-verifying against a newly added model is one pass per skill.

## Example

A TypeScript reference — 37 lines of Guideline / Rationale / Example / Techniques → 12 lines: the rule + the exact `@typescript-eslint/require-await` id + the exception cases + one example. Rationale ("adds overhead") and generic Techniques ("run tests", "commit") cut; the lint-rule id and the return-a-Promise nuance kept.
