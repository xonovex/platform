# compound-literals: Compound Literals

Create temporary structs/arrays inline with `(Type){initializer-list}` instead of a named variable; combine with designated initializers. Lifetime extends to the end of the enclosing scope (block scope for automatic ones) — don't return a pointer into one.

```c
int d = distance((struct Point){.x = 0, .y = 0}, (struct Point){.x = 3, .y = 4});
```
