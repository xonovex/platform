# performance-optimization: Performance Optimization Techniques

Use `memo()` for expensive components, `lazy()` + Suspense for route code-splitting, and Server Components/SSG for FCP and SEO.

## Example

```tsx
// Memoize expensive component (skips re-render if props unchanged)
const ExpensiveList = memo(function ({items, onItemClick}) {
  return (
    <ul>
      {items.map((item) => (
        <li key={item.id} onClick={() => onItemClick(item.id)}>
          {item.name}
        </li>
      ))}
    </ul>
  );
});

// Lazy load routes
const Dashboard = lazy(() => import("./pages/Dashboard"));
function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </Suspense>
  );
}
```

## Techniques

- `memo()`: wrap expensive components to skip re-render on unchanged props
- `lazy()` + Suspense: code splitting; cuts initial bundle
- `useCallback`: stabilize callbacks passed to memoized children (pair with memo)
- `useMemo`: memoize expensive one-time calculations
- SSR/SSG: Server Components and static generation for FCP and SEO
