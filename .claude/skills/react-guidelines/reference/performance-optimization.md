# performance-optimization: Performance Optimization Techniques

**Guideline:** Use memo() for expensive components; lazy-load routes with React.lazy + Suspense; prefer SSR/SSG; let Compiler handle calculations.

**Rationale:** memo() prevents unnecessary re-renders; code splitting reduces initial bundle; SSR improves FCP; React 19 Compiler removes manual memoization burden.

**Example:**
```tsx
// Memoize expensive component
const ExpensiveList = memo(function({ items, onItemClick }) {
  return (
    <ul>
      {items.map(item => (
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

**Techniques:**
- memo(): Wrap expensive components to skip re-render if props unchanged
- lazy() + Suspense: Code splitting; reduces initial bundle by ~50-70%
- useCallback: Stabilize callbacks passed to memoized children (pair with memo)
- useMemo: Only for effect dependencies or expensive one-time calculations
- React Compiler: In React 19, removes need for manual memoization (vite/next.js config)
- SSR/SSG: Server Components and static generation for better FCP and SEO
