# merge: Port Elements from One Prompt into Another

Extracts elements from a source prompt and integrates them into a target prompt while strictly preserving the target's structure, style, and voice.

## Arguments

- `target` (required) — target prompt file (augmented)
- `source` (required) — source prompt file (provides elements)
- `--aspects <aspects>` (optional) — focus aspects (e.g. `"workflow,validation,error-handling"`)
- `--percentage <percent>` (optional) — intensity 10-100 (default 50)
- `--interactive` (optional) — ask clarifying questions
- `--dry-run` (optional) — preview without modifying

## Core Workflow

1. Track steps in a task list
2. Read target/source prompts
3. Analyze target's DNA (structure, style, voice, formatting, conventions)
4. Extract source elements (workflow, arguments, validation, error handling, examples)
5. Filter by aspects/percentage
6. Ask questions if `--interactive`
7. Rewrite source in target's voice, match formatting exactly
8. Preview or apply
9. Report summary

## Integration Rules

**Preserve (CRITICAL):** metadata block, section order, formatting (bullets/numbers), voice/tone, code style, argument format, example structure, spacing

**Extract from source:** workflow steps, arguments/flags, validation, error handling, safety guidelines, examples

**Style matching:** match workflow format, argument style, example format, heading caps, whitespace, vocabulary

**Approach:** rewrite in target's voice → insert in existing sections → match format exactly → adapt examples → avoid duplicates

**Percentage scale:** 10-30% critical only, 30-50% important (default), 50-70% comprehensive, 70-100% extensive

**Aspect filtering:** `workflow` (steps/sequence), `arguments` (patterns/validation), `error-handling` (cases/messages), `examples` (use cases), `validation` (rules/checks), `safety` (guidelines/warnings)

## Implementation

**Discovery:** Accept prompt-file paths or bare names; resolve to the harness-specific location and extension (see [harness-formats.md](harness-formats.md))

**Analysis:** Parse target structure → analyze formatting → detect voice → extract conventions → build template

**Extraction:** Parse source → extract workflow/arguments/validation → collect examples → filter by aspects

**Integration:** Rewrite in target's voice → insert in existing sections → merge workflows → validate consistency

## Error Handling

File not found, invalid percentage (10-100), no new content, aspect not found, incompatible prompts, style detection failed

## Safety

Recommend git commit; never modify metadata block without confirmation; preserve all target content (add only); use `--dry-run`; warn if >40% added; abort if style confidence <80%.

## Gotchas

- A target prompt's argument shape is part of its public contract — merging different argument styles silently breaks callers
- "Same workflow, different aspect" merges are the most useful; "different workflow" merges usually mean you want a new prompt instead
- Style confidence <80% means the target has too much manual customization to safely overwrite — bail
