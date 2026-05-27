# state-replication: Replicating Object and Component State

**Guideline:** Replicate the game by synchronizing the state of opted-in objects and their components across nodes: make replication an explicit per-component capability and a per-object opt-in flag, detect changes on the owning node, and send those changes only to nodes that are interested in that object — defaulting to a simple whole-object copy and allowing a per-field description when bandwidth matters.

**Rationale:** A multiplayer game is, at bottom, the same simulation running on several nodes kept in agreement about object state. If everything were replicated automatically the wire would drown in irrelevant data; if replication required hand-written serialization for every type, prototyping would be painful. The resolution is a two-level opt-in: a component declares whether and how it replicates (no declaration means it never goes on the wire), and each object instance is individually flagged for replication (an object whose components are replicable still isn't sent unless flagged). The owning node watches replicated components for changes and pushes diffs outward. The default path can be deliberately inefficient — copy the whole component as an opaque blob — so multiplayer "just works" early; later you hand the system a field-by-field layout description and per-field change-check cadence so only changed members travel. This lets you ship a prototype fast and optimize the wire format exactly where profiling says to.

## Contents

- Two-level opt-in (component capability + per-object flag)
- Change detection and watch cadence
- Whole-object blob vs. described layout
- Relevancy: sending only to interested nodes
- Replicating events and variables (RPC-like)

**How to Apply:**

1. Make replication a per-component capability: a component without a replication declaration is never synchronized. The declaration says _whether_ and _how_ it replicates.
2. Require a separate per-object replication flag: only objects explicitly marked for replication are sent, even if their components are replicable. This keeps editor scratch objects off the wire.
3. On the owning node, detect changes by watching replicated components on a cadence (a watch timer): every N seconds/frames, compare and, if changed, queue a diff. A larger N trades freshness for bandwidth.
4. Default to the cheap path: with no layout info, send the whole component as one opaque binary blob — correct, easy, wasteful.
5. Optimize by description: hand the system the struct layout and a per-field watch cadence so it sends only the members that actually changed.
6. Send updates only to interested nodes — nodes that have the object in scope/relevancy — not to every connected node. See delta-and-snapshots.md for how the first send to a newly-interested node differs.
7. For one-shot facts (events, variable sets) replicate the event/variable directly: send a compact identifier (e.g. an event name hash) plus the target object id; the receiver looks up the object by id and applies it — see data-model-guide for stable networked ids.

**Example:**

```c
// Component declares IF/HOW it replicates. No declaration => never on the wire.
typedef struct replicated_transform {
    float pos[3];
    float rot[4];
} replicated_transform;

static net_component_replication_i transform_replication = {
    .watch_seconds = 0.05f,   // owner checks for changes ~20x/sec
    // .layout = &transform_layout  // OPTIONAL: per-field diffing instead of
                                     // a whole-struct blob (bandwidth opt)
};

component_i transform = {
    .name = "transform",
    .bytes = sizeof(replicated_transform),
    .replication = &transform_replication,
};
// And the OBJECT must still be flagged for replication in the world/tree,
// or none of its components are ever sent.
```

**Gotchas:**

- Two independent opt-ins: a component being replicable does nothing unless the object instance is also flagged — forgetting the object flag is a silent "nothing replicates."
- Without a layout description the whole component is sent even if one field changed; for big or frequently-touched components this is the first thing to optimize.
- The watch cadence is per data, not global: a position may need a tight cadence while a name needs almost none. One global rate is either too chatty or too laggy.
- Replicating to every connected node instead of only interested ones scales bandwidth with players × objects; gate by relevancy.
- Change detection only sees what it was told to watch; a field mutated outside the watched component (or via an aliased pointer) won't be diffed and won't replicate.
- Replicating an event by id assumes the receiver already has that object; if the object hasn't replicated yet the event has nothing to apply to — order object state before object-targeted events.

**Related:** [references/delta-and-snapshots.md](./delta-and-snapshots.md), [references/topology-and-authority.md](./topology-and-authority.md), **ecs-guide**, **data-model-guide**, **data-oriented-design-guide**
