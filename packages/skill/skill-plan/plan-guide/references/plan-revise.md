# plan-revise: Revise Plan from User Feedback

Apply user feedback to an existing plan document — from inline annotations, prompt instructions, or both. Resolve every item, update the plan in place, present the result, STOP. Repeat until approved. For settling open decisions instead of applying feedback, use `plan-decide`.

## Feedback Sources

### Inline annotations (markers the user adds in the plan markdown)

- `<!-- NOTE: ... -->` — HTML comment
- `> **NOTE:** ...` — blockquote
- `[!NOTE] ...` / `[!FIXME] ...` — GitHub-style callout
- `// ...` at end of a line — inline comment
- `~~struck text~~ replacement` — strikethrough-then-replace
- Lines prefixed `>>>` / `<<<` — insertion / deletion markers
- Any line containing `TODO:` / `FIXME:` / `NOTE:` / `QUESTION:` (case-insensitive)

### Prompt instructions

Revision requests in the prompt (e.g. "swap manual validation for zod and drop the caching layer"); each distinct instruction is one item. Scope to a named section; otherwise infer affected sections.

## Core Workflow

**Do NOT switch into plan-authoring. Do NOT implement — this only edits the plan document.**

1. **Locate plan** — user message, git config, or most recent `plans/*.md`
2. **Collect feedback** — annotations (with line numbers) + parsed prompt instructions
3. **Report findings** — list every item before changing anything; if zero in both sources, inform the user and STOP
4. **Resolve each item** top-to-bottom by type — Correction → update section · Deletion → remove · Addition → add in place · Question → ask user then update · Rejection → rework approach · Scope change → move to Future Work, adjust dependents
5. **Strip annotation markers** so the plan reads cleanly
6. **Reconcile dependencies** — if subplan structure changed, update proposed subplans + execution groups
7. **Update frontmatter** — `updated` date; final pass sets `status: approved`
8. **Write in place**, present change summary, STOP

## Resolution Rules

- Never ignore an item — address each or flag it explicitly unresolvable
- Conflicting items (annotation vs annotation, prompt vs annotation) → ask the user; the prompt does NOT automatically win
- Items needing unavailable codebase context → flag and suggest `plan-research` first
- Propagate a technology-choice change through ALL affected sections (approach, risks, subplans, success criteria)
- Preserve untouched content exactly

## Gotchas

- Implementing instead of revising is the #1 mistake — this command only edits the plan
- The prompt doesn't automatically beat annotations — conflict resolution requires the user
- Stripping annotation markers without resolving the underlying intent loses information silently
- A technology-choice change not propagated through risks / subplans / success criteria leaves the plan internally inconsistent
