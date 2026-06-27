# authoring-and-introspection: Data-Driven Node Types, Default Inputs, Asset Granularity, Introspection, and Hot-Iteration

## Guideline

Define node kinds as data a registry consumes (pins, defaults, evaluate callback) so any plugin extends the toolbox; let an unwired input fall back to an inline-edited constant so literals need no node; let the author choose asset granularity (one resource per graph, or a coarse "uber-graph" of many); build introspection that previews the value at any pin; and make graphs hot-iterable so an edit re-evaluates without a rebuild.

## Contents

- Data-driven node-type registration (the toolbox is data)
- Default / constant inputs for unwired pins
- Asset granularity: fine-grained vs uber-graph, and project data layout
- Introspection: previewing intermediate values; discoverability
- Hot-iteration: edit -> recompile -> re-evaluate

### Rationale

A node graph is valuable precisely because it ships a _toolbox_ rather than a black box: instead of a fixed importer/processor/renderer, the author is handed composable parts and builds the case-specific solution — closer to a customizable kit than a prebuilt feature. That only works if node kinds are data: a plugin registers a new node (its pins, its defaults, its evaluate callback) and a new wire type, and the editor and evaluator pick them up with no core change. Inline constant defaults keep simple graphs readable — most inputs are literals, and forcing a node for each would drown the graph. Asset granularity is a deliberate authoring lever: a graph that outputs a single image or material gives the same fine granularity as a conventional engine and stays shareable across objects, while a graph free to output more becomes a coarser "uber-graph" bundling mesh, material, and textures — appropriate for terrain sculpt/paint maps, generated vertex data such as UV-unwraps, or particle-state buffers in a VFX system, where the data is intrinsically per-object. The granularity choice directly sets the project's data layout, so it is a design decision, not an accident. Introspection is what makes all of this debuggable: the recurring pain in node systems is discoverability (which nodes fit together) and the inability to see the data at an intermediate node — so per-pin preview is not a luxury. Hot-iteration closes the loop: an authoring tool the artist edits live must re-evaluate on change.

### How to Apply

1. Register each node kind as data: its input/output pins, per-input default constants, settings schema, and an evaluate callback — keep the core registry-driven so plugins add nodes and wire types.
2. For each input pin, store a default constant edited inline in the node; when no wire feeds the pin, evaluation uses the default.
3. Distinguish "deliberately constant" from "unwired by mistake" in the UI (e.g. a visibly empty connector) so missing wires are not mistaken for intent.
4. Let a graph output one resource (fine granularity, shareable) or many (coarse uber-graph); pick per asset, knowing it sets the project's data layout — keep immutable, broadly shared assets fine-grained.
5. Build introspection: let the author select any pin and preview the value flowing through it; show which node kinds are compatible to aid discoverability.
6. Support hot-iteration: on edit, recompile (re-flatten + re-sort) and re-evaluate only what the validity hashes show changed (see evaluation-and-compilation.md).

### Example

```c
// Node kinds are DATA in a registry — plugins extend the toolbox, core untouched.
typedef struct node_type_t {
    uint64_t    type_hash;
    const char *category;                 // grouping for discoverability in the picker
    pin_t      *inputs,  *outputs;
    uint32_t    num_inputs, num_outputs;
    void       *default_inputs;           // inline constant per input pin (used when unwired)
    value_t   (*evaluate)(node_t *, const value_t *inputs);
} node_type_t;

void register_node_type(registry_t *r, const node_type_t *t); // any plugin may call

// Asset granularity is the author's choice and sets the project's data layout:
//   fine-grained : [Import Image] -> [Image Output]            // shareable single resource
//   uber-graph   : [Mesh] [Material] [Textures] -> [Output]    // one coarse per-object asset
//
// Introspection: selecting a wire previews the value at that pin (e.g. the
// intermediate image), so the author can tell which node produced wrong data.
```

### Gotchas

- An unwired input quietly using its default hides a forgotten connection; make "no wire" visually distinct from "constant on purpose".
- A coarse uber-graph that bundles a shared texture with per-object data makes that texture unshareable; keep immutable, broadly shared assets in their own fine-grained graphs.
- Granularity is not free-form polish — it sets the project's data layout, so changing it late forces re-authoring; decide it up front (see data-model-guide).
- Discoverability degrades as the toolbox grows; without categories and compatibility hints, authors cannot find or correctly combine nodes.
- No intermediate-value preview means a wrong result is undebuggable — you cannot tell which node along the chain produced bad data; build per-pin introspection early.
- Hot-iteration that recompiles the whole graph on every keystroke stalls the editor; gate re-evaluation on validity-hash changes so only affected nodes rerun.

### Related

[references/node-model.md](./node-model.md), [references/evaluation-and-compilation.md](./evaluation-and-compilation.md), [references/subgraphs-and-functions.md](./subgraphs-and-functions.md), **data-model-guide**, **data-oriented-design-guide**
