# memory-management: Memory Management

Default to stack allocation for fixed-size, function-scoped data; reach for `malloc`/`calloc` only for dynamic or longer-lived data. Pair every allocation with exactly one `free`, check the return before dereferencing, and free on every error path. Document who owns each allocation.

```c
int *array = malloc(size * sizeof *array);
if (!array) return ERR_NOMEM;   // check before use
/* ... */
free(array);
```

For arenas, pools, and ownership models beyond plain malloc/free, see **memory-management-guide**.
