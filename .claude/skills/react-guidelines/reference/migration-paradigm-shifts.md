# migration-paradigm-shifts: React 19 Mental Model Changes

**Guideline:** Think server-first, compiler-optimized, declarative; React 19 shifts from client-side to Server Components and from manual to automatic optimization.

**Rationale:** Server Components eliminate useEffect data fetching; Compiler removes manual memoization burden; Form Actions replace useState form fields.

**Example:**
```tsx
// OLD: Client fetch + loading state
function ProductPage({ productId }) {
  const [product, setProduct] = useState(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    fetch(`/api/products/${productId}`).then(data => { setProduct(data); setLoading(false); });
  }, [productId]);
  if (loading) return <Skeleton />;
  return <ProductDetails product={product} />;
}

// NEW: Server Component - direct DB access
async function ProductPage({ productId }) {
  const product = await db.products.find(productId);
  return <ProductDetails product={product} />;
}
```

**Techniques:**
- Server-first: Async Server Components (SC) by default, 'use client' for islands only
- Data fetching: SC replaces useEffect; no loading state needed (Suspense handles it)
- Form handling: useActionState replaces useState; FormData replaces controlled inputs
- Memoization: React Compiler handles it; write clean code, compiler optimizes
- Loading UX: Suspense boundaries for progressive disclosure instead of all-or-nothing
- Paradigm shift: Trust the framework; write simple code, compiler/SC handle complexity
