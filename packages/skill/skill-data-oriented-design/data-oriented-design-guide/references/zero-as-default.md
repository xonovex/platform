# zero-as-default: Make All-Zero a Valid Default

## Guideline

Design data so that an all-zero state is a sensible, valid default. Reserve `0` to mean "none / not found / neutral" instead of picking magic sentinels, so zero-initialization produces correct defaults and special cases disappear.

## How to Apply

- Return `0` for "not found" so call sites read `if (handle)`, not a comparison against a magic constant.
- Reserve the first enum entry as `_NONE`/nil, so a default-initialized enum is meaningful.
- Reserve array index `0` as a zero-initialized dummy: index 0 means "no item" yet is still safe to dereference (no validity branch needed).

## Example

```c
// Good: zero is "none"; not-found returns 0; index 0 is a safe nil slot
enum controller_type { CONTROLLER_NONE, CONTROLLER_MOUSE, CONTROLLER_KEYBOARD };
struct bone ulna = { .name = "ulna", .length = 1.0f };   // every other field zeroed
float len = bone_length(find_bone("ulna"));              // find returns 0 -> length 0.0f
bones[0] = (struct bone){0};                             // reserve slot 0 as nil

// Bad: magic sentinel every reader and writer must remember
#define NO_BONE 0xffffffff
if (b != NO_BONE) ...                                    // forget it once -> silent bug
```

## Counter-Example

This only works where the domain has a natural "nothing/neutral" case. Where `0` is a legitimate value (a temperature, a signed offset, an external API that treats `0` as a reserved/invalid handle), don't overload it: use an explicit presence flag or an out-of-domain sentinel instead.

## Related

[existence-based-processing.md](./existence-based-processing.md), [handles-and-indices.md](./handles-and-indices.md)
