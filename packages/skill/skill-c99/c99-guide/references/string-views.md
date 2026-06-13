# String views and builders

C strings carry no length, so every `strlen`/`strcmp`/`strtok` rescans for the terminator. Two small types remove the rescans and the hidden-allocation surprises: a non-owning **view** (read) and an owning **builder** over caller memory (write).

## View — non-owning (pointer + length)

```c
typedef struct { const char *data; size_t len; } strview_t;  /* not required to be null-terminated */
```

- Carries its length, so comparison rejects unequal lengths without scanning, and a sub-string is a `{data + off, n}` slice — no copy, no terminator.
- Build it once at the boundary: `strview_from_cstr` does the single `strlen`; a literal view (`{"lit", sizeof("lit") - 1}`) does none.
- Equality is `len` check then `memcmp` — never `strcmp` over borrowed bytes.
- Tokenize with a pure `split(rest, v, sep)` that returns the field and advances `rest` — a stateless, non-mutating replacement for `strtok`.
- Hash the spanned bytes (e.g. FNV-1a) so equal views hash equally regardless of surrounding storage.

A view borrows: it is valid only while the bytes it points at outlive it, and it never frees them.

## Builder — owning, over a caller-provided buffer

```c
typedef struct { char *data; size_t len, capacity; bool truncated; } strbuilder_t;
```

- Caller owns the buffer (no allocation inside): a `_req(max_len)` returns the byte count, the caller sizes storage, `_init(buf, cap)` binds it.
- Every append is **bounded**: what doesn't fit is dropped, the result stays null-terminated, and a `truncated` flag latches so the caller can check once after a run of appends instead of per call.
- `data` is always a valid C string; expose both a `view()` (length-carrying) and a `cstr()` accessor.
- A formatted append wraps `vsnprintf` into the remaining space and latches `truncated` the same way.

This pair replaces the `strcpy`/`strcat`/`sprintf`/`strtok` cluster: reads borrow length-carrying views, writes go through a bounded builder over memory the caller owns.
