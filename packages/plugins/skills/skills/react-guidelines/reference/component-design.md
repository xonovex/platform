# component-design: Component Design Patterns

**Guideline:** Build small, composable components with minimal state lifting and derived values.

**Rationale:** Small components are easier to test, reuse, and maintain. Keeping state local reduces complexity and re-renders. Deriving values prevents synchronization bugs.

**Example:**

```tsx
import {memo, useMemo, useState} from "react";

type Item = {id: string; value: number};

// Small, focused component
export const List = memo(function List({items}: {items: Item[]}) {
  const [query, setQuery] = useState("");

  // Derived state - calculated from items and query
  const filteredItems = useMemo(
    () => items.filter((item) => String(item.value).includes(query)),
    [items, query],
  );

  return (
    <div>
      <SearchInput value={query} onChange={setQuery} />
      <ItemList items={filteredItems} />
    </div>
  );
});

// ✅ Good: Derived state
function CartTotal({items}: {items: CartItem[]}) {
  const total = useMemo(
    () => items.reduce((sum, item) => sum + item.price, 0),
    [items],
  );
  return <div>Total: ${total}</div>;
}

// ❌ Bad: Duplicated state
function CartTotalBad({items}: {items: CartItem[]}) {
  const [total, setTotal] = useState(0);

  useEffect(() => {
    setTotal(items.reduce((sum, item) => sum + item.price, 0));
  }, [items]);

  return <div>Total: ${total}</div>;
}
```

**Techniques:**
- Break down large components into smaller, focused pieces
- Keep state in the lowest common ancestor that needs it
- Calculate derived values using `useMemo` instead of storing them in state
- Extract reusable logic into custom hooks
