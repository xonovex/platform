# suspense-streaming: Suspense, Streaming & Error Handling

**Guideline:** Use granular Suspense boundaries for independent progressive loading; nest for hierarchical control; combine with Error Boundaries.

**Rationale:** Suspense enables streaming—content loads incrementally, not all-or-nothing; React 19 pre-warms siblings for faster perceived performance.

**Example:**

```tsx
function Dashboard() {
  return (
    <div>
      <Suspense fallback={<StatsSkeleton />}>
        <StatsSection /> {/* Shows at 150ms */}
      </Suspense>
      <Suspense fallback={<UserSkeleton />}>
        <UserSection /> {/* Shows at 200ms independently */}
      </Suspense>
    </div>
  );
}

// Error + Suspense: graceful failure
<ErrorBoundary fallback={<ErrorMessage />}>
  <Suspense fallback={<Loading />}>
    <DataDisplay />
  </Suspense>
</ErrorBoundary>;
```

**Techniques:**

- Independent boundaries: Each Suspense boundary loads and shows independently (faster perceived time)
- Nested boundaries: Hierarchy for critical→secondary→tertiary with fallbacks at each level
- Error Boundaries: Catch component errors; pair with Suspense for complete error handling
- Root-level errors: React 19 onUncaughtError and onCaughtError callbacks
- Skeleton sizing: Match dimensions to prevent layout shift (CLS optimization)
- Pre-warming: React 19 fetches siblings during suspension for faster transitions
