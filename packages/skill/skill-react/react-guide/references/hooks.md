# hooks: Custom Hooks & Rules of Hooks

Call hooks only at top-level (not in loops/conditions), only from functions or custom hooks. Extract reusable logic into custom hooks with a `use` prefix.

## Example

```tsx
// useApi - fetch with loading/error states
function useApi<T>(url: string) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(url)
      .then((r) => r.json())
      .then((json) => {
        if (!cancelled) {
          setData(json);
          setError(null);
        }
      })
      .catch((err) => {
        if (!cancelled) setError(err);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [url]);

  return {data, loading, error};
}
```

## Techniques

- useApi: fetch with cancellation flag; handle loading/error/data states; type-generic
- useLocalStorage: parse JSON on read; stringify on write; handle errors
- usePrevious: store value in `useRef`; update in `useEffect`; return previous on next render
- useDebounce: delay state updates with `setTimeout`; cleanup on unmount
- Error handling: type-check errors (`instanceof Error`); cleanup in finally blocks
- Dependencies: include full dependency array; avoid stale closures
