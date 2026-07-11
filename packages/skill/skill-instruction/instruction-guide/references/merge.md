# merge: Port Organizational Patterns Between Projects

Extracts organizational patterns from a source AGENTS.md and integrates them into a target file while strictly preserving the target's structure, style, and project-specific context.

## Core Workflow

Read target + source → analyze target's DNA (structure, style, voice, formatting, conventions, project context) → extract source's organizational patterns → filter by aspect/percentage → (ask clarifying questions if interactive) → rewrite in target's voice with target's tech names → preview or apply → report.

Discovery accepts AGENTS.md paths or directories (`services/api` → `services/api/AGENTS.md`).

## Integration Rules

### Preserve (CRITICAL)

section order, project tech names (`moon`, `npm`, `Terraform`, etc.), paths/directories, command syntax, integration docs, dependencies, notation style (arrows/parens), spacing, terminology — all project context

### Extract from source

organizational patterns only (section grouping, hierarchy, workflow presentation, integration docs, structure styles) — never the content: rewrite with target's tech names → insert in existing sections → match formatting exactly

### Style matching

match section presentation, command notation, arrow style, inline details format, heading caps, whitespace, tech vocabulary

### Percentage scale

10-30% critical only, 30-50% important (default 45), 50-70% comprehensive, 70-100% extensive (valid 10-100)

### Aspect filtering

`workflow` (sequences/delegation), `structure` (directory/hierarchy), `integration` (doc styles), `dependencies` (doc approaches), `commands` (notation styles)

### Thresholds

warn if >30% content added; abort if style confidence <85% or project context at risk

## Gotchas

- **Pattern ≠ content** — if you copy the bullet text verbatim instead of the structural pattern, you've ported the wrong thing
- A target's tech names are sacred — `npm install` doesn't become `gradle build` just because the source uses Gradle
- Style confidence below 85% usually means the target has too much manual customization to safely overwrite — bail out and ask
