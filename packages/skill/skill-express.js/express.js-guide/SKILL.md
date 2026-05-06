---
name: express.js-guide
description: "Use when editing or scaffolding Express 5+ API servers in TypeScript. Triggers on `.ts` files with `express` or `@types/express` imports, route definitions, middleware, controllers, and prompts about routes, error handling, JWT, CORS, Morgan, or Zod validation in Express, even when the user doesn't say 'Express'. Skip Hono (use hono-guide / hono-opinionated-guide), Fastify, and non-HTTP work."
---

# Express.js Coding Guidelines

## Requirements

- Express ≥ 5, TypeScript ≥ 5.8, Helmet/CORS/Morgan, Zod.

## Essentials

- **Project structure** - Routes, controllers, middleware as separate modules, see [reference/project-structure.md](reference/project-structure.md)
- **Input validation** - Validate params/body/query with Zod at route edges, see [reference/validation.md](reference/validation.md)
- **Error handling** - Central error handler, never leak stack traces in prod, see [reference/error-handling.md](reference/error-handling.md)
- **Authentication** - JWT auth with role/permission middleware, see [reference/authentication.md](reference/authentication.md)
- **Response format** - Consistent JSON shape and status codes, see [reference/responses.md](reference/responses.md)
- **Testing** - Unit-test controllers/middleware, integration-test routes, see [reference/testing.md](reference/testing.md)

## Progressive disclosure

- Read [reference/project-structure.md](reference/project-structure.md) - When organizing a new Express project
- Read [reference/routes.md](reference/routes.md) - When defining REST endpoints or route patterns
- Read [reference/controllers.md](reference/controllers.md) - When implementing request handlers
- Read [reference/validation.md](reference/validation.md) - When adding input validation to routes
- Read [reference/authentication.md](reference/authentication.md) - When implementing JWT auth or session management
- Read [reference/error-handling.md](reference/error-handling.md) - When centralizing error responses
- Read [reference/responses.md](reference/responses.md) - When standardizing API response formats
- Read [reference/app-setup.md](reference/app-setup.md) - When configuring Express app initialization
- Read [reference/testing.md](reference/testing.md) - When writing unit or integration tests
