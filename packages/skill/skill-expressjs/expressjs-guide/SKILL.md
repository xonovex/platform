---
name: expressjs-guide
description: "Use when editing or scaffolding Express 5+ API servers in TypeScript. Triggers on `.ts` files with `express` or `@types/express` imports, route definitions, middleware, controllers, and prompts about routes, error handling, JWT, CORS, Morgan, or Zod validation in Express, even when the user doesn't say 'Express'. Skip Hono (use hono-guide / hono-opinionated-guide), Fastify, and non-HTTP work."
---

# Express.js Coding Guidelines

## Requirements

- Express ≥ 5, TypeScript ≥ 5.8, Helmet/CORS/Morgan, Zod.

## Essentials

- **Project structure** - Routes, controllers, middleware as separate modules, see [references/project-structure.md](references/project-structure.md)
- **Input validation** - Validate params/body/query with Zod at route edges, see [references/validation.md](references/validation.md)
- **Error handling** - Central error handler, never leak stack traces in prod, see [references/error-handling.md](references/error-handling.md)
- **Authentication** - JWT auth with role/permission middleware, see [references/authentication.md](references/authentication.md)
- **Response format** - Consistent JSON shape and status codes, see [references/responses.md](references/responses.md)
- **Testing** - Unit-test controllers/middleware, integration-test routes, see [references/testing.md](references/testing.md)

## Gotchas

- Express 4 swallows unhandled async errors silently — wrap async handlers or upgrade to Express 5 (which forwards to error middleware)
- Error-handling middleware needs **four** parameters `(err, req, res, next)` — three-arg middleware is a regular handler, not an error one
- Middleware order is execution order — auth before route, error handler last; misordering creates security holes or silent skips
- `res.json()` ends the response — calling it twice (e.g. after `next()`) throws `Cannot set headers after they are sent`

## Progressive disclosure

- Read [references/project-structure.md](references/project-structure.md) - When organizing a new Express project
- Read [references/routes.md](references/routes.md) - When defining REST endpoints or route patterns
- Read [references/controllers.md](references/controllers.md) - When implementing request handlers
- Read [references/validation.md](references/validation.md) - When adding input validation to routes
- Read [references/authentication.md](references/authentication.md) - When implementing JWT auth or session management
- Read [references/error-handling.md](references/error-handling.md) - When centralizing error responses
- Read [references/responses.md](references/responses.md) - When standardizing API response formats
- Read [references/app-setup.md](references/app-setup.md) - When configuring Express app initialization
- Read [references/testing.md](references/testing.md) - When writing unit or integration tests
