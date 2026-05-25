# composability: Composable Stages Over a Uniform Currency

**Guideline:** Build small, unopinionated primitives/stages that compose over one uniform data currency; let the caller wire the sequence explicitly. A stage's "phase"/role is how it is used, not a category the library bakes in — the consumer picks the 1..N stages it needs.

**Rationale:** A monolithic `do_everything()` forces every caller into one fixed pipeline and one set of trade-offs. Decomposing the work into stages that read and write the _same_ data type lets each consumer assemble exactly the funnel it needs (1 stage for the trivial case, more for the complex one), swap one stage's implementation without touching the others, and test each stage in isolation. Keeping composition explicit (plain function calls in the caller, not a registered callback/vtable pipeline) preserves readability, debuggability, and the no-hidden-dispatch property of systems C — the "pipeline" is just the code the caller writes.

**How to Apply:**

1. Define the **currency**: one caller-owned data type that flows between stages (e.g. an index-pair set, a manifold/contact buffer, a mesh-vertex stream). Index-based, not pointer-based, so stages stay relocatable and cache-friendly.
2. Split the work into **role-shaped stages** over that currency: a _source_ (produces the currency), zero or more _filters_ (currency → narrower currency), and a _sink/solver_ (currency → result). Name stages by what they _do_, never by a fixed "phase".
3. Keep each stage **unopinionated**: caller-owns-memory, no global state, no policy baked in (the consumer supplies layers/masks/predicates/config).
4. **Compose explicitly** in the consumer — sequential calls, no function-pointer dispatch table or registration framework. Fuse adjacent stages only as a measured optimization.
5. Let trivial cases **skip stages**: a one-stage path (e.g. all-pairs solve) must not pay for machinery the N-stage path needs.

**Example:**

```c
// Bad: one monolithic call bakes in the whole pipeline and its policy.
void collide_everything(world_t *w);   // can't reuse a phase, swap a stage, or test in isolation

// Good: composable stages over a uniform `pairs` currency; caller wires the funnel.
//   source -> (filter)* -> solver
uint32_t pairs_generate(const bodies_t *b, pairs_t *out);          // SOURCE: bodies -> candidate pairs
uint32_t pairs_filter_layer(const uint32_t *layer, const uint32_t *mask,
                            const pairs_t *in, pairs_t *out);      // FILTER: pairs -> pairs
size_t   solve_aabb(manifold_t *m, const aabb_t *boxes,
                    const pairs_t *pairs);                         // SOLVER: pairs -> contacts

// 1-stage consumer (spheres): no pairs at all.
solve_sphere_all(&m, px, py, pz, r, n);

// 3-stage consumer: the caller IS the pipeline — explicit, no dispatch object.
pairs_generate(&bodies, &pairs);
pairs_filter_layer(layer, mask, &pairs, &pairs);   // in/out may alias (compaction)
solve_aabb(&manifold, boxes, &pairs);
```

**Counter-Example:** A genuinely single-step operation with no reuse or variation does not need a stage split — forcing a funnel there adds ceremony without payoff. Composability biases the design; it is not a mandate to fragment every function.

**Related:** [references/caller-owns-memory.md](./caller-owns-memory.md), [references/implementation-variants.md](./implementation-variants.md), **data-oriented-design-guide**
