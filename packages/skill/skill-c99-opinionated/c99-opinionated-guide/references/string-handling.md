# Strings: length-carrying views, caller-owned builders

**Guideline:** Strings follow the same caller-owns rule as everything else in this style — a non-owning **view** for reads, a bounded **builder** over caller memory for writes — and the libc terminator-scan cluster (`strlen`/`strcmp`/`strcat`/`strtok`) is avoided.

**Rationale:** A C string carries no length, so every `strlen`/`strcmp`/`strcat`/`strtok` rescans to the terminator — an O(n) hidden cost that becomes O(n²) inside a loop. The infamous case: GTA Online spent minutes at load because a JSON parser re-`strlen`'d the whole buffer once per token. Carry the length explicitly and the rescans, and the hidden allocations, both disappear.

## View — non-owning slice (pointer + length)

```c
typedef struct { const char *data; size_t len; } strview_t;   /* not required null-terminated */
```

- Carries its length: equality is a `len` check then `memcmp`, never `strcmp` over borrowed bytes; a substring is a `{data + off, n}` slice — no copy, no terminator.
- Built once at the boundary (`strview_from_cstr` does the single `strlen`); a literal view does none.
- Replace `strtok` with a pure `split(rest, &field, sep)` that returns the field and advances `rest` — stateless, non-mutating, re-entrant.
- A view borrows: it is valid only while the bytes outlive it, and it never frees them.

## Builder — owning, over a caller-provided buffer

```c
typedef struct { char *data; size_t len, capacity; bool truncated; } strbuilder_t;
```

- Caller owns the buffer: `_req(max_len)` returns the byte count, the caller sizes storage, `_init(buf, cap)` binds it — no allocation inside, consistent with [references/caller-owns-memory.md](./caller-owns-memory.md).
- Every append is bounded against `capacity`; overflow is dropped, the result stays null-terminated, and a `truncated` flag latches so the caller checks once after a run of appends, not per call.
- Expose both `view()` (length-carrying) and `cstr()`; a formatted append wraps `vsnprintf` into the remaining space.

This pair replaces the `strcpy`/`strcat`/`sprintf`/`strtok`/`strlen` cluster: reads borrow length-carrying views, writes go through a bounded builder over memory the caller owns. Where a bounded libc call is unavoidable at a boundary, use the `n` variants (`strnlen`, `strncmp`) with an explicit cap and carry the length onward rather than rescanning.

**Related:** [references/caller-owns-memory.md](./caller-owns-memory.md), [references/build-warnings-policy.md](./build-warnings-policy.md) (snprintf truncation)
