# Sources

## RFC 9457: Problem Details for HTTP APIs

- **URL:** https://www.rfc-editor.org/rfc/rfc9457.html
- **Last reviewed:** 2026-07-22
- **Used for:**
  - `SKILL.md` → Problem Details version and media type
  - `references/error-handling.md`
- **Aspects extracted:**
  - Canonical problem-detail members, extension members, problem-type URIs, and the `application/problem+json` media type

## Hono: Full Documentation (LLM-friendly distillation)

- **URL:** https://hono.dev/llms-full.txt
- **Version:** 4.0.0
- **Content SHA256:** 8df37b8acbf053d00a810f9bfcb1ae9cfddf2d68a984a2a111c6638d15a80d96
- **Last reviewed:** 2026-07-22
- **Used for:**
  - `SKILL.md` → all sections
  - All files under `references/`
- **Aspects extracted:**
  - Application structure / app composition → `references/application-structure.md`
  - Context storage and request-scoped state → `references/context-storage.md`
  - Cookie handling → `references/cookie-handling.md`
  - Error handling patterns → `references/error-handling.md`
  - Middleware combination (`every` / `some` / `except`) → `references/middleware-combine.md`
  - Middleware authoring patterns → `references/middleware-patterns.md`
  - Platform / runtime adapters (Cloudflare Workers, Bun, Deno, Node, etc.) → `references/platform-runtime.md`
  - Security middleware (CORS, CSRF, secure headers, etc.) → `references/security-middleware.md`
  - Validation and end-to-end type safety (Zod, Valibot integrations, `c.req.valid`) → `references/validation-type-safety.md`
  - WebSocket / upgrade support → `references/websocket-support.md`
