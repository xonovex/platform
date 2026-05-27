# typing-and-connections: Pin Types, Connection Validity, Conversions, and Variant Nodes

**Guideline:** Give every pin a stable type identity (a type hash), permit a wire only when the producer's output type satisfies the consumer's input type, surface mismatches at author time, insert explicit conversion nodes instead of silently coercing, and let a node resolve concrete pin types from what is wired into it when it is genuinely polymorphic.

**Rationale:** The whole point of a visual authoring graph is that a non-programmer can wire it without producing a runtime crash; the type system is what makes that safe. If a pin carries a stable type hash, the editor can reject illegal connections as they are drawn, grey out incompatible connectors, and explain why — feedback the author needs immediately, not at evaluation. Refusing to silently coerce is equally important: the difference between a CPU image (bits in system memory) and a GPU image (bits in video memory) is not cosmetic, and quietly papering over it hides real cost (an upload, a readback) and real failure modes. Making the conversion an explicit node keeps the data flow honest and lets the author see and place the cost. Variant nodes (one node that adapts its pin types to its neighbors) avoid a combinatorial explosion of near-identical node kinds, but they push a burden onto the editor: pin types must be re-resolved every time a wire changes, or the graph drifts into a half-typed state where validation lies.

**How to Apply:**

1. Assign each pin a `type_hash`; treat it as the connection contract, not a hint.
2. On a connect attempt, compare output `type_hash` to input `type_hash`; accept on a match (or a declared subtype/satisfies rule), reject otherwise with a visible reason.
3. When two types are related but not identical (e.g. CPU image vs GPU image), offer an explicit conversion node (upload, readback, format change) rather than coercing.
4. For a polymorphic node, mark pins as "resolved from connection"; when a neighbor wire changes, re-derive the concrete types and re-validate every wire the change touched.
5. Re-run validation on every connect/disconnect, not only at evaluation, so the editor never reports a graph as valid while a polymorphic pin is unresolved.
6. Let plugins register new wire data types with their own type hashes and any legal conversions.

**Example:**

```c
// A connection is legal only if the producer's type satisfies the consumer's.
bool can_connect(const pin_t *out, const pin_t *in) {
    if (out->type_hash == in->type_hash) return true;
    return type_satisfies(out->type_hash, in->type_hash); // declared subtype / variant rule
}

// Don't coerce CPU<->GPU silently; require an explicit conversion node.
//   [Load Image]--(CPU Image)-->[Upload]--(GPU Image)-->[Filter Image (GPU)]
//                                  ^ explicit node makes the upload cost visible

// A polymorphic "Filter Image" resolves its pin types from what is wired in:
void resolve_pins(node_t *n, const graph_t *g) {
    uint64_t in_t = incoming_type(g, n, /*input*/ 0); // CPU Image or GPU Image
    n->inputs[0].type_hash  = in_t;
    n->outputs[0].type_hash = in_t;                   // output matches resolved input
    // MUST re-run on every connect/disconnect, then re-validate downstream wires.
}
```

**Gotchas:**

- Comparing pin types only at evaluation is too late; the author has already wired a broken graph and lost the context to fix it. Validate on connect.
- A variant pin left unresolved (nothing wired yet) has no concrete type; a downstream type check against it must report "unresolved", not "compatible".
- Silent coercion between CPU and GPU payloads hides an upload or a readback — costs that dominate a frame; force an explicit conversion node.
- A conversion node is still a node with a validity hash; an upload/readback is cacheable, so do not re-run it when its input is unchanged (see evaluation-and-compilation.md).
- Without a central registry of "blessed" types, hand-added input/output types drift and two graphs disagree on what a type means; register types as data.

**Related:** [references/node-model.md](./node-model.md), [references/evaluation-and-compilation.md](./evaluation-and-compilation.md), [references/subgraphs-and-functions.md](./subgraphs-and-functions.md), **data-model-guide**
