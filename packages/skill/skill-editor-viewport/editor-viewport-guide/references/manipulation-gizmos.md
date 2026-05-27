# manipulation-gizmos: Move/Rotate/Scale Gizmos Decoupled From Object Type

**Guideline:** Keep gizmo rendering and interaction central and generic; let any object plug into it by implementing a tiny get/set-transform interface (queried, not inherited) instead of giving each component its own gizmo — so one move/rotate/scale tool manipulates transforms, spline control points, wire endpoints, or any custom "position" concept with identical behavior.

**Rationale:** If every component type implements its own gizmo, behavior drifts, code duplicates, and improving the tool means editing N components. The opposite extreme — letting components fully draw and hit-test their own gizmos — is even worse for the same reasons. The stable middle is: the editor owns all gizmo geometry, picking, and drag logic; components supply only data through a small interface the editor queries at runtime. Because the interface is queried (a function-pointer table looked up by name) rather than baked into a base class, a component opts in without the editor knowing it exists, and a component that has no transform simply does not implement it. Returning both world and local transforms lets the same gizmo operate in either coordinate space on a modifier key. A priority value resolves the ambiguous case of an object with several transformable aspects, letting a primary transform win while still allowing exotic position concepts (2D, double-precision astronomical coordinates) to participate.

**How to Apply:**

1. Define a queried interface, not a base class: components register an implementation and the editor looks it up by name (id hash), so coupling is one-directional.
2. Give the gizmo interface two callbacks: `get_transform(object) -> {world, local}` and `set_transform(object, local, undo_scope)`.
3. Have `get_transform` return both world and local so the gizmo can apply a delta in world space or local space depending on the user's mode.
4. Use the `undo_scope` argument to distinguish an in-progress drag (zero — coalesce, no undo entry yet) from a committed change (non-zero — open one undo step).
5. On selection, walk up the ownership chain in the data model to the component responsible for what was clicked (e.g. a spline control point resolves to its spline component).
6. When an object exposes several transformable components, query each one's `priority()` and manipulate the highest — let the canonical transform win by default.
7. Keep all handle drawing, hit-testing, and the drag math (see gizmo-math.md) in the central gizmo code; components contribute only the get/set.

**Example:**

```c
// Queried interface (function-pointer table), looked up by name — no inheritance.
typedef struct gizmo_transform_i {
  // Both spaces, so the gizmo can work in world or local on a modifier key.
  bool  (*get_transform)(data_model_o *m, entity_ctx_o *ctx, uint64_t object,
                         transform_t *world, transform_t *local);
  // undo_scope == 0 -> drag in progress (coalesce); != 0 -> commit one undo step.
  void  (*set_transform)(data_model_o *m, entity_ctx_o *ctx, uint64_t object,
                         const transform_t *local, uint64_t undo_scope);
  // Highest priority wins when an object has several transformable aspects.
  float (*priority)(void);
} gizmo_transform_i;

// Components register; the editor never names them. A spline's control point and a
// wire's endpoint both implement this, so one gizmo edits all of them identically.
registry_add(GIZMO_TRANSFORM_INTERFACE, transform_component_gizmo);
registry_add(GIZMO_TRANSFORM_INTERFACE, spline_component_gizmo);
```

**Gotchas:**

- Letting each component draw/hit-test its own gizmo duplicates code and lets behavior drift; keep rendering and interaction central, expose only data.
- Using inheritance/a base class couples the editor to component types and forces transform-less components to stub it out; query an interface by name instead so opting in is optional.
- Returning only the world (or only the local) transform breaks the other coordinate mode; always provide both.
- Ignoring `undo_scope` either floods undo with per-frame steps or never records the change; coalesce while in-progress, commit once on release.
- Selecting a sub-feature (control point, endpoint) without walking up to the owning component leaves the gizmo with nothing actionable; resolve the ownership chain.
- Without a priority, an object with multiple transformable components is ambiguous and the gizmo grabs the wrong one; pick the highest priority.

**Related:** [references/gizmo-math.md](./gizmo-math.md), [references/render-editor-integration.md](./render-editor-integration.md), **ecs-guide**, **data-model-guide**
