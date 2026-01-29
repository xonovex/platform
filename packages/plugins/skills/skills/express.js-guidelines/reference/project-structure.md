# project-structure: Application Structure

**Guideline:** Organize Express apps with separated concerns - routes, controllers, middleware, schemas in dedicated directories.

**Rationale:** Clear separation prevents tight coupling and improves maintainability in larger codebases.

**Example:**

```
src/
├── app.ts
├── server.ts
├── routes/          # Route definitions with middleware chains
├── controllers/     # Business logic handlers
├── middleware/      # Auth, validation, error handling
├── schemas/         # Zod schemas for validation
├── services/        # Database/external service logic
├── types/           # TypeScript type extensions
└── utils/           # Helper functions
```

**Techniques:**

- routes/: Define route handlers with middleware chaining, avoid business logic
- controllers/: Implement typed handlers with try-catch, call services
- middleware/: Auth, validation, CORS, logging, error handling middleware
- schemas/: Zod schemas for type-safe request/response validation
- services/: Database queries, external APIs, business logic
- types/: Extend Express Request/Response interfaces globally
- Separation: Keep concerns isolated, each file has single responsibility
