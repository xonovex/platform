# migration-anti-patterns: React 19 Anti-Patterns & Mental Model

Think server-first and compiler-optimized: replace `useEffect` data fetching with async Server Components, manual loading state with Suspense/`useActionState`, controlled-input forms with `FormData` + Form Actions, and manual `useMemo`/`useCallback` with the React Compiler.

## Example

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

// NEW: async Server Component - direct DB access, no loading state
async function UserProfile({userId}) {
  const user = await fetchUser(userId);
  return <Profile user={user} />;
}
```

## Techniques

- useEffect data fetching → async Server Components (direct DB/fs); Suspense handles loading
- Manual loading state → `useActionState` + Suspense boundaries (progressive, not all-or-nothing)
- useState form fields → `FormData` + Server Actions
- Excessive useMemo/useCallback → let React Compiler optimize; write clean code
- forwardRef wrapper → `ref` as regular prop
- `'use client'` everywhere → Server Components by default; client islands only
- `<Suspense>` alone → combine with Error Boundaries for full coverage
