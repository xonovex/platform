# routes: Routes with Validation and Auth

**Guideline:** Chain validation and auth middleware before controller handlers in route definitions.

**Rationale:** Declarative middleware chaining ensures validation/auth happens before business logic.

**Example:**

```typescript
const router = express.Router();

router.get(
  "/",
  requireAuth,
  validateQuery(ListUsersQuerySchema),
  usersController.list,
);
router.get(
  "/:id",
  requireAuth,
  validateParams(UserParamsSchema),
  usersController.getById,
);
router.post("/", validateBody(CreateUserSchema), usersController.create);
router.patch(
  "/:id",
  requireAuth,
  validateParams(UserParamsSchema),
  validateBody(UpdateUserSchema),
  usersController.update,
);
router.delete(
  "/:id",
  requireAuth,
  requireRole("admin"),
  validateParams(UserParamsSchema),
  usersController.remove,
);

export {router as userRoutes};
```

**Techniques:**

- Middleware order: Auth (requireAuth) before validation before controller
- requireAuth: Verify JWT token and attach user to request
- requireRole: Check user role after requireAuth
- validateBody: Validate req.body against Zod schema
- validateParams: Validate req.params against Zod schema
- validateQuery: Validate req.query against Zod schema
- Controller last: Place controller handler after all middleware
- Type safety: Import schema types for typed Request generics in controllers
