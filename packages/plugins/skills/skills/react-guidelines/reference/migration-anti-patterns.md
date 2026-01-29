# migration-anti-patterns: React 19 Anti-Patterns

**Guideline:** Replace useEffect data fetching with Server Components; use Form Actions instead of useState for forms; let Compiler handle memoization.

**Rationale:** React 19 provides cleaner solutions; old patterns create unnecessary complexity and performance overhead.

**Example:**

```tsx
// OLD: useEffect + loading state
function UserProfile({userId}) {
  const [user, setUser] = useState(null),
    [loading, setLoading] = useState(true);
  useEffect(() => {
    fetchUser(userId)
      .then(setUser)
      .finally(() => setLoading(false));
  }, [userId]);
  if (loading) return <Spinner />;
  return <Profile user={user} />;
}

// NEW: Server Component
async function UserProfile({userId}) {
  const user = await fetchUser(userId);
  return <Profile user={user} />;
}
```

**Techniques:**

- Remove useEffect data fetching → Use Server Components (async/await, direct DB)
- Remove manual loading state → Use useActionState and Suspense
- Remove useState form fields → Use FormData + Server Actions
- Remove excessive useMemo/useCallback → Let React Compiler optimize
- Remove forwardRef wrapper → Pass ref as regular prop in React 19
- Remove 'use client' everywhere → Server Components by default; client islands only
- Remove <Suspense/> for errors only → Combine with Error Boundaries for full coverage
