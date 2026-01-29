# error-handling: Error Handling Patterns

**Guideline:** Return error codes from functions and use output parameters for results to enable proper error checking and recovery.

**Rationale:** Error codes force callers to handle errors explicitly, unlike exceptions. Using return values for errors and output parameters for results creates a consistent pattern throughout the codebase.

**Example:**

```c
typedef enum {
    ERR_OK = 0,
    ERR_NOMEM,
    ERR_INVALID,
    ERR_IO
} error_t;

error_t load_file(const char *path, char **out, size_t *len) {
    FILE *f = fopen(path, "rb");
    if (!f) return ERR_IO;

    fseek(f, 0, SEEK_END);
    long size = ftell(f);
    fseek(f, 0, SEEK_SET);

    char *buf = malloc(size);
    if (!buf) {
        fclose(f);
        return ERR_NOMEM;
    }

    if (fread(buf, 1, size, f) != (size_t)size) {
        free(buf);
        fclose(f);
        return ERR_IO;
    }

    fclose(f);
    *out = buf;
    *len = size;
    return ERR_OK;
}
```

**Techniques:**

- Error enums: Define descriptive error types with zero for success
- Output parameters: Use pointers to return multiple results alongside error codes
- Error checking: Always check return values immediately after function calls
- Cleanup: Properly free resources in all error paths before returning
- Context: Provide specific error codes to indicate failure reason, not generic errors
