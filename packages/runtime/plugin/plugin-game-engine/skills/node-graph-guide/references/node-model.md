# node-model: Nodes, Typed Pins, Connections, and the Graph as Data

## Guideline

Model the graph as plain serialized data (a node list, a connection list, and per-node settings) where each node is a typed object exposing named, typed input and output pins, and dedicated terminal "output" nodes hand finished results to the system that owns the graph.

## Rationale

The terminal-node split matters because the graph computes values but does not decide when or where they run: an output node exposes a result (an image, a draw call, a buffer) that an owning system later schedules.

## How to Apply

1. Store the graph as a typed object with at least: a node list, a connection list, per-node data/settings, and (optionally) comments, see data-model-guide for the store itself.
2. Define each node as a typed object that declares named input pins and named output pins, each pin carrying a type identity.
3. Represent a connection as `(from_node, from_output_pin) -> (to_node, to_input_pin)`.
4. Let unwired inputs fall back to a stored constant (see authoring-and-introspection.md) so literals need no node.
5. Provide terminal output-node kinds whose job is to surface a result to the owning system; the owning system reads the output and schedules the work it represents.
6. Keep the node-type registry open: plugins register new node kinds and new wire data types as data.

## Example

```c
// The graph is data: arrays of nodes and connections, plus per-node settings.
typedef struct pin_t {
    const char *name;
    uint64_t    type_hash;     // identity of the data this pin carries
} pin_t;

typedef struct node_t {
    uint64_t  type_hash;       // which registered node kind this is
    uint64_t  id;              // stable id, used by connections
    pin_t    *inputs,  *outputs;
    uint32_t  num_inputs, num_outputs;
    void     *settings;        // node-local data + constant defaults for unwired inputs
} node_t;

typedef struct connection_t {
    uint64_t from_node, to_node;   // one output -> one input; one output may feed many inputs
    uint32_t from_output, to_input;
} connection_t;

typedef struct graph_t {
    node_t       *nodes;       uint32_t num_nodes;
    connection_t *connections; uint32_t num_connections;
} graph_t;

// A terminal node does not schedule work; it exposes a result the owner consumes.
// e.g. an "Image Output" exposes a finished image; a "Draw Call" output exposes
// recorded draw work that the owning render component later schedules.
```

## Gotchas

- Node `id`s must be stable across edits, because connections reference nodes by id; reusing an id after delete silently rewires.
- A wire data type is itself data: register it so plugins can introduce new payloads (e.g. a particle buffer, an SDF volume) the core never hard-codes.
- The graph computes results but does not run them; if no owning system reads an output node, its work never executes (see evaluation-and-compilation.md).

## Related

[references/typing-and-connections.md](./typing-and-connections.md), [references/evaluation-and-compilation.md](./evaluation-and-compilation.md), [references/authoring-and-introspection.md](./authoring-and-introspection.md), **data-model-guide**, **data-oriented-design-guide**
