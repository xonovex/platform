# plan-refine: Refine Plan from User Feedback

Process user feedback on a plan document — from inline annotations in the plan, from instructions in the prompt, or both. Resolve every item, update the plan in place, present the result for review. Repeat until approved.

## Prerequisites

- A plan document exists (from `plan-create` or `plan-tdd-create`)
- The user has provided feedback as inline annotations, prompt instructions, or both

## Goal

- Find and resolve every piece of user feedback
- Update the plan in place, preserving structure and frontmatter
- Present a summary of changes for the next review round
- STOP after each pass (user reviews, annotates again, or approves)

## Feedback Sources

### Inline annotations

Markers the user adds directly in the plan markdown:

- `<!-- NOTE: ... -->` — HTML comment annotations
- `> **NOTE:** ...` — blockquote annotations
- `[!NOTE] ...` or `[!FIXME] ...` — GitHub-style callout annotations
- `// ...` at end of a line — inline comment annotations
- `~~struck-through text~~ replacement text` — strikethrough-then-replace
- Lines prefixed with `>>>` or `<<<` — insertion / deletion markers
- Any line containing `TODO:` / `FIXME:` / `NOTE:` / `QUESTION:` (case-insensitive)

### Prompt instructions

Refinement requests stated in the user's prompt (e.g., "swap manual validation for zod schemas and drop the caching layer"). Each distinct instruction is a feedback item. When the user names a section, scope the change there; otherwise infer affected sections from the plan.

## Core Workflow

**IMPORTANT: Do NOT switch into a plan-authoring mode. Do NOT implement anything. This command only refines the plan document.**

1. **Locate plan** — read from user message, git config, or most recent `plans/*.md`
2. **Collect feedback** — gather annotations (with line numbers) + parse prompt instructions
3. **Report findings** — list every item before making changes; if zero found in either source, inform the user and STOP
4. **Resolve each item** top-to-bottom:
   - **Correction** ("use zod here, not manual validation") → update the section
   - **Deletion** ("drop the caching layer") → remove section / item
   - **Addition** ("also need a retry strategy") → add in appropriate location
   - **Question** ("SSE or WebSockets?") → ask user, then update
   - **Rejection** ("this won't scale — rethink batch processing") → rework with revised approach
   - **Scope change** ("move notifications to follow-up") → move to Future Work, adjust dependents
5. **Strip annotation markers** — the plan should read cleanly after refining
6. **Reconcile dependencies** — if changes affected subplan structure, update proposed subplans + execution groups
7. **Update frontmatter** — `updated` date; final pass: `status: approved`
8. **Write the updated plan** in place
9. **Present change summary**, STOP

## Resolution Rules

- Never ignore a feedback item — every one must be addressed or explicitly flagged as unresolvable
- Conflicting items (annotation vs annotation, prompt vs annotation, prompt vs prompt) → ask the user; the prompt does NOT automatically win
- Items needing unavailable context (codebase exploration) → flag and suggest running `plan-research` first
- Propagate technology-choice changes through ALL affected sections (approach, risks, subplans, success criteria)
- Preserve untouched content exactly

## Output

```
Processed plan: plans/feature-name.md

Feedback items: 5 (3 annotations, 2 prompt instructions)
  L12: Correction — "use zod schemas, not manual validation" → Updated
  L29: Deletion — "drop the caching layer" → Removed subplan 3, updated execution groups
  L47: Addition — "add retry strategy for webhooks" → Added to subplan 2 tasks
  Prompt: Question — "SSE or WebSockets?" → Asked user, chose SSE → Updated
  Prompt: Scope change — "move notifications to follow-up" → Moved to Future Work

Sections modified: Validation, Proposed Approach, Proposed Child Plans
Subplan structure: Changed (was 4 subplans, now 3)
```

## Error Handling

- Plan file not found / invalid frontmatter → error
- No feedback in either source → warning, suggest adding annotations or describing in prompt
- Item needs unavailable context → warn, suggest `plan-research`
- Conflicting items can't auto-resolve → fall back to asking user

## Gotchas

- Implementing instead of refining is the #1 mistake — this command only edits the plan
- The prompt doesn't automatically beat annotations — conflict resolution requires the user
- Stripping annotation markers without resolving the underlying intent loses information silently
- A change that touches a technology choice but isn't propagated through risks / subplans / success criteria leaves the plan internally inconsistent
