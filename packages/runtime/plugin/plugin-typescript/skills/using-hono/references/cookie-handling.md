# cookie-handling: Secure Cookie Configuration and Signed Cookies

Set `secure`/`httpOnly`/`sameSite` explicitly. Use `setSignedCookie`/`getSignedCookie` (both async) for integrity; `getSignedCookie` returns `false` on a tampered value. Non-obvious constraints: `__Host-` prefix requires `secure: true` and `path: "/"` and forbids `domain`. Helpers are in `hono/cookie`.

```typescript
import {getSignedCookie, setCookie, setSignedCookie} from "hono/cookie";

setCookie(c, "session", sessionId, {
  secure: true,
  httpOnly: true,
  sameSite: "Strict",
  maxAge: 86400,
  path: "/",
});

await setSignedCookie(c, "prefs", JSON.stringify(prefs), "secret-key", {
  secure: true,
  httpOnly: true,
});
const prefs = await getSignedCookie(c, "secret-key", "prefs");
if (prefs === false) return c.json({error: "Invalid signature"}, 400);

// __Host- prefix: no domain allowed
setCookie(c, "__Host-session", token, {secure: true, path: "/"});
```
