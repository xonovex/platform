# project-structure: Application Structure

Separate concerns into dedicated directories; routes stay free of business logic, controllers call services, `types/` holds the global `Express.Request`/`Response` augmentations.

```
src/
├── app.ts
├── server.ts
├── routes/          # route + middleware chains, no business logic
├── controllers/     # typed handlers with try-catch, call services
├── middleware/      # auth, validation, error handling
├── schemas/         # Zod schemas
├── services/        # database / external APIs / business logic
├── types/           # Express Request/Response augmentations
└── utils/
```
