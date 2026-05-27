# Sources

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Node model, Typing and connections, Evaluation and reuse, Gotchas
  - The whole node-based creation-graph model: mechanisms exposed as nodes, features built by wiring connectors, typed data on wires, validity-hash caching, output nodes emitting work an owning system schedules, subgraphs/function graphs, and author-chosen asset granularity — generalized away from the engine to any node-graph authoring system.
- **Aspects extracted:**
  - "Creation Graphs" — mechanisms as nodes with connectors, wires between connectors, typed wire data (CPU Image / GPU Image with descriptor + validity_hash), output/terminal nodes (Draw Call) handed to an owning component to schedule, plugins extending nodes and wire data types → `references/node-model.md`
  - "Creation Graphs" — typed data on wires, CPU vs GPU image distinction (system vs video memory), explicit upload/readback/mipchain conversion nodes, shader output node compiled on demand from connected inputs → `references/typing-and-connections.md`
  - "Creation Graphs" — validity_hash as a 64-bit hash a node computes from local settings + each input resource's validity hash, consumer skips cached work, optional per-node caching for heavy ops, execution-context-dependent cost (offline vs runtime) → `references/evaluation-and-compilation.md`
  - "Subgraphs and Function Graphs" — graph Interface (inputs/outputs with display name, id, type hash), Input/Output nodes with a `+` connector, Subgraph Node generating connectors from the interface, import_shallow vs import_flattened, the flatten pre-pass that inlines subgraph nodes and patches dangling boundary connections, no-recursion limit, Create Subgraph from a selection using boundary wires, promotion to an Asset + instancing + overrides → `references/subgraphs-and-functions.md`
  - "More on Creation Graphs" — author-defined asset granularity, single-resource fine-grained graph vs coarse uber-graph (mesh + material + textures), reusing one graph's output as another graph's input, fine vs coarse tradeoff and shareability of immutable assets, granularity sets the project's data layout → `references/authoring-and-introspection.md`
  - "Summer Fun with Creation Graphs" — toolbox vs black-box framing (customizable kit), data-driven/plugin node kinds, procedural texture generation with adjustable parameters, component-exposed nodes (grass placement, SDF volume) producing parametric/procedural content, and identified usability gaps: discoverability, organization, and inability to visualize intermediate node data (introspection) → `references/authoring-and-introspection.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
