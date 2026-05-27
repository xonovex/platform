# subgraphs-and-functions: Subgraphs as Nodes, Function Graphs, Flattening, and Instancing

**Guideline:** Let a graph contain another graph behind a typed input/output interface so a subgraph appears as a single node in its parent (the data-flow analog of a function call); generate the subgraph node's pins from that interface; flatten subgraphs into the parent at compile time by inlining their nodes and patching boundary wires; and promote a subgraph to a reusable asset that many graphs instance and locally override.

## Contents

- Subgraph as a node, with a declared interface (inputs/outputs)
- Interface nodes inside the subgraph and the `+` connector
- Creating a subgraph from a selection (boundary/dangling wires)
- Flattening (inlining) at compile time, and why not recursion
- Function graphs: promoting to an asset, instancing, overrides

**Guideline (detail):** A subgraph is a graph stored inside a node of its parent. The parent treats it as one node; the subgraph treats the boundary as a set of declared inputs and outputs. Define that interface as data on the graph (a list of inputs and a list of outputs, each with a display name for the UI, a stable id for connection targeting, and a type hash for the expected data). The subgraph node reads the interface and dynamically creates a connector per input and per output; special Input and Output nodes inside the subgraph expose those same boundary pins to the inner nodes.

**Rationale:** Encapsulation is how a node graph scales past a screenful. Behind one node you hide arbitrary internal complexity, name the boundary, and reuse it. Defining the interface as data (display name + id + type hash) means the same definition drives the UI, drives connection validity, and survives serialization — the parent's subgraph node and the inner Input/Output nodes are both generated from it, so they cannot disagree. Flattening at compile time is the key implementation move: rather than teach the evaluator to recurse into subgraphs (which would require it to understand subgraph wrappers and interface nodes), a pre-pass walks the graph and copies every real node and connection into flat lists, skips the wrapper/Input/Output nodes, and patches the wires that crossed the boundary so they connect directly to the inner nodes. The evaluator then sees one ordinary flat graph and needs no subgraph awareness at all. Promoting a subgraph to an asset turns "a reusable block" into "a function": many parent graphs hold a subgraph node that points at the asset, get an instance of it, and may override individual inner nodes locally — edit the asset and every instance updates, while local tweaks stay local.

**How to Apply:**

1. Add an interface to the graph object: a list of inputs and a list of outputs, each `{display_name, id, type_hash}`.
2. Make the subgraph node generate one connector per interface input/output; double-click opens the contained graph.
3. Inside the subgraph, use Input and Output nodes that generate their connectors from the interface; give each a `+` connector that adds a new interface entry when something is wired to it.
4. To build a subgraph from a selection: place a subgraph node at the selection's center, cut the selected nodes and their internal connections into it, and use the wires that crossed the selection boundary ("dangling" connections) to derive sensible inputs and outputs; replace those boundary wires in the parent with wires to the new subgraph node, and spawn matching Input/Output nodes inside.
5. To compile: run a flatten pre-pass that adds all nodes and connections to flat lists _except_ subgraph, Input, and Output nodes; on a subgraph node, recurse into its contents and patch the dangling boundary connections onto the parent's wires. Feed the flattened result to the evaluator (see evaluation-and-compilation.md) — the only evaluator change is to consume the flattened form.
6. To make a function graph: create an asset wrapping the subgraph, instance the asset's object into each subgraph node, and allow per-instance overrides (move/change/remove inner nodes) layered on the asset.

**Example:**

```c
// Interface stored on the graph object — drives both the wrapper node's pins
// and the inner Input/Output nodes. (Conceptually like a function signature.)
typedef struct graph_iface_entry_t {
    const char *display_name;  // UI label
    uint64_t    id;            // stable target for connections
    uint64_t    type_hash;     // expected data type of this boundary pin
} graph_iface_entry_t;

typedef struct graph_iface_t {
    graph_iface_entry_t *inputs;  uint32_t num_inputs;
    graph_iface_entry_t *outputs; uint32_t num_outputs;
} graph_iface_t;

// Flatten: inline subgraphs so the evaluator sees one flat DAG, never recursion.
void flatten_graph(const graph_t *g, flat_graph_t *out) {
    for (uint32_t i = 0; i < g->num_nodes; ++i) {
        node_t *n = &g->nodes[i];
        if (is_input_node(n) || is_output_node(n)) continue;   // boundary: patched, not copied
        if (is_subgraph_node(n)) {
            flatten_graph(subgraph_of(n), out);                // inline contents
            patch_boundary_connections(n, out);                // wire parent <-> inner directly
            continue;
        }
        add_node(out, n);
    }
    add_internal_connections(g, out);
    // NOTE: a subgraph that (transitively) contains itself recurses forever -> reject cycles.
}
```

**Gotchas:**

- Flattening cannot express recursion; a self-containing subgraph inlines without end and overflows the stack — detect the cycle and refuse before inlining.
- The interface id, not the display name, is the connection target; renaming for the UI must not change the id, or every wire to that pin breaks.
- Boundary ("dangling") wires are the only signal for what the inputs/outputs should be; mis-detecting which wires cross the selection produces a wrong interface.
- Without a central list of "blessed" types, manually added interface inputs/outputs drift in type and two instances of the same function graph diverge.
- Compiling re-flattens every time unless you cache the imported/flattened result; cache it and invalidate on source change (see evaluation-and-compilation.md).
- An instance override that removes an inner node the asset later depends on leaves a dangling wire after an asset update; resolve overrides against the current asset, not a stale copy.

**Related:** [references/node-model.md](./node-model.md), [references/evaluation-and-compilation.md](./evaluation-and-compilation.md), [references/typing-and-connections.md](./typing-and-connections.md), **data-model-guide**
