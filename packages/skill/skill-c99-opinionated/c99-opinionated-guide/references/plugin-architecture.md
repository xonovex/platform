# plugin-architecture: Plugins and a Runtime Interface Registry

## Guideline

Build the system as many small, independently-loadable plugins that talk only through a central registry of named interfaces — each interface a plain-C struct of function pointers registered under a unique string id — so components discover each other at runtime with zero compile-time coupling and the core stays tiny.

## Rationale

A monolith couples everything at compile time: to use a capability you must link and include its headers, so the dependency graph and the core both grow without bound. Inverting this into "lots of little machines co-operating" needs a runtime rendezvous point. A registry that maps a string id to a struct of function pointers gives exactly that: a plugin `add()`s its interface, any other plugin `first()`s it by name and calls through the table — neither knows the other at build time. Grouping functions into one struct also acts as a namespace and lets you pass a whole API by one pointer. Plain C for the interface is deliberate: C has a stable de-facto ABI (C++ does not), so plugins built by different compilers/versions stay binary-compatible, and the simple shape keeps every API consistent. Supporting _multiple_ implementations of one interface (`first`/`next`) turns extension points into mere "register another implementation of this id" — unit tests, asset importers, and tools all become plugins.

## How to Apply

1. Define the registry: `add(name, iface)`, `remove(iface)`, `first(name)`, `next(prev)` — a string-keyed multimap of interface pointers.
2. Express each capability as a struct of function pointers plus a unique string-id constant and docs, all in one plain-C header (no header-to-header includes — see [references/physical-design.md](./physical-design.md)).
3. On load, a plugin fills in its interface struct and `add()`s it under its id; on unload it `remove()`s it (this is also the hot-reload hook — see [references/hot-reload.md](./hot-reload.md)).
4. To use a capability, `first()` it by id and call through the table; treat a missing interface as an optional dependency, not a link error.
5. Make extension points "register another implementation of interface X": iterate `first`/`next` to invoke all providers (tests, importers, exporters) without the core knowing them.
6. Version interfaces by id (append a version to the name, or add a version field) so a plugin can ask for the exact shape it understands.

## Example

```c
// The registry: string id -> struct-of-function-pointers, multiple impls allowed.
struct api_registry_i {
    void  (*add)(const char *name, void *iface);
    void  (*remove)(void *iface);
    void *(*first)(const char *name);          // first implementation registered under name
    void *(*next)(void *prev);                 // walk additional implementations
};

// A capability is just a named struct of function pointers.
#define COMPRESSION_I_NAME "compression_i"
struct compression_i {
    void (*compress)(uint8_t *dst, const uint8_t *src, uint64_t n);
    void (*decompress)(uint8_t *dst, const uint8_t *src, uint64_t n);
};

// Producer plugin: register on load.
reg->add(COMPRESSION_I_NAME, &my_compression);
// Consumer plugin: discover by name, no compile-time link to the producer.
struct compression_i *c = reg->first(COMPRESSION_I_NAME);
if (c) c->compress(dst, src, n);               // optional dependency: handle absence
```

## Gotchas

- A cached interface pointer dangles after the providing plugin is unloaded/reloaded — re-fetch from the registry after a reload, never stash it across that boundary (see [references/hot-reload.md](./hot-reload.md)).
- String-id typos fail silently as "interface not found"; centralize the id in a `#define` next to the struct and use that constant everywhere.
- C++ interfaces break this — no stable ABI across compilers; keep the _interface_ plain C even if the _implementation_ is C++.
- `first()` returning the most-recently-added (or first) implementation makes order matter when only one is expected — design for either "exactly one" or "iterate all," and document which.
- The registry is global mutable state; register during init and treat the interface tables as immutable afterwards to avoid races with running plugins.

## Related

[references/physical-design.md](./physical-design.md), [references/hot-reload.md](./hot-reload.md), [references/composability.md](./composability.md), [references/cross-language-api.md](./cross-language-api.md)
