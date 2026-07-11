# error-handling: Error Handling Patterns

Return a descriptive error enum (0 = success) and pass results through out-params. Check every fallible call immediately, and free every already-acquired resource on each error path before returning.

```c
typedef enum { ERR_OK = 0, ERR_NOMEM, ERR_INVALID, ERR_IO } error_t;

error_t load_file(const char *path, char **out, size_t *len) {
    FILE *f = fopen(path, "rb");
    if (!f) return ERR_IO;
    /* ... size the file ... */
    char *buf = malloc(size);
    if (!buf)                              { fclose(f); return ERR_NOMEM; }
    if (fread(buf, 1, size, f) != (size_t)size) { free(buf); fclose(f); return ERR_IO; }
    fclose(f);
    *out = buf; *len = size;
    return ERR_OK;
}
```

For a single small result, return it (or a `{bool ok; T value;}` struct) by value instead of an out-param — reserve out-params for multiple or large results, see [references/value-types.md](./value-types.md).
