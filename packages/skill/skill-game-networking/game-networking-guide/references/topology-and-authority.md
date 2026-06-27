# topology-and-authority: Network Nodes, Topology, and Authority

## Guideline

Model a multiplayer session as a graph of network nodes — each node a self-contained simulation bound to an address — and make the topology (who connects to whom, who owns state, who accepts connections) a per-node configuration value rather than a protocol baked into the engine; do not assume a single fixed client/server shape.

## Rationale

Different games want different shapes — authoritative client/server, peer-to-peer, dedicated server plus login server, listen server — and a runtime that hard-codes one shape forces every other game to fight it. By treating each participant as a node (a simulation instance plus an address and a behavior description), the same primitives express every topology: a node's behavior asset declares whether it accepts incoming connections, which simulation systems it runs, and whether its state changes are sent outward. Authority then becomes "which node owns and replicates a given object," not a special case in the transport. This keeps the wire layer generic and pushes game-specific topology decisions to data the user controls, so moving from client/server to peer-to-peer is a configuration change, not a rewrite.

## How to Apply

1. Give each running simulation its own isolated world/context; never share mutable world state between nodes in-process — communicate only through the network API so local and remote behave identically.
2. Define a node as `(simulation instance, address, behavior description)`; bind it to a host/port at creation so it can be addressed by other nodes.
3. Put the topology in the behavior description, not in engine code: does this node accept connection requests, which systems does it simulate, does it replicate its changes outward.
4. Ship sensible default node behaviors (e.g. a server profile and a client profile) but let users author their own to express arbitrary topologies.
5. Decide authority per object/component: exactly one node owns a given networked object and is the source of truth that replicates it; other nodes treat their copy as read-only and apply received updates.
6. Keep the connection/accept decision at the node: an authoritative server accepts client connections; a peer may accept or reject based on its own policy.

## Example

```c
// A node = a simulation instance + an address + a behavior description.
typedef struct net_node_desc {
    const char *behavior;   // "server" | "client" | custom topology profile
    bool        accept_connections;
    bool        replicate_changes; // does this node push its state outward?
    uint16_t    port;
} net_node_desc;

net_node_o *server = net_create_node(net, &(net_node_desc){
    .behavior = "server", .accept_connections = true,
    .replicate_changes = true, .port = 7777});

net_node_o *client = net_create_node(net, &(net_node_desc){
    .behavior = "client", .accept_connections = false,
    .replicate_changes = false, .port = 0});
// Authority: `server` owns gameplay objects and replicates them; `client`
// applies received updates and owns only its local input.
```

## Gotchas

- A node is not a process: many nodes can run inside one executable for testing, so never assume "remote" means "another machine" — see dev-iteration-and-observability.md.
- Hard-coding the client/server split into transport or replication code blocks every non-standard topology (P2P, multiple server roles) later; keep the shape in node configuration.
- Two nodes with `replicate_changes = true` that both think they own the same object will fight; authority must be singular per object.
- A node that accepts connections but runs no simulation systems is a valid relay/login role — don't assume every node simulates gameplay.
- Sharing a world context between two in-process nodes "to save copies" silently couples them and breaks the illusion that they are separate machines.

## Related

[references/state-replication.md](./state-replication.md), [references/dev-iteration-and-observability.md](./dev-iteration-and-observability.md), **ecs-guide**, **data-model-guide**
