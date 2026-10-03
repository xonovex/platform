# component-design: Component Design Patterns

Build small, composable components; keep state in the lowest common ancestor that needs it; derive computed values instead of storing them in state.

## Example

```tsx
// ✅ Derived state - calculated, not stored
function CartTotal({items}: {items: CartItem[]}) {
  const total = useMemo(
    () => items.reduce((sum, item) => sum + item.price, 0),
    [items],
  );
  return <div>Total: ${total}</div>;
}

// ❌ Duplicated state - drifts out of sync
function CartTotalBad({items}: {items: CartItem[]}) {
  const [total, setTotal] = useState(0);
  useEffect(() => {
    setTotal(items.reduce((sum, item) => sum + item.price, 0));
  }, [items]);
  return <div>Total: ${total}</div>;
}
```

## Techniques

- Break large components into smaller, focused pieces
- Keep state in the lowest common ancestor that needs it
- Derive values with `useMemo` instead of storing them in state
- Extract reusable logic into custom hooks
