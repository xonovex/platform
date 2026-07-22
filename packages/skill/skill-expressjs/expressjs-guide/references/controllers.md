# controllers: Controller Pattern

Type handlers with `Request<Params, ResBody, ReqBody, Query>` (note the generic order) and return `Promise<void>`. Express 5 forwards thrown errors and rejected promises to the central error handler, so leave routine async failures uncaught. Catch only when the controller can recover locally, must add context, or enters callback/timer work outside the returned Promise. Check for missing resources (404) before returning data. Call business logic through an injected service, not inline.

```typescript
export async function list(
  req: Request<{}, {}, {}, ListUsersQuery>,
  res: Response,
): Promise<void> {
  const {page, limit} = req.query;
  const result = await userService.list({page, limit});
  res.json({
    data: result.users,
    pagination: {page, limit, total: result.total},
  });
}

export async function getById(
  req: Request<UserParams>,
  res: Response,
): Promise<void> {
  const user = await userService.getById(req.params.id);
  if (!user) {
    res.status(404).json({error: "User not found"});
    return;
  }
  res.json({data: user});
}
```
