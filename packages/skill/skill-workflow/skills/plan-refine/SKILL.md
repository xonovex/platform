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

Process inline annotations the user has added to a plan document. Resolves every annotation, updates the plan in place, and presents the result for another review round. Repeats until the user approves.

## Prerequisites

- A plan document exists (created by `/xonovex-workflow:plan-create` or `/xonovex-workflow:plan-tdd-create`)
- The user has opened the plan in their editor and added inline annotations

## Goal

- Find and resolve every user annotation in the plan
- Update the plan in place, preserving structure and frontmatter
- Present a summary of changes for the next review round
- STOP after each pass (user reviews, annotates again, or approves)

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)
- `--final` (optional): Treat this as the final pass — after resolving annotations, mark plan as `approved` in frontmatter

## Annotation Format

Annotations are inline markers the user adds directly in the plan markdown. Recognize all of these forms:

- `<!-- NOTE: ... -->` — HTML comment annotations
- `> **NOTE:** ...` — Blockquote annotations
- `[!NOTE] ...` or `[!FIXME] ...` — GitHub-style callout annotations
- `// ...` at end of a line — Inline comment annotations
- `~~struck-through text~~ replacement text` — Strikethrough-then-replace pattern
- Lines prefixed with `>>>` or `<<<` — Insertion/deletion markers
- Any line containing `TODO:`, `FIXME:`, `NOTE:`, `QUESTION:` (case-insensitive)

## Core Workflow

**IMPORTANT: Do NOT use EnterPlanMode. Do NOT implement anything. This command only refines the plan document.**

1. **Locate plan**: Read the plan file from argument, git config, or most recent `plans/*.md`
2. **Scan for annotations**: Read the full document and collect every annotation with its line location and content
3. **Report findings**: List all annotations found with line numbers before making changes — if zero found, inform the user and STOP
4. **Resolve each annotation**: Process annotations top-to-bottom:
   - **Correction** (e.g., "use zod schemas here, not manual validation"): Update the relevant plan section to incorporate the correction
   - **Deletion** (e.g., "drop the caching layer, not needed"): Remove the section or item
   - **Addition** (e.g., "also need a retry strategy for webhook delivery"): Add the requested content in the appropriate location
   - **Question** (e.g., "SSE or WebSockets for live updates?"): Use AskUserQuestion to resolve, then update
   - **Rejection** (e.g., "this won't scale — rethink the batch processing"): Rework the rejected section with a revised approach
   - **Scope change** (e.g., "move notifications to a follow-up"): Remove the scoped-out items and adjust dependent sections
5. **Remove annotation markers**: Strip all annotation syntax after resolving (the plan should read cleanly)
6. **Reconcile dependencies**: If annotations changed subplan structure, update the proposed subplans list and execution groups
7. **Update frontmatter**: Set `updated` date; if `--final`, set `status: approved`
8. **Write updated plan**: Edit the plan file in place
9. **Present change summary**: Show what was changed, then STOP

## Resolution Rules

- Never ignore an annotation — every one must be addressed or explicitly flagged as unresolvable
- If an annotation contradicts another annotation, use AskUserQuestion to resolve the conflict
- If an annotation requires research or codebase exploration not available in context, flag it and suggest the user run `/xonovex-workflow:plan-research` before the next annotation pass
- Maintain consistency: if an annotation changes a technology choice, propagate that change through all affected sections (approach, risks, subplans, success criteria)
- Preserve all non-annotated content exactly as-is

## Output

```
Processed plan: plans/feature-name.md

Annotations found: 5
  L12: Correction — "use zod schemas, not manual validation" → Updated validation approach
  L29: Deletion — "drop the caching layer" → Removed subplan 3, updated execution groups
  L47: Addition — "add retry strategy for webhooks" → Added to subplan 2 tasks
  L63: Question — "SSE or WebSockets?" → Asked user, chose SSE → Updated streaming approach
  L81: Scope change — "move notifications to follow-up" → Moved to Future Work section

Sections modified: Validation, Proposed Approach, Proposed Child Plans
Subplan structure: Changed (was 4 subplans, now 3)

Review the updated plan:
  - Open plans/feature-name.md in your editor
  - Add annotations for any remaining issues
  - Run /xonovex-workflow:plan-refine again, or
  - Run /xonovex-workflow:plan-refine --final to approve
  - Then: /xonovex-workflow:plan-subplans-create plans/feature-name.md
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
- Warning if no annotations detected (suggest adding annotations first)
- Warning if annotation requires unavailable context (suggest `/xonovex-workflow:plan-research`)
- Error if conflicting annotations cannot be auto-resolved (falls back to AskUserQuestion)
