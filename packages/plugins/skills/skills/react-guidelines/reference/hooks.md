# filename: hooks

**Guideline:** Follow Rules of Hooks (top-level only, from functions/custom hooks); extract reusable logic into custom hooks with `use` prefix.

**Rationale:** Hook rules ensure predictable behavior; custom hooks enable logic reuse without prop drilling; focused hooks maintain single responsibility.

**Example:**
```tsx
// useApi - fetch with loading/error states
function useApi<T>(url: string) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(url).then(r => r.json()).then(json => {
      if (!cancelled) {setData(json); setError(null);}
    }).catch(err => {
      if (!cancelled) setError(err);
    }).finally(() => {
      if (!cancelled) setLoading(false);
    });
    return () => {cancelled = true;};
  }, [url]);

  return {data, loading, error};
}
```

**Techniques:**
- Rules of Hooks: Call only at top-level (not loops/conditions); only from functions/custom hooks
- Custom hooks: Extract reusable logic; name with `use` prefix; single responsibility
- useApi: Fetch with cancellation flag; handle loading/error/data states; type-generic
- useLocalStorage: Parse JSON on read; stringify on write; handle errors gracefully
- usePrevious: Store value in `useRef`; update in `useEffect`; return previous on next render
- useDebounce: Delay state updates with `setTimeout`; cleanup on unmount
- Error handling: Type check errors (`instanceof Error`); cleanup in finally blocks
- Dependencies: Always include full dependency array; avoid stale closures
