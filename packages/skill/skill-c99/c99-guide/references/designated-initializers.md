# designated-initializers: Designated Initializers

Initialize with `.field = value` (structs) and `[index] = value` (sparse arrays). Fields may appear in any order; unmentioned members are zero-filled — so a missing field is a silent zero, not an error.

```c
struct Point p = {.x = 1, .y = 2, .z = 3};
int days[12] = {[0] = 31, [1] = 28, [2] = 31, [3] = 30};  // rest zeroed
```

## ZII (Zero Is Initialization)

A partial initializer or `= {0}` zero-fills the rest, so the all-zero state is a valid default — C99's analogue of leaning on C++ RAII. Give the zeroed struct meaning instead of writing an init function. Pair a config struct with a `_DEFAULT` const and pass it by value so callers override only what differs (the broader value-oriented API shape lives in [references/value-types.md](./value-types.md)):

```c
typedef struct { int width, height; bool vsync; const char *title; } window_cfg_t;
#define WINDOW_CFG_DEFAULT (window_cfg_t){ .width = 1280, .height = 720, .title = "app" }

window_t w = window_create((window_cfg_t){ .width = 800, .height = 600 });  // vsync=false, title="" via ZII
```
