# const-correctness: Const Correctness

**Guideline:** Use const qualifier to mark immutable data, function parameters, and return values to prevent unintended modifications.

**Rationale:** Const correctness catches bugs at compile time, documents intent, and helps maintain API contracts.

**Example:**

```c
// Use const for immutable data
void process_buffer(const char *input, size_t len) {
    // input cannot be modified
}

// Const pointers
const int *p1;        // pointer to const int
int *const p2;        // const pointer to int
const int *const p3;  // const pointer to const int

// Const in function return
const char *get_name(void) {
    static const char name[] = "example";
    return name;
}
```

**Techniques:**
- Read-only parameters: Mark function parameters with const when not modified
- Pointer variants: Distinguish `const T *p` (pointer to const) vs `T *const p` (const pointer)
- Return values: Use const for immutable return values to prevent modification
- Intent documentation: Const clearly indicates what can/cannot be modified
- Compile-time safety: Catches unintended modifications at compile time
