# react-compiler: React Compiler & Automatic Memoization

Write clean code without manual memoization; the Compiler applies it where beneficial. Use `useMemo`/`useCallback` only for stable effect dependencies.

## Example

```tsx
// Compiler memoizes this automatically
function ProductList({items, onSelect}) {
  const sorted = items.filter((x) => x.active).sort(byName);
  return sorted.map((item) => (
    <li key={item.id} onClick={() => onSelect(item.id)}>
      {item.name}
    </li>
  ));
}

// Manual useMemo only for a stable effect dependency
function SearchResults({query}) {
  const searchParams = useMemo(() => ({query, timestamp: Date.now()}), [query]);
  useEffect(() => {
    fetchResults(searchParams);
  }, [searchParams]);
  return <Results />;
}
```

## Techniques

- Enable: Vite `react({ babel: { plugins: ['babel-plugin-react-compiler'] } })` or Next.js `experimental.reactCompiler`
- Requires Rules of React: pure functions, immutable state, unconditional hooks (compiler enforces)
- Opt-in per component: `'use memo'`; opt-out: `'use no memo'`
