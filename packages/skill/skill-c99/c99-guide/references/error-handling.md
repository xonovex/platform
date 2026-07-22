# error-handling: Error Handling Patterns

Return a descriptive error enum (0 = success) and pass results through out-params. Check every fallible call immediately, and free every already-acquired resource on each error path before returning.

```c
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

typedef enum { ERR_OK = 0, ERR_NOMEM, ERR_INVALID, ERR_IO } error_t;

error_t load_file(const char *path, char **out, size_t *len) {
    if (!path || !out || !len) return ERR_INVALID;
    *out = NULL;
    *len = 0;

    FILE *f = fopen(path, "rb");
    if (!f) return ERR_IO;

    if (fseek(f, 0, SEEK_END) != 0) { fclose(f); return ERR_IO; }
    long end = ftell(f);
    if (end < 0 || (uintmax_t)end > SIZE_MAX) { fclose(f); return ERR_IO; }
    if (fseek(f, 0, SEEK_SET) != 0) { fclose(f); return ERR_IO; }

    size_t size = (size_t)end;
    char *buf = malloc(size == 0 ? 1 : size);
    if (!buf) { fclose(f); return ERR_NOMEM; }
    if (fread(buf, 1, size, f) != size) {
        free(buf);
        fclose(f);
        return ERR_IO;
    }
    if (fclose(f) != 0) { free(buf); return ERR_IO; }

    *out = buf;
    *len = size;
    return ERR_OK;
}
```

For a single small result, return it (or a `{bool ok; T value;}` struct) by value instead of an out-param — reserve out-params for multiple or large results, see [references/value-types.md](./value-types.md).
