# memory-management: Memory Management Patterns

**Guideline:** Prefer stack allocation and ensure proper cleanup of heap allocations with clear ownership tracking.

**Rationale:** Proper memory management prevents leaks, reduces fragmentation, and improves performance. Stack allocation is faster and automatically cleaned up.

**Example:**

```c
// Stack allocation (preferred)
void process_data(void) {
    char buffer[1024];
    int values[100];
    // Automatically freed when function returns
}

// Heap allocation with explicit free
void dynamic_array(size_t size) {
    int *array = malloc(size * sizeof(int));
    if (!array) {
        return;  // Handle allocation failure
    }

    // Use array...

    free(array);  // Always free
}
```

**Techniques:**

- Default stack: Use stack allocation for fixed-size data by default
- Heap when needed: Use malloc/calloc only for dynamic or longer-lived data
- Malloc/free pairing: Always pair every malloc with corresponding free call
- Failure checking: Check allocation return value before dereferencing
- Clear ownership: Document who owns and must free each allocation
