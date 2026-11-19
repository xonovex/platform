# Propagation and verification

## Search for the old labels

Before reporting, search every file that could restate the changed claim: old decision concept IDs, old titles, the names of a dropped option.

```bash
rg -n '<old decision id or title>' plans/<change> --glob '*.md' --glob '!log/**'
rg -n 'FIXME' plans/<change> --glob '*.md' --glob '!log/**'
```

A hit in an append-only file is correct and stays. A hit in the plan, a work file or a derived page is stale: fix it in this run.

## Carry the change

A changed decision reaches these places, in this order:

1. **Work files** that wait on it or apply it: readiness, `waits_on`, `decisions`, criteria. Every file found by the search above.
2. **The plan**: objective, scope and success criteria that restate the old claim.
3. **The plan summary block**: rewrite it as a whole. It lists the current decisions and the work by link; a superseded decision's line is replaced by the line of its successor. Edit nothing inside the block by hand.
4. **The indexes**: the decisions and work indexes and the log come from the script.
5. **Diagrams**: this skill changes none. A diagram whose topology the change altered is listed in the reply for **illustrating-plan**.

If a step cannot be done, say why and what is needed. Do not defer a step that can be done now.

## Regenerate

```bash
uv run scripts/plan-bundle.py index plans/<change> --write
uv run scripts/plan-bundle.py log plans/<change> --write
```

Run the commands from the skill directory of **recording-handoff** or with its script's full path. Say in the reply that the run regenerated them.

## Verify

1. `uv run scripts/plan-bundle.py validate plans/<change>`: no error. A new decision file that lacks a `supersedes` back reference, or an old one without `superseded_by`, shows up here.
2. No work file with `readiness: waiting` lacks `waits_on`, and none waits on a decision whose state is `settled` and not deprecated.
3. Run the project's prose check on the files you edited and fix what it reports.
4. Compare each deprecated decision with its earlier version: only `status` and `superseded_by` differ.

## Report

Lead with the result in a sentence. Then the table of dispositions, what was not verified (nothing was run beyond reading and the validator), what changed outside the plan, the items sent to **settling-questions** and the pages that belong to other owners. Offer the next steps and stop.
