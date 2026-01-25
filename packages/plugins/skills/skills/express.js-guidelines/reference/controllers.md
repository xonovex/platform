# controllers: Controller Pattern

**Guideline:** Implement async controllers with typed request/response, wrap in try-catch, pass errors to next().

**Rationale:** Type-safe handlers with proper error handling ensure predictable API behavior.

**Example:**

```typescript
export async function list(
  req: Request<{}, {}, {}, ListUsersQuery>,
  res: Response,
  next: NextFunction,
): Promise<void> {
  try {
    const {page, limit} = req.query;
    const result = await userService.list({page, limit});
    res.json({data: result.users, pagination: {page, limit, total: result.total}});
  } catch (error) {
    next(error);
  }
}

export async function getById(
  req: Request<UserParams>,
  res: Response,
  next: NextFunction,
): Promise<void> {
  try {
    const user = await userService.getById(req.params.id);
    if (!user) {
      res.status(404).json({error: "User not found"});
      return;
    }
    res.json({data: user});
  } catch (error) {
    next(error);
  }
}
```

**Techniques:**
- Type generics: Use Request<Params, ResBody, ReqBody, Query> for type safety
- Return Promise<void>: Indicates async function that doesn't return a value
- Try-catch: Wrap all async logic and pass errors to next(error)
- Validate first: Check for missing resources (404) before returning data
- Service layer: Call business logic via injected service dependencies
- Error handler: Pass errors to next() for centralized error handling
- Standard responses: Return {data: value} or {error: message} structures
