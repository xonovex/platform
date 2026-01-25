# cookie-handling: Secure Cookie Configuration and Signed Cookies

**Guideline:** Set secure cookie options explicitly, use signed cookies for sensitive data, and follow browser-enforced prefix requirements for additional security.

**Rationale:** Cookies require careful configuration to ensure HTTPS-only transmission, prevent JavaScript access, defend against CSRF, and verify integrity. Cookie prefixes enforce additional security guarantees, while signed cookies prevent tampering and provide strong integrity verification for sensitive client-side state.

**Example:**

```typescript
import {
  getCookie,
  getSignedCookie,
  setCookie,
  setSignedCookie,
} from "hono/cookie";

// Standard secure cookie
app.post("/login", (c) => {
  setCookie(c, "session", sessionId, {
    secure: true,
    httpOnly: true,
    sameSite: "Strict",
    maxAge: 86400, // 1 day
    path: "/",
  });
  return c.json({success: true});
});

// Signed cookie for integrity (async)
app.post("/preferences", async (c) => {
  await setSignedCookie(c, "prefs", JSON.stringify(prefs), "secret-key", {
    secure: true,
    httpOnly: true,
    sameSite: "Lax",
  });
  return c.json({saved: true});
});

// Verify signed cookie
app.get("/preferences", async (c) => {
  const prefs = await getSignedCookie(c, "secret-key", "prefs");
  if (prefs === false) {
    return c.json({error: "Invalid signature"}, 400);
  }
  return c.json(JSON.parse(prefs ?? "{}"));
});

// Host-prefixed cookie (strictest)
setCookie(c, "__Host-session", token, {
  secure: true,
  path: "/",
  // domain must NOT be set for __Host- prefix
});
```

**Techniques:**
- Import cookie helpers from `hono/cookie` (setCookie, getCookie, setSignedCookie, getSignedCookie)
- Always set `secure: true` and `httpOnly: true` for sensitive cookies
- Use `sameSite: 'Strict'` for maximum CSRF protection or `'Lax'` for better compatibility
- Use `setSignedCookie` (async) for cookies requiring integrity verification
- Verify signed cookies with `getSignedCookie` and check for false value on tampering
- Use `__Host-` prefix for strictest cookie restrictions (requires `secure: true` and `path: '/'`)
- Keep `maxAge` under 400 days to respect browser limits
