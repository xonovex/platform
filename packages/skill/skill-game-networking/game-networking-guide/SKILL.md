---
name: game-networking-guide
description: "Use when architecting real-time multiplayer networking for a game or simulation engine: client/server vs peer topology and per-object authority, replicating a typed world/component state across nodes, snapshot-vs-delta updates with baselines and acks, and reliable/unreliable/ordered channels layered over UDP. Triggers on netcode, network nodes, pipes/connections, packet types, state replication, snapshots and deltas, server authority, ack/out-of-order packets, lag/bandwidth simulation, and single-player-to-multiplayer transitions, even when the user doesn't say 'networking'. Skip in-engine data layout (use data-oriented-design-guide), the component model itself (use ecs-guide), the typed object/asset model (use data-model-guide), and GPU/rendering work (use gpu-rendering-guide)."
---

# Game networking Guidelines

API-agnostic architecture for real-time multiplayer over an unreliable network: how participants are modeled and who owns state, what to put on the wire, and how reliability is layered on UDP. This skill replicates a typed world — see **data-model-guide** for the object/id model, **ecs-guide** for the component model, and **data-oriented-design-guide** for packing wire data.

## Essentials

- **Nodes form a graph, topology is data** - Model each participant as a simulation bound to an address; keep client/server-vs-peer and accept/own/replicate as per-node config, not baked-in protocol, see [references/topology-and-authority.md](references/topology-and-authority.md)
- **One owner per object** - Exactly one node is the source of truth for a networked object and replicates it; others apply received updates read-only, see [references/topology-and-authority.md](references/topology-and-authority.md)
- **Replicate opted-in state** - Replication is a per-component capability plus a per-object flag; the owner watches for changes and sends only to interested nodes, see [references/state-replication.md](references/state-replication.md)
- **Delta to the in-sync, snapshot to the new** - Stream per-frame changes to nodes that have a baseline; send a full snapshot to a freshly connected node, see [references/delta-and-snapshots.md](references/delta-and-snapshots.md)
- **Guarantee per packet type, over UDP** - Open one-way pipes between nodes; tag every payload with a type that carries its delivery guarantee (unreliable / ordered / reliable), see [references/transport-and-channels.md](references/transport-and-channels.md)

## Replicating world state

- **Two-level opt-in** - A component declares whether/how it replicates; each object must also be flagged, or nothing is sent, see [references/state-replication.md](references/state-replication.md)
- **Cheap default, optimize by description** - Default to a whole-component blob; supply a field layout and per-field watch cadence to send only changed members, see [references/state-replication.md](references/state-replication.md)
- **Baselines and acks** - A delta is meaningless without the receiver's baseline; reuse the save/load apply path and watch ack/out-of-order counts, see [references/delta-and-snapshots.md](references/delta-and-snapshots.md)

## Transport and channels

- **Pipes are unidirectional** - A pipe is a one-way A→B path opened by a request/accept handshake; bidirectional talk needs a pipe each way, see [references/transport-and-channels.md](references/transport-and-channels.md)
- **Reliability is per-message policy** - Send by pipe + type + pointer + size; the type fixes retransmission and ack behavior so cheap data isn't head-of-line-blocked behind must-arrive data, see [references/transport-and-channels.md](references/transport-and-channels.md)

## Iteration and observability

- **Make remote an abstraction** - Route all traffic through the API so in-process nodes skip the wire and behave like remote ones; the shipped code is unchanged, see [references/dev-iteration-and-observability.md](references/dev-iteration-and-observability.md)
- **Inject bad networks, inspect packets** - Simulate latency and bandwidth caps on demand and inspect acks/order/counts; branch single-player vs multiplayer on node type, see [references/dev-iteration-and-observability.md](references/dev-iteration-and-observability.md)

## Gotchas

- A component being replicable does nothing unless the object instance is also flagged for replication — the most common "nothing syncs" bug is a missing object flag.
- Choosing reliable-ordered delivery for everything recreates TCP head-of-line blocking over UDP: one dropped packet stalls everything queued behind it.
- A delta applied without its baseline silently corrupts the receiver's world; a newly connected node must get a full snapshot before any delta.
- Pipes are one-way — forgetting the return pipe means the other node literally cannot reply.
- The connection handshake must be idempotent: a re-sent request (because the response was lost) has to return "accepted," not open a second pipe.
- In-process local delivery never drops or reorders, so loss/reorder bugs stay hidden until you inject adverse conditions or test against a real socket.
- "Deliver all, allow duplicates" guarantees arrival, not uniqueness — handlers must tolerate seeing the same packet twice.
- Two nodes that both believe they own the same object will fight; authority must be singular per object.

## Progressive Disclosure

- Read [references/topology-and-authority.md](references/topology-and-authority.md) - Load when deciding client/server vs peer, what each node owns, or who has authority over an object
- Read [references/state-replication.md](references/state-replication.md) - Load when choosing what state to replicate, change detection, relevancy, or replicating events/variables
- Read [references/delta-and-snapshots.md](references/delta-and-snapshots.md) - Load when designing per-frame delta updates, full snapshots for new connections, baselines, or acks
- Read [references/transport-and-channels.md](references/transport-and-channels.md) - Load when designing the UDP transport, pipes/connections, packet types, or reliable/unreliable/ordered delivery
- Read [references/dev-iteration-and-observability.md](references/dev-iteration-and-observability.md) - Load when setting up in-process testing, simulating latency/bandwidth, packet inspection, or single-player/multiplayer transitions
