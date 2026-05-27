# Sources

## Game-engine development blog (archive)

- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → all sections
  - The node-graph networking model, UDP transport with per-packet-type delivery guarantees, opt-in component/object replication, delta-vs-snapshot synchronization, and the in-process virtual-network debugging workflow
- **Aspects extracted:**
  - "The Network Frontier — Part 1" — nodes binding a simulation to an address (the internet as a graph of nodes), unidirectional pipes with ever-incrementing ids and a request/accept handshake, packet types carrying retransmission/ack policy, UDP with caller-chosen guaranteed-delivery modes, and the receiver/accepter/bootstrapper callback interfaces; plus the goals of flexible (non-fixed) topology, low single-player↔multiplayer friction, and automate-by-default/optimize-when-needed → `references/topology-and-authority.md`, `references/transport-and-channels.md`, `references/state-replication.md`, `references/dev-iteration-and-observability.md`
  - "The Network Frontier, Part 2" — per-node behavior assets declaring accept/own/replicate, two-level replication opt-in (per-component capability + per-object flag) with watch-timer change detection and whole-blob-vs-per-field-layout, per-frame change dumps (delta) vs full dumps (snapshot) for new connections reusing the save/load apply path, replicating events/variables by id with ordered-vs-unordered semantics, the topology-switch catch-all for single-player, and the network profiler (in-process virtual network, injected latency/bandwidth, ack/out-of-order packet inspection) → `references/state-replication.md`, `references/delta-and-snapshots.md`, `references/topology-and-authority.md`, `references/dev-iteration-and-observability.md`, `references/transport-and-channels.md`

## Real-time networking — general prior art

- **URLs:**
  - RFC 768, User Datagram Protocol — https://www.rfc-editor.org/rfc/rfc768
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Transport and channels, Gotchas
  - The baseline UDP semantics (no ordering or delivery guarantee) that the engine layers reliability on top of
- **Aspects extracted:**
  - UDP datagrams are unordered and unreliable; ordering/reliability are added by the application, motivating per-packet-type delivery policy and the head-of-line tradeoff → `references/transport-and-channels.md`

## Refresh Workflow

1. Re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
