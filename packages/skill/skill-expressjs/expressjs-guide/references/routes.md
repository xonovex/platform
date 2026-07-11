# routes: Routes with Validation and Auth

Chain middleware in fixed order per route: `requireAuth` → `requireRole` → `validateParams`/`validateQuery`/`validateBody` → controller. Import the schema types so controller `Request` generics stay typed.

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
