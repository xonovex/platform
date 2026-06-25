---
name: node-graph-guide
description: "Use when designing a visual node-based / data-flow graph for content authoring: typed input/output pins, wires between nodes, the graph stored as plain data, compiling/lowering to executable form vs interpreting it, topological evaluation with caching, reusable subgraphs, function graphs with declared inputs/outputs, and parametric/procedural content. Triggers on node editors, pins/connectors/sockets, wiring nodes, type-checking connections, variant/polymorphic nodes, validity hashing, flattening subgraphs, output nodes that emit work, data-driven node definitions, hot-reloading a graph, even when the user doesn't say 'node graph'."
---

# Node graph Guidelines (Visual Data-Flow Authoring)

Engine-agnostic architecture for a visual node-based graph that authors and non-programmers use to assemble behavior by wiring typed nodes together, then compile or evaluate into executable work. The graph itself is plain typed data (see data-model-guide); its node arrays are laid out for traversal (see data-oriented-design-guide); when it emits render/compute work, the scheduling of that work belongs to gpu-rendering-guide.

## Requirements

- A reflectable typed object/data store the graph serializes into (nodes, connections, per-node settings, interface), so the editor, evaluator, and tools all read the same definition — see data-model-guide.
- A node-type registry that plugins can extend with new node kinds and new wire data types without modifying the core.

## Essentials

- **Graph is data, not code** - Nodes, typed pins, wires, and constant inputs are serialized objects an editor/evaluator/tool all consume, see [references/node-model.md](references/node-model.md)
- **Typed pins gate every wire** - A connection is legal only when the producer's output type satisfies the consumer's input type; check at author time, see [references/typing-and-connections.md](references/typing-and-connections.md)
- **Compile, then run** - Lower the authored graph into a flattened, ordered executable form and evaluate that, instead of walking the editor graph every frame, see [references/evaluation-and-compilation.md](references/evaluation-and-compilation.md)
- **Cache by validity hash** - Each output carries a hash of its node's settings plus its inputs' hashes; a consumer skips recompute when the hash is unchanged, see [references/evaluation-and-compilation.md](references/evaluation-and-compilation.md)
- **Subgraphs are reusable nodes** - Encapsulate a selection behind a typed input/output interface and instance it from an asset, see [references/subgraphs-and-functions.md](references/subgraphs-and-functions.md)

## Node model

- **Nodes and pins** - A node is a typed object with named, typed input and output pins; wires connect one output pin to one input pin, see [references/node-model.md](references/node-model.md)
- **Data-driven node types** - Register node kinds (pins, defaults, evaluate callback) as data so plugins extend the toolbox, see [references/authoring-and-introspection.md](references/authoring-and-introspection.md)
- **Constant/default inputs** - An unwired input falls back to a stored constant edited inline; no node is needed for a literal, see [references/authoring-and-introspection.md](references/authoring-and-introspection.md)
- **Output nodes emit work** - Terminal nodes hand their result (image, draw call, buffer) to the owning system that schedules it, see [references/node-model.md](references/node-model.md)

## Typing and connections

- **Type hashes on pins** - Each pin stores a type identity; connection validity and conversion are decided from it, see [references/typing-and-connections.md](references/typing-and-connections.md)
- **Variant/polymorphic nodes** - A node may resolve concrete pin types from what is wired in (e.g. CPU vs GPU image), see [references/typing-and-connections.md](references/typing-and-connections.md)
- **Conversions, not silent coercion** - Insert explicit convert nodes (e.g. upload, readback) rather than coercing mismatched types, see [references/typing-and-connections.md](references/typing-and-connections.md)

## Evaluation and reuse

- **Topological evaluation** - Resolve each node's inputs before it runs; the wire DAG gives the order, see [references/evaluation-and-compilation.md](references/evaluation-and-compilation.md)
- **Flatten subgraphs at compile** - A pre-pass inlines subgraph nodes and patches boundary wires, so the evaluator never recurses, see [references/subgraphs-and-functions.md](references/subgraphs-and-functions.md)
- **Function graphs and instancing** - Promote a subgraph to an asset; many graphs instance it and override locally, see [references/subgraphs-and-functions.md](references/subgraphs-and-functions.md)
- **Asset granularity** - One graph can output a single resource or an "uber-graph" of many; the choice sets project data layout, see [references/authoring-and-introspection.md](references/authoring-and-introspection.md)

## Gotchas

- A node only caches correctly when its validity hash folds in every input that changes its output; an input read but left out of the hash yields stale reuse with no error.
- Flattening cannot express recursion — a subgraph that (transitively) contains itself inlines forever and overflows the stack; detect cycles before inlining.
- An unwired input silently using its stored default is convenient but hides missing connections; distinguish "deliberately constant" from "forgot to wire" in the UI.
- Polymorphic pins resolved from neighbors can leave a graph in a half-typed state mid-edit; re-resolve and re-validate affected wires on every connect/disconnect, not just on evaluate.
- Editing the editor graph and re-flattening per evaluation is correct but wasteful; cache the compiled/flattened result and invalidate it only when the source graph changes.
- A coarse "uber-graph" that bundles a shared texture with per-object data makes the texture unshareable across objects; keep immutable, broadly shared assets in their own fine-grained graphs.
- Intermediate values are invisible unless you build introspection; without per-pin preview, authors cannot tell which node produced wrong data.
- The owning system, not the graph, schedules emitted work; if no system consumes an output node the work silently never runs.

## Progressive Disclosure

- Read [references/node-model.md](references/node-model.md) - Load when defining nodes, typed pins, wires, output/terminal nodes, or the graph-as-data representation
- Read [references/typing-and-connections.md](references/typing-and-connections.md) - Load when type-checking connections, adding conversions, or building variant/polymorphic nodes
- Read [references/evaluation-and-compilation.md](references/evaluation-and-compilation.md) - Load when compiling/lowering a graph, ordering evaluation, or caching results by validity hash
- Read [references/subgraphs-and-functions.md](references/subgraphs-and-functions.md) - Load when encapsulating a subgraph, defining a function-graph interface, flattening/inlining, or instancing from an asset
- Read [references/authoring-and-introspection.md](references/authoring-and-introspection.md) - Load when registering data-driven node types, default/constant inputs, asset granularity, introspection, or hot-iterating a graph
