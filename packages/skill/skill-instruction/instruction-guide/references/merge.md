# merge: Port Organizational Patterns Between Projects

Extracts organizational patterns from a source AGENTS.md and integrates them into a target file while strictly preserving the target's structure, style, and project-specific context.

## Core Workflow

1. Track steps in a task list
2. Read target/source instructions
3. Analyze target's DNA (structure, style, voice, formatting, conventions, project context)
4. Extract source patterns (organizational, workflow, integration, structure, dependencies)
5. Filter by aspects/percentage
6. Ask clarifying questions if interactive mode was requested
7. Rewrite patterns in target's voice with target's technology names
8. Preview or apply
9. Report summary

## Integration Rules

**Preserve (CRITICAL):** section order, project tech names (`moon`, `npm`, `Terraform`, etc.), paths/directories, command syntax, notation style (arrows/parens), spacing, terminology, all project context

**Extract from source:** organizational patterns only — section grouping, hierarchy, workflow presentation, integration docs, structure styles

**Style matching:** match section presentation, command notation, arrow style, inline details format, heading caps, whitespace, tech vocabulary

**Approach:** extract patterns NOT content → rewrite with target's tech names → insert in existing sections → match formatting exactly → preserve all project-specific elements

**Project preservation:** never replace tech names, keep paths/directories, preserve commands, maintain integration docs, keep dependencies unchanged

**Percentage scale:** 10-30% critical only, 30-50% important (default 45), 50-70% comprehensive, 70-100% extensive

**Aspect filtering:** `workflow` (sequences/delegation), `structure` (directory/hierarchy), `integration` (doc styles), `dependencies` (doc approaches), `commands` (notation styles)

## Implementation

**Discovery:** Accept AGENTS.md paths or directories (e.g. `services/api` → `services/api/AGENTS.md`)

**Analysis:** Parse target structure → analyze formatting → detect voice → extract conventions → identify project context → build template

**Extraction:** Identify organizational patterns → extract workflow presentation → find integration approaches → filter by aspects

**Integration:** Extract patterns only (not content) → rewrite with target's tech names → insert in existing sections → match formatting → validate project-specificity

## Error Handling

File not found, invalid percentage (10-100), no new patterns, aspect not found, incompatible structure, style detection failed

## Safety

Recommend git commit; never modify project tech names/paths/commands; preserve all target content; preview before writing; warn if >30% added; abort if style confidence <85% or project context at risk.

## Gotchas

- **Pattern ≠ content** — if you copy the bullet text verbatim instead of the structural pattern, you've ported the wrong thing
- A target's tech names are sacred — `npm install` doesn't become `gradle build` just because the source uses Gradle
- Style confidence below 85% usually means the target has too much manual customization to safely overwrite — bail out and ask
