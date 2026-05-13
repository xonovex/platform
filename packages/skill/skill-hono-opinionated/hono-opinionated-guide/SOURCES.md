# Sources

## Hono: Full Documentation (LLM-friendly distillation)

- **URL:** https://hono.dev/llms-full.txt
- **Last reviewed:** 2026-05-13
- **Used for:**
  - `SKILL.md` → all sections
  - All files under `references/`
- **Aspects extracted:**
  - Application structure / app composition → `references/application-structure.md`
  - Body size limits → `references/body-limit.md`
  - Context storage and request-scoped state → `references/context-storage.md`
  - Controller patterns (this opinionated variant's split-out approach) → `references/controllers.md`
  - Cookie handling → `references/cookie-handling.md`
  - Error handling patterns → `references/error-handling.md`
  - Middleware combination (`every` / `some` / `except`) → `references/middleware-combine.md`
  - Middleware authoring patterns → `references/middleware-patterns.md`
  - OpenAPI: explicit status codes → `references/openapi-explicit-status-codes.md`
  - OpenAPI: inline handlers (this opinionated variant's preferred style) → `references/openapi-inline-handlers.md`
  - OpenAPI: router hierarchy / nesting → `references/openapi-router-hierarchy.md`
  - OpenAPI: spec generation → `references/openapi-spec-generation.md`
  - Platform / runtime adapters → `references/platform-runtime.md`
  - Router selection (RegExpRouter, SmartRouter, TrieRouter, PatternRouter) → `references/router-selection.md`
  - Security middleware → `references/security-middleware.md`
  - Validation and end-to-end type safety → `references/validation-type-safety.md`
  - WebSocket / upgrade support → `references/websocket-support.md`

## Notes on the "Opinionated" Variant

This skill layers project-specific opinions on top of the upstream Hono docs (OpenAPI-first via inline handlers, explicit status codes, controller-style splits, preferred router choice). When refreshing from `llms-full.txt`, preserve those opinionated decisions — only update upstream-derived content (API surface, available middleware, runtime support).
