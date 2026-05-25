# Sources

## Hono: Full Documentation (LLM-friendly distillation)

- **URL:** https://hono.dev/llms-full.txt
- **Last reviewed:** 2026-05-13
- **Used for:**
  - `SKILL.md` → all sections
  - All files under `references/` (the opinionated overlay only)
- **Aspects extracted:**
  - Body size limits → `references/body-limit.md`
  - Controller patterns (this opinionated variant's split-out approach) → `references/controllers.md`
  - OpenAPI: explicit status codes → `references/openapi-explicit-status-codes.md`
  - OpenAPI: inline handlers (this opinionated variant's preferred style) → `references/openapi-inline-handlers.md`
  - OpenAPI: router hierarchy / nesting → `references/openapi-router-hierarchy.md`
  - OpenAPI: spec generation → `references/openapi-spec-generation.md`
  - Router selection (RegExpRouter, SmartRouter, TrieRouter, PatternRouter) → `references/router-selection.md`

## Notes on the "Opinionated" Variant

This skill is an overlay on `hono-guide` and covers only the opinionated decisions (OpenAPI-first via inline handlers, explicit status codes, controller-style splits, preferred router choice, body limits). Generic Hono concepts — app structure, validation, middleware, errors, cookies, security, WebSockets, context storage, runtimes — are owned by `hono-guide`, and their provenance lives in that skill's `SOURCES.md`. When refreshing from `llms-full.txt`, preserve the opinionated decisions here — only update upstream-derived content (API surface, available middleware, runtime support).
