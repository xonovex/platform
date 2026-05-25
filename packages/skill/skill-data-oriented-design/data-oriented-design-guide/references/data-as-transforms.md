# data-as-transforms: Data as Transforms (Where There Is One, There Are Many)

**Guideline:** Design from the real data and its statistics, modeling each system as a bulk transform from an input stream to an output stream, rather than from an idealized object model.

**Rationale:** The principle "where there is one, there are many": singletons are rare in real workloads — you almost always have arrays of the same thing. Object-oriented modeling optimizes the single instance (encapsulation, virtual dispatch, per-object state) and pessimizes the common case of processing thousands. Thinking in transforms makes the data and its access pattern primary, which is exactly what the hardware rewards. It also clarifies what work actually exists: you read the input statistics ("how many, how often, what distribution") and design the layout around them.

**Techniques:**

- **Tables and streams mental model** - Picture each piece of state as a table (columns = fields, rows = instances). A system is a function `table_in -> table_out`. This naturally pushes you toward column storage (SoA) and batch loops.
- **One function, many items** - Replace `thing.update()` called per object with `update_all(things, n)` that loops internally. The per-call overhead, branch prediction, and instruction cache all improve.
- **Separate the three jobs** - Reading input, transforming, writing output are distinct phases; keep them as separate linear passes rather than interleaving them per object.
- **Design around the statistics** - Most entities idle, most rays miss, most cells are empty: lay out for the common case and handle the rare case as a separate, smaller stream.
- **Data first, code second** - Decide the layout from how data is produced and consumed, then write the code that fits it — not the reverse.

**How to Apply:**

1. List the inputs each system actually reads and the outputs it produces; ignore the conceptual "object."
2. Express the system as `process(const In *in, Out *out, size_t n)`.
3. Group same-typed instances into contiguous arrays (one transform sees one type).
4. Order phases so each pass is a single linear sweep over its input.

**Example:**

```c
// Bad: OO model — virtual update per object, scattered state, one at a time.
class Entity { public: virtual void update(float dt) = 0; };
for (Entity *e : entities) e->update(dt); // dispatch + pointer chase per item

// Good: transform model — homogeneous stream in, stream out.
typedef struct { vec3 pos; vec3 vel; } body_t; // a "row"
static void integrate(const body_t *in, body_t *out, size_t n, float dt) {
  for (size_t i = 0; i < n; i++) {        // one transform, many rows
    out[i].vel = in[i].vel;               // (gravity etc. applied here)
    out[i].pos = vec3_add(in[i].pos, vec3_scale(in[i].vel, dt));
  }
}
```

**Gotchas:**

- The object model is a comfortable lie about the data; the hardware sees bytes and streams, not objects.
- Resist hiding the loop behind a per-item method — that reintroduces dispatch and kills batching.

**Related:** [references/soa-aos-aosoa.md](./soa-aos-aosoa.md), [references/existence-based-processing.md](./existence-based-processing.md), [references/access-patterns.md](./access-patterns.md), [references/cache-behavior.md](./cache-behavior.md)
