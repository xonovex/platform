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
- ZII (Zero Is Initialization): A partial initializer or `= {0}` zero-fills the rest, so a zeroed struct is a valid default — C99's analogue of leaning on C++ RAII; give the all-zero state meaning instead of writing an init function
- Declarative config: Pair a config struct with a `_DEFAULT` const and pass it by value so callers override only what differs — configuration as data, not imperative setter chains (the broader value-oriented API shape lives in [references/value-types.md](./value-types.md))

**Example (ZII default + declarative override):**

```c
typedef struct { int width, height; bool vsync; const char *title; } window_cfg_t;
#define WINDOW_CFG_DEFAULT (window_cfg_t){ .width = 1280, .height = 720, .title = "app" }

window_t w = window_create((window_cfg_t){ .width = 800, .height = 600 });  // vsync=false, title="" via ZII
```
