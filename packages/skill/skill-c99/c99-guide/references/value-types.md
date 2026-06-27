# Value-oriented APIs and result types

Modern C passes and returns small plain-old-data structs _by value_, and bundles a result with its validity in one returned struct, instead of threading everything through out-parameters. Fewer pointers means fewer aliasing hazards for the optimizer and fewer lifetime questions for the caller.

## Return small results by value

```c
/* Out-param style: caller must declare, pass &, and remember to check. */
void vec3_add(const vec3_t *a, const vec3_t *b, vec3_t *out);

/* Value style: composes in expressions, no aliasing, no lifetime question. */
vec3_t vec3_add(vec3_t a, vec3_t b);
vec3_t r = vec3_add(vec3_add(a, b), c);   /* reads like the math it is */
```

- A struct returned by value lets the compiler keep it in registers and assume no aliasing between the arguments and the result. An out-pointer forces it to assume `out` may alias the inputs and reload across the call.
- Pass small structs (vectors, handles, small configs) by value too; take `const T *` only when the struct is large or you genuinely need in-place mutation.

## Bundle value + validity in a result struct

For a single fallible computation, return the value and whether it is valid together — an `optional`-like value with no heap and no out-param:

```c
typedef struct { bool ok; uint32_t value; } u32_result_t;

u32_result_t parse_u32(strview_t text) {
    /* ... */
    if (bad) return (u32_result_t){ .ok = false };
    return (u32_result_t){ .ok = true, .value = n };
}

u32_result_t r = parse_u32(field);
if (!r.ok) return ERR_INVALID;
use(r.value);
```

- These compose: a chain stops at the first `!ok` without a pyramid of out-param checks.
- It is value-oriented and allocation-free — the result lives on the stack and is copied by value.

## When out-parameters are still right

Value returns do not replace every out-param:

- **Multiple heterogeneous results** — return a named result struct, or keep out-params if the callee is the natural owner of the writes.
- **Large results / caller-owned storage** — when the result is big or the caller must own the buffer, write through a caller-provided pointer (the caller-owns-memory style). Returning a 4 KB struct by value just to copy it is the wrong trade.
- **Status code + payload** — return the status enum and fill the payload through a checked out-param; this is the dominant pattern once a function can fail several distinct ways, see [references/error-handling.md](./error-handling.md).

The bias: a single small result → return it (or a result struct) by value; a fallible op with one cheap value → result struct; anything heavier → status code + out-param.

### Related

[references/error-handling.md](./error-handling.md), [references/designated-initializers.md](./designated-initializers.md)
