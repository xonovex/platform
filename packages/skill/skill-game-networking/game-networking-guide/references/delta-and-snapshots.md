# delta-and-snapshots: Delta Changes vs. Full Snapshots

## Guideline

Stream per-frame deltas of changed state to nodes already in sync, and send a full snapshot only when a node first becomes interested (a fresh connection) and so has no baseline to diff against; reuse the same change-tracking and load path the engine already uses for save/load.

## Rationale

Sending the whole world every frame is bandwidth suicide; sending only what changed since the receiver's last known state is the core bandwidth win of networked simulation. But a delta is meaningless without a baseline — the receiver must already hold the state the delta is relative to. A node that just connected has no baseline, so it needs the complete state once (a snapshot / full dump) before incremental deltas make sense. The clean way to get both is to lean on the engine's existing change-tracking system: each frame, propagate watched changes into a tracked state, dump the uncompressed changes since last frame to a buffer, and send that buffer; for a new connection, dump _all_ state instead. The receiver applies the buffer through exactly the same path that loads a saved game, so replication and persistence share one serialization and one apply step rather than maintaining two. Packet-level acknowledgements and out-of-order counts (surfaced by tooling) tell you whether deltas are actually landing.

## Contents

- Baseline requirement: why deltas need prior state
- Per-frame change dump (delta) vs. full dump (snapshot)
- Triggering a snapshot on new connections
- Sharing the save/load apply path
- Acks and out-of-order accounting

### How to Apply

1. Each frame on the owner: propagate watched component/object changes into the engine's tracked game state.
2. Dump the changes accumulated since last frame into one or more buffers (the delta) and send to interested nodes.
3. On a _new_ connection, dump the entire tracked state (the snapshot) and send it first — the new node has no baseline, so a delta would be undefined.
4. Receivers apply incoming buffers through the same code path that loads a saved game from disk; the world applies them like any other state load.
5. Treat the snapshot as the baseline the subsequent deltas are diffed against; don't interleave a node's first delta before its snapshot has been applied.
6. Use packet acknowledgements to learn what the receiver has confirmed, and watch out-of-order/unacked counts to detect a node falling behind.

### Example

```c
// Per frame on the owning node:
gamestate_propagate_changes(world);         // watched edits -> tracked state
buffer_t delta = gamestate_dump_changes(world); // only what changed last frame
for (node_o *n = interested; n; n = n->next)
    net_send(net, n->pipe, PKT_STATE_DELTA, delta.data, delta.size);

// On a brand-new connection (no baseline yet): send the whole thing once.
buffer_t snapshot = gamestate_dump_all(world);
net_send(net, joiner->pipe, PKT_STATE_SNAPSHOT, snapshot.data, snapshot.size);

// Receiver: same apply path as loading a save game.
void on_state(buffer_t buf) { gamestate_load_changes(world, &buf); }
```

### Gotchas

- A delta applied without its baseline corrupts the receiver's world silently; always deliver and apply the snapshot first for a new node.
- Reusing the save/load path is a feature, not a coincidence — but it means a change to the on-disk format is also a wire-format change; version them together.
- A full snapshot can be large; sending it as one unreliable datagram will fail — it needs reliable, possibly fragmented delivery (see transport-and-channels.md).
- "Changes since last frame" assumes every prior delta arrived; over an unreliable channel a dropped delta means the receiver's baseline drifts unless deltas are delivered reliably or periodically re-based.
- Acks confirm receipt, not application order; a high out-of-order count means deltas are arriving but possibly being applied in the wrong sequence.
- Dumping all changes every frame regardless of relevancy defeats the point — gate by which nodes are interested (see state-replication.md).

### Related

[references/state-replication.md](./state-replication.md), [references/transport-and-channels.md](./transport-and-channels.md), **data-oriented-design-guide**, **data-model-guide**
