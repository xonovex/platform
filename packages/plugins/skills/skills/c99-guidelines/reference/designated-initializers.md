# designated-initializers: Designated Initializers

**Guideline:** Use designated initializers for struct and array initialization to make code self-documenting and reduce errors.

**Rationale:** Designated initializers explicitly name fields being initialized, making code more readable and maintainable.

**Example:**

```c
struct Point {
    int x, y, z;
};

// Named field initialization
struct Point p = {.x = 1, .y = 2, .z = 3};

// Partial initialization (rest zeroed)
struct Point origin = {.x = 0, .y = 0};

// Array designated initializers
int days[12] = {
    [0] = 31, [1] = 28, [2] = 31,
    [3] = 30, [4] = 31, [5] = 30
};
```

**Techniques:**
- Field syntax: Use `.field = value` for struct initialization with clarity
- Array syntax: Use `[index] = value` for sparse array initialization
- Partial init: Unspecified fields are automatically zeroed without explicit values
- Order-independent: Fields can be specified in any order within initializer
- Large structs: Especially useful for structs with many fields or optional fields
