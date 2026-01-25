# compound-literals: Compound Literals

**Guideline:** Use compound literals to create temporary values inline without declaring separate variables.

**Rationale:** Compound literals reduce boilerplate code and improve readability by creating temporary structs or arrays directly in expressions.

**Example:**

```c
struct Point {
    int x, y;
};

int distance(struct Point a, struct Point b);

void example(void) {
    // Create temporary Point inline
    int d = distance(
        (struct Point){.x = 0, .y = 0},
        (struct Point){.x = 3, .y = 4}
    );

    // Initialize array inline
    int arr[] = {(int[]){1, 2, 3, 4, 5}};
}
```

**Techniques:**
- Inline syntax: Use `(Type){initializer-list}` to create temporaries inline
- With designators: Combine with designated initializers for clarity
- Function arguments: Create temporary structs directly in function calls
- Aggregate types: Works with any struct or array type
- Automatic cleanup: Lifetime extends to end of enclosing scope automatically
