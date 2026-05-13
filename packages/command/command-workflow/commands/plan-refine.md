---
description: >-
  Process user annotations in a plan document and refine iteratively until
  approved
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
argument-hint: "[plan-file] [--final]"
---

# /xonovex-workflow:plan-refine – Refine Plan from Annotations

Process user feedback on a plan document — from inline annotations in the plan, from instructions in the prompt, or both. Resolves every item, updates the plan in place, and presents the result for another review round. Repeats until the user approves.

## Prerequisites

- A plan document exists (created by `/xonovex-workflow:plan-create` or `/xonovex-workflow:plan-tdd-create`)
- The user has provided feedback as inline annotations, prompt instructions, or both

## Goal

- Find and resolve every piece of user feedback
- Update the plan in place, preserving structure and frontmatter
- Present a summary of changes for the next review round
- STOP after each pass (user reviews, annotates again, or approves)

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)
- `--final` (optional): Treat this as the final pass — after resolving annotations, mark plan as `approved` in frontmatter

## Feedback Sources

### Inline annotations

Markers the user adds directly in the plan markdown:

- `<!-- NOTE: ... -->` — HTML comment annotations
- `> **NOTE:** ...` — Blockquote annotations
- `[!NOTE] ...` or `[!FIXME] ...` — GitHub-style callout annotations
- `// ...` at end of a line — Inline comment annotations
- `~~struck-through text~~ replacement text` — Strikethrough-then-replace pattern
- Lines prefixed with `>>>` or `<<<` — Insertion/deletion markers
- Any line containing `TODO:`, `FIXME:`, `NOTE:`, `QUESTION:` (case-insensitive)

### Prompt instructions

Refinement requests stated in the user's prompt (e.g., "swap manual validation for zod schemas and drop the caching layer"). Each distinct instruction is a feedback item. When the user names a section, scope the change there; otherwise infer affected sections from the plan.

## Core Workflow

**IMPORTANT: Do NOT switch into a plan-authoring mode. Do NOT implement anything. This command only refines the plan document.**

1. **Locate plan**: Read the plan file from argument, git config, or most recent `plans/*.md`
2. **Collect feedback**: Gather annotations (with line numbers) from the document AND parse refinement requests from the user's prompt
3. **Report findings**: List every feedback item (annotation or prompt instruction) before making changes — if zero found in either source, inform the user and STOP
4. **Resolve each item**: Process top-to-bottom:
   - **Correction** (e.g., "use zod schemas here, not manual validation"): Update the relevant plan section to incorporate the correction
   - **Deletion** (e.g., "drop the caching layer, not needed"): Remove the section or item
   - **Addition** (e.g., "also need a retry strategy for webhook delivery"): Add the requested content in the appropriate location
   - **Question** (e.g., "SSE or WebSockets for live updates?"): Use AskUserQuestion to resolve, then update
   - **Rejection** (e.g., "this won't scale — rethink the batch processing"): Rework the rejected section with a revised approach
   - **Scope change** (e.g., "move notifications to a follow-up"): Remove the scoped-out items and adjust dependent sections
5. **Remove annotation markers**: Strip all annotation syntax after resolving (the plan should read cleanly)
6. **Reconcile dependencies**: If changes affected subplan structure, update the proposed subplans list and execution groups
7. **Update frontmatter**: Set `updated` date; if `--final`, set `status: approved`
8. **Write updated plan**: Edit the plan file in place
9. **Present change summary**: Show what was changed, then STOP

## Resolution Rules

- Never ignore a feedback item — every one must be addressed or explicitly flagged as unresolvable
- Conflicting items (annotation vs annotation, prompt vs annotation, prompt vs prompt) — use AskUserQuestion; the prompt does NOT automatically win
- If an item requires research or codebase exploration not available in context, flag it and suggest the user run `/xonovex-workflow:plan-research` before the next pass
- Maintain consistency: if a change affects a technology choice, propagate that change through all affected sections (approach, risks, subplans, success criteria)
- Preserve all untouched content exactly as-is

## Output

```
Processed plan: plans/feature-name.md

Feedback items: 5 (3 annotations, 2 prompt instructions)
  L12: Correction — "use zod schemas, not manual validation" → Updated validation approach
  L29: Deletion — "drop the caching layer" → Removed subplan 3, updated execution groups
  L47: Addition — "add retry strategy for webhooks" → Added to subplan 2 tasks
  Prompt: Question — "SSE or WebSockets?" → Asked user, chose SSE → Updated streaming approach
  Prompt: Scope change — "move notifications to follow-up" → Moved to Future Work section

Sections modified: Validation, Proposed Approach, Proposed Child Plans
Subplan structure: Changed (was 4 subplans, now 3)
```

## Examples

```bash
# Process annotations on auto-detected plan
/xonovex-workflow:plan-refine

# Process annotations on specific plan
/xonovex-workflow:plan-refine plans/auth.md

# Final pass — approve after resolving
/xonovex-workflow:plan-refine plans/auth.md --final
```

## Error Handling

- Error if plan file not found or has invalid frontmatter
- Warning if no feedback detected in either source (suggest adding annotations or describing in prompt)
- Warning if an item requires unavailable context (suggest `/xonovex-workflow:plan-research`)
- Error if conflicting items cannot be auto-resolved (falls back to AskUserQuestion)

## Gotchas

- Implementing instead of refining is the #1 mistake — this command only edits the plan
- The prompt doesn't automatically beat annotations — conflict resolution requires the user
- Stripping annotation markers without resolving the underlying intent loses information silently
- A change that touches a technology choice but isn't propagated through risks / subplans / success criteria leaves the plan internally inconsistent
