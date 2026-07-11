# A Metapatterns Lens (One Synthesis, Not Canonical)

One author's synthesis for building intuition — a mental model, NOT a standard definition. Mainstream catalogs keep hexagonal and microkernel as parallel, distinct patterns. Read it as a lens, not a law.

## The lens

- **A microkernel ≈ many ports-and-adapters cores sharing a common middleware** — each plug-in is a small core behind a port; the kernel is the shared substrate they plug into.
- **A kernel reads as a "middleware-gateway"** — a _gateway_ down to drivers/resources, a _middleware_ up to the hosted applications.
- **A distributed microkernel relates to a mesh** — spread the plug-in cores across nodes and the substrate becomes a network of intermediaries.

## Caveats — state them plainly

- **Two different "microkernels" got fused** — the OS sense (a minimal privileged core delegating to user-space servers) and the architecture plug-in pattern are kept _separate_ by standard references. The lens blends them; real definitions do not.
- **"Middleware-gateway" pairs a category with its sub-type** — a gateway is a specialization of middleware, so the compound names a general thing and one of its kinds together. Handy shorthand, sloppy taxonomy.
- **A service mesh does NOT virtualize system resources** — it is a networking data plane (traffic management, mTLS, observability). The "distributed kernel ≈ mesh" intuition is about mediation between distributed parts, not virtualizing compute/memory/storage.

## What to carry

"Plug-ins are adapters behind ports" and "a distributed plug-in core can ride a mesh (keep the port contract identical to the in-process case)" are good instincts. For load-bearing definitions use **hexagonal-pattern-guide** (ports/adapters, the dependency rule) and [core-plugins-registry.md](core-plugins-registry.md) (kernel/registry/composition-root). For coupling vocabulary, **connascence-guide**.

Back to [SKILL.md](../SKILL.md).
