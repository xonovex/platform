# String views

C strings carry no length, so every `strlen`/`strcmp`/`strtok` rescans for the terminator — an O(n) hidden cost that becomes O(n²) inside a loop. A non-owning **view** carries the length and removes the rescans.

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

## Writing

Where a write goes through libc, use the bounded `n` variants (`snprintf`, `strnlen`, `strncmp`) with an explicit cap, check the return for truncation, and carry the resulting length onward instead of rescanning. Who owns and sizes the destination buffer is a project design decision, not a property of C99.
