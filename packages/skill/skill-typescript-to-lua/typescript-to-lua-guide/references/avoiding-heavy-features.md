# avoiding-heavy-features: Skip Runtime-Heavy TS

Deep inheritance, `async`/`await`, and heavy OOP pull in verbose TSTL runtime support. Prefer composition and namespaces over class hierarchies (see namespaces-vs-classes.md); replace `async`/`await` with `function*` generators, which compile to Lua coroutines (see coroutine-patterns.md). Model data as plain objects, not methods on classes.

```typescript
// AVOID
async function loadData() {
  return await fetch();
}

// PREFER: generator → coroutine
function* loadData() {
  const data = yield fetchAsync();
  return data;
}
```
