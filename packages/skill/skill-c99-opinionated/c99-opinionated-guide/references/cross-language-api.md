# cross-language-api: Designing a C API for Cross-Language Binding

## Guideline

When an API must be callable from other languages, expose a plain-C interface restricted to a portable subset (no untagged unions, variadics, or globals), prefer passing flat data over pointers, scope every borrowed pointer to the call itself, and describe the API in a machine-readable spec so idiomatic bindings can be generated rather than hand-written.

## Rationale

The C ABI is the one calling convention every platform and nearly every language's FFI understands; C++ has no stable ABI, so a C surface is the lingua franca for inter-language calls. But raw FFI is awkward — it needs unsafe blocks, has no string marshaling, no memory-management integration, and no native idioms — so "just expose C" is necessary but not sufficient for a _nice_ binding. Two moves fix that. First, keep the API shape inside what FFIs can express: untagged unions don't exist in most languages, variadics aren't universally supported, and global structures are discouraged everywhere outside C/C++. Second, drive bindings from a spec (IDL/XML/JSON) instead of by hand, so a generator can emit idiomatic, always-in-sync wrappers for each target language and enforce the constraints automatically — the same approach large C APIs use to generate headers plus Rust/C# bindings.

## How to Apply

1. Make the public API plain C in a portable subset (roughly C89-callable): avoid untagged unions, variadic functions, and global structures.
2. Prefer large flat data structures (pass/return by value or filled-out struct) over pointer-graph designs, which create lifetimes a GC language can't track.
3. Scope borrowed pointers to the call: a pointer argument is valid only during that function call — copy out anything needed later. Document the few exceptions (callbacks, returned arrays) explicitly.
4. Use opaque handles for objects the other language shouldn't own, with explicit create/destroy entry points, rather than exposing raw struct layouts.
5. Define the API in a machine-readable spec and generate the C header _and_ per-language bindings from it; let the generator enforce the subset and add idiomatic wrappers (strings, error mapping, RAII/`Dispose`).
6. Map errors to a return code / out-param the generator can turn into the target language's native error idiom (exceptions, `Result`), not to C-only conventions like `errno`.

## Example

```c
// Bindable C surface: opaque handle, flat data, call-scoped pointers, explicit lifetime.
typedef struct image_o image_o;            // opaque; other languages never see the layout

image_o *image_create(uint32_t w, uint32_t h);
void     image_destroy(image_o *img);

// Flat struct in/out beats a pointer graph the binding can't lifetime-manage.
struct image_info { uint32_t width, height, channels; };
struct image_info image_get_info(const image_o *img);

// `pixels` is borrowed for the duration of THIS call only — the binding copies if needed.
int image_write(image_o *img, const uint8_t *pixels, uint64_t size);  // returns error code
```

## Gotchas

- A pointer you let the caller keep past the call is a lifetime contract no GC language can honor — the default must be "valid for this call only," with exceptions documented and rare.
- Untagged unions, bitfields, and variadics compile fine in C but silently can't be bound by many FFIs — the failure shows up only in the binding, not your build.
- Hand-written bindings drift the moment the C API changes; generate them from one spec so they can't fall out of sync.
- Direct 1:1 C bindings are almost never pleasant to use in another language — budget for an idiomatic wrapper layer (the generator should produce it) over the raw FFI.
- Returning a pointer into internal state leaks your layout and ties the caller to your allocator; return a value/struct or fill a caller-provided buffer instead (pairs with [references/caller-owns-memory.md](./caller-owns-memory.md)).

## Related

[references/plugin-architecture.md](./plugin-architecture.md), [references/caller-owns-memory.md](./caller-owns-memory.md), [references/physical-design.md](./physical-design.md)
