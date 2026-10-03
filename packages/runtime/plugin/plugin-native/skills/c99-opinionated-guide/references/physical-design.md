# physical-design: Physical Design and Header Discipline

Enforce one machine-checkable rule: a header may not `#include` another header (bar a tiny allowlist like `<stdint.h>`/`<stdbool.h>`), so each header is the self-contained public interface of exactly one system, dependencies stay acyclic and explicit, and editing a header doesn't cascade a project-wide rebuild. Cycles become impossible: a header can't name another header.

## How to Apply

1. Make each header the public interface of one system; map folders to system boundaries, with no cross-system header dependencies.
2. Forbid header-including-header; keep a minimal allowlist (`<stdint.h>`, `<stdbool.h>`). Lint it in CI so it can't erode.
3. Use opaque handle types to hide platform/implementation detail in the interface: `struct file_o { uint64_t opaque; };`. The `.c` casts to the real type.
4. Replace would-be header dependencies with forward declarations and pointer-based interfaces, so the header needs only an incomplete type.
5. Keep public interfaces in plain C even if implementations are C++/other: C headers avoid templates/inheritance leaking into the interface.
6. If you need templates/generics, isolate them in a separate include kind (e.g. `.inl`) that's allowed to include its own kind, and keep those few and deliberate.

## Example

```c
// system_a.h: self-contained interface, no other project header included.
#include <stdint.h>
struct other_system_o;                 // forward declaration instead of #include "other.h"
struct file_o { uint64_t opaque; };    // opaque handle hides the platform type

void system_a_run(struct system_a_o *a, struct other_system_o *dep, struct file_o out);

// system_a.c: the ONLY place that pulls in concrete dependencies.
#include "system_a.h"
#include "other_system.h"               // .c may include freely; the graph stays directional
#include "platform_file_impl.h"
```

## Gotchas

- The rule only holds if it's enforced. One "harmless" header include reintroduces the cascade; gate it in CI, don't rely on discipline.
- Opaque handles push type-confusion bugs to runtime (you cast `opaque` to a concrete type in the `.c`); keep the cast in exactly one place and assert the tag if you can.
- Forward-declaration-only interfaces mean callers can't see struct sizes/layout: pass by pointer and apply the caller-owned storage contract from **memory-management-guide**.
- A header that "needs" another system's struct by value is a design smell. It's reaching across a boundary; pass a handle/pointer instead.
- This governs _physical_ layout; the _runtime_ wiring of those systems (discovery, registration) is [references/plugin-architecture.md](./plugin-architecture.md).

## Related

[references/plugin-architecture.md](./plugin-architecture.md), [references/file-naming.md](./file-naming.md), **memory-management-guide**, [references/hot-reload.md](./hot-reload.md)
