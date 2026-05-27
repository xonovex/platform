# dev-iteration-and-observability: Virtual Networking, Latency Simulation, and Packet Inspection

**Guideline:** Make multiplayer debuggable and iterable by abstracting the wire behind the network API so many nodes can run in one process and behave like remote machines, by simulating adverse conditions (added latency, bandwidth caps) on demand, by exposing a packet inspector for acks/order/counts, and by letting one project run single-player or multiplayer through a topology branch with no code changes.

**Rationale:** Multiplayer is hard to debug because the number of possible states is multiplied across several running processes that you can't pause together or step in one debugger. The fix is to make "remote" an abstraction, not a deployment fact: route all node-to-node traffic through one API, and when both nodes live in the same process, short-circuit the wire entirely — deliver to the receiver callback immediately without copying — so you can launch a server node and a client node side by side in one tool and step them together, yet the same code runs unchanged across real machines later. On top of that, real network pain (latency, jitter, limited bandwidth) is something you must be able to _inject_ to hit corner cases without standing up a lossy network, and you need a profiler that can pause every simulation at once and show what packets flew, how many were acked, and how many arrived out of order. Finally, the gap between single-player and multiplayer should be a data branch on node type — a "what node am I" switch whose unnamed/catch-all path runs every branch in sequence — so the single-player build is the multiplayer build with one node running all roles.

## Contents

- Virtual networking: in-process nodes that skip the wire
- Same code local and remote
- Injecting latency and bandwidth limits
- Packet inspection: acks, order, per-type counts
- Single-player ↔ multiplayer via a topology branch

**How to Apply:**

1. Route every node-to-node message through the network API; never let nodes touch each other's state directly, so the API can choose wire vs. in-process delivery.
2. Short-circuit local delivery: if the destination node is in the same process, call its receiver callback immediately and skip serialization/copies — identical behavior, zero wire cost.
3. Run multiple nodes in one process for development (e.g. a server tab and a client tab) and step them under one debugger; the code is byte-for-byte what ships across real machines.
4. Provide knobs to inject adverse conditions per pipe/node: artificial latency and upload/download bandwidth caps, so corner cases are reproducible without a real lossy network.
5. Build a packet inspector that can pause all simulations together and answer "what packets of type X were sent N frames ago," "how many were acked last frame," "how many arrived out of order."
6. Branch single-player vs. multiplayer on node type via a topology switch; give it a catch-all (no node name) path that runs every branch in sequence, so a single node can play all roles for the offline build.

**Example:**

```c
// One process, two nodes: local sends skip the wire entirely.
net_o *net = net_create();
net_node_o *server = net_create_node(net, 7777);
net_node_o *client = net_create_node(net, 7778);
// Dev-only knobs: reproduce bad networks without one.
net_pipe_set_added_latency_ms(net, c2s_pipe, 120);
net_node_set_bandwidth_caps(net, server, /*up*/ 256 * 1024, /*down*/ 0);

// Topology branch: server and client take different paths; a node with no
// name is "catch-all" and runs ALL branches in order -> single-player build.
switch (node_role(self)) {
    case ROLE_SERVER: run_server_systems(); break;
    case ROLE_CLIENT: run_client_systems(); break;
    case ROLE_NONE:   run_server_systems(); run_client_systems(); break;
}
```

**Gotchas:**

- In-process short-circuit must not copy or re-serialize, or you measure the wrong thing and hide serialization bugs that only bite over a real socket; but skipping serialization can also hide format errors — test against a real socket before shipping.
- Latency/bandwidth injection belongs in a debug layer, never in shipping code paths; leaving it on tanks live performance.
- Pausing "the simulation" must pause every node's simulation instance together, or the inspector shows an inconsistent cross-node snapshot.
- A catch-all single-player path that runs every branch in sequence assumes the branches don't both mutate the same authoritative state in conflicting ways; ordering matters.
- Local-only testing can pass while the shipped build fails: in-process delivery never drops or reorders, so bugs that only appear under loss/reorder stay hidden until you inject conditions or use a real socket.
- Per-type packet counts and ack/out-of-order stats are only as good as the packet typing (see transport-and-channels.md); untyped or mistyped packets distort the profiler.

**Related:** [references/transport-and-channels.md](./transport-and-channels.md), [references/topology-and-authority.md](./topology-and-authority.md), **ecs-guide**
