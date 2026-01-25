# authentication: JWT Authentication

**Guideline:** Implement JWT-based auth with middleware for authentication and role verification.

**Rationale:** Stateless JWT tokens enable scalable authentication; role-based middleware enforces authorization at route level.

**Example:**

```typescript
declare global { namespace Express {
  interface Request { user?: JwtPayload; }
}}

export function requireAuth(req: Request, res: Response, next: NextFunction): void {
  const token = req.headers.authorization?.replace("Bearer ", "");
  if (!token) return res.status(401).json({error: "Missing token"});
  try {
    req.user = jwt.verify(token, process.env.JWT_SECRET!) as JwtPayload;
    next();
  } catch { res.status(401).json({error: "Invalid token"}); }
}

export function requireRole(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) return res.status(401).json({error: "Not authenticated"});
    if (!roles.includes(req.user.role)) return res.status(403).json({error: "Insufficient"});
    next();
  };
}
```

**Techniques:**
- Extend Express Request: Add optional user property with JwtPayload interface
- requireAuth middleware: Verify Bearer token from Authorization header, attach user to request
- requireRole middleware: Check user role against allowed roles, return 403 if insufficient
- Sign tokens: Include userId, email, role with 7-day expiration
- Bearer scheme: Use "Bearer TOKEN" format in Authorization header
- Hash passwords: Compare with bcrypt instead of storing plaintext
- Status codes: 401 for missing/invalid auth, 403 for insufficient permissions
