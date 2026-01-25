# react-compiler: React Compiler & Automatic Memoization

**Guideline:** Write clean code without manual memoization; let Compiler handle it automatically; use useMemo/useCallback only for effect dependencies.

**Rationale:** Compiler analyzes data flow and applies memoization where beneficial; manual optimization adds overhead and is often incorrect.

**Example:**
```tsx
// Compiler handles this automatically
function ProductList({ items, onSelect }) {
  const sorted = items.filter(x => x.active).sort(byName);
  return sorted.map(item => (
    <li key={item.id} onClick={() => onSelect(item.id)}>{item.name}</li>
  ));
}

// Manual useMemo only for stable effect dependency
function SearchResults({ query }) {
  const searchParams = useMemo(
    () => ({ query, timestamp: Date.now() }),
    [query]
  );
  useEffect(() => { fetchResults(searchParams); }, [searchParams]);
  return <Results />;
}
```

**Techniques:**
- Compiler enabled: Vite react({ babel: { plugins: ['babel-plugin-react-compiler'] } }) or Next.js experimental.reactCompiler
- Rules of React: Pure functions, immutable state, unconditional hooks (compiler enforces)
- Manual memoization: Only for stable effect dependencies; let Compiler handle render optimizations
- Opt-in ('use memo'): Incremental adoption; compile specific components
- Opt-out ('use no memo'): Skip compilation for legacy/problematic code
- Keep it simple: Write readable code; Compiler finds and optimizes bottlenecks
