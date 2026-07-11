# const-correctness: Const Correctness

Mark read-only pointer parameters and immutable data `const`. Read the qualifier by position — the `const` binds to whatever is on its left:

```c
const int *p1;        // pointer to const int (data immutable, pointer rebindable)
int *const p2;        // const pointer to int (pointer fixed, data mutable)
const int *const p3;  // both fixed
```
