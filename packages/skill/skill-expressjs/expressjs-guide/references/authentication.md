# authentication: JWT Authentication

Verify the `Bearer` token in `requireAuth`, attach the payload to `req.user` (declaration-merged onto `Express.Request`), and gate roles with a `requireRole(...roles)` factory that runs after `requireAuth`. Return 401 for missing/invalid auth, 403 for insufficient role. Sign tokens over `{userId, email, role}` with an expiration (`expiresIn: "7d"`); compare passwords with bcrypt, never plaintext.

```typescript
declare global {
  namespace Express {
    interface Request {
      user?: JwtPayload;
    }
  }
}

export function requireAuth(
  req: Request,
  res: Response,
  next: NextFunction,
): void {
  const token = req.headers.authorization?.replace("Bearer ", "");
  if (!token) return res.status(401).json({error: "Missing token"});
  try {
    req.user = jwt.verify(token, process.env.JWT_SECRET!) as JwtPayload;
    next();
  } catch {
    res.status(401).json({error: "Invalid token"});
  }
}

export function requireRole(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) return res.status(401).json({error: "Not authenticated"});
    if (!roles.includes(req.user.role))
      return res.status(403).json({error: "Insufficient"});
    next();
  };
}
```
