# A Metapatterns Lens (One Synthesis, Not Canonical)

One author's synthesis for building intuition — useful as a mental model, but NOT a standard definition; mainstream catalogs keep hexagonal and microkernel as parallel, distinct patterns.

**Read this as a lens, not a law** — it fuses ideas to spark intuition; the load-bearing definitions live elsewhere. Attribution for the synthesis lives in SOURCES.md.

## The lens in one breath

**A microkernel ≈ many ports-and-adapters cores sharing a common middleware** — each plug-in is a small core hiding behind a port; the kernel is the shared substrate they all plug into.

**A kernel reads as a "middleware-gateway"** — a _gateway_ down to lower drivers/resources, and a _middleware_ up to the applications it hosts.

**A distributed microkernel relates to a mesh** — spread the plug-in cores across nodes and the shared substrate becomes a network of intermediaries between them.

```
        applications (upper)
                |
   [ kernel = middleware up / gateway down ]
                |
        drivers / resources (lower)
```

```go
// each plug-in is a small hexagonal core behind a port
type Channel interface{ Send(ctx Context, msg Payload) (Receipt, error) }
type Capable interface{ Capabilities() CapabilitySet }

// the Registry is the shared substrate (composition root) all cores plug into
type Registry struct {
    ch   map[string]Channel
    caps map[string]Capable
}
func (r *Registry) Add(name string, c Channel) { r.ch[name] = c } // open for extension
```

The kernel mediates: applications ask for a capability, the registry routes to the right adapter, the adapter speaks to the lower driver. That mediation is exactly what makes "gateway down, middleware up" feel right.

## Caveats — state them plainly

**Two different "microkernels" got fused** — the OS-microkernel sense (a minimal privileged core delegating to user-space servers) and the software-architecture plug-in pattern (a core system extended by independent plug-in modules) are kept _separate_ by standard references. The lens blends them; real definitions do not.

**"Middleware-gateway" pairs a category with its sub-type** — a _gateway_ is a _specialization_ of _middleware_, so the compound names a general thing and one of its kinds together. Handy shorthand, sloppy taxonomy.

**"A service mesh virtualizes system resources" is not standard** — a service mesh is a _networking data plane_: service-to-service traffic management, security (mTLS), and observability. It does **not** virtualize compute, memory, or storage. The "distributed kernel ≈ mesh" intuition is about _mediation between distributed parts_, not resource virtualization.

```text
BAD  -> "Use a service mesh to virtualize the cluster's CPU and memory."
GOOD -> "A service mesh routes, secures, and observes traffic between
         distributed plug-in cores; it does not virtualize resources."
```

```text
BAD  -> "Microkernel and hexagonal are the same pattern."
GOOD -> "A plug-in is an adapter behind a port; the patterns rhyme,
         but the catalogs define them separately."
```

## Where the lens earns its keep

**Plug-ins are adapters behind ports** — model each message (a neutral Payload value, a self-declared Capability) as something the kernel consumes through a stable port, never a concrete type.

**A distributed plug-in core can ride a mesh** — when cores live on different nodes, let the mesh carry their traffic; keep the _contract_ (the port) identical to the in-process case.

**The kernel stays a fail-closed mediator** — complete mediation plus fail-safe defaults: an unknown or unregistered capability is denied, not improvised.

```go
func (r *Registry) Send(name string, ctx Context, msg Payload) (Receipt, error) {
    c, ok := r.ch[name]
    if !ok {
        return Receipt{}, fmt.Errorf("channel %q not registered", name) // fail closed
    }
    return c.Send(ctx, msg) // every call mediated through the port
}
```

## Bottom line

**Take the intuition, drop the dogma** — "plug-ins are adapters behind ports" and "a distributed plug-in core can ride a mesh" are good instincts to carry.

**Rely on the real definitions, not this lens** — for the load-bearing concepts use **hexagonal-pattern-guide** for ports/adapters and the dependency rule, and [core-plugins-registry.md](core-plugins-registry.md) for the kernel/registry/composition-root mechanics. For coupling vocabulary reach for **connascence-guide**.

Back to [SKILL.md](../SKILL.md).
