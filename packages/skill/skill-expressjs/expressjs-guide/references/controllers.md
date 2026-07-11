# controllers: Controller Pattern

Type handlers with `Request<Params, ResBody, ReqBody, Query>` (note the generic order), return `Promise<void>`, wrap the body in try-catch, and forward errors via `next(error)` to the central handler. Check for missing resources (404) before returning data. Call business logic through an injected service, not inline.

```typescript
export async function list(
  req: Request<{}, {}, {}, ListUsersQuery>,
  res: Response,
  next: NextFunction,
): Promise<void> {
  try {
    const {page, limit} = req.query;
    const result = await userService.list({page, limit});
    res.json({
      data: result.users,
      pagination: {page, limit, total: result.total},
    });
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
