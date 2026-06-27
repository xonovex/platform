# transport-and-channels: UDP Transport, Pipes, Packet Types, and Delivery Guarantees

## Guideline

Build the wire layer on UDP and expose a small transport API where the caller opens a one-way pipe between two nodes, tags every payload with a packet type, and the packet type carries the delivery guarantee (none / ordered / reliable-deliver-all) — so reliability and ordering are per-message policy layered over UDP, not a single fixed channel.

## Rationale

Games need both fire-and-forget data (this frame's input, which is stale next frame) and must-arrive data (a snapshot, a connection handshake), and a single reliable-ordered stream like TCP head-of-line-blocks the cheap data behind the expensive data. UDP gives raw datagrams with no ordering or delivery promises; the engine adds the guarantees selectively on top. The unit of communication is a pipe: a unidirectional path from node A to node B identified by an ever-incrementing id, established by a request/response handshake the receiver can accept or reject; bidirectional talk needs a pipe each way. Every payload sent on a pipe is stamped with a packet type, and the type — not each send call — defines the retransmission policy and whether acknowledgements are required. Callers then send a typed payload with just a pointer and a size; the transport handles fragmentation, acks, and retransmission according to the type's guarantee. Receiving is callback-driven: a receiver callback handles an arriving packet, an accepter decides incoming pipe requests, a bootstrapper runs at node startup.

## Contents

- UDP base and why guarantees are layered on
- Pipes: unidirectional, ids, one-per-direction
- Handshake: request / accept / response
- Packet types and per-type delivery policy
- Delivery guarantees: unreliable / ordered / reliable
- Callback interfaces: receiver, accepter, bootstrapper

### How to Apply

1. Use UDP as the base transport; treat ordering and delivery as optional services you add per packet type, not properties of the socket.
2. Make a pipe the unit of communication: unidirectional A→B, identified by an ever-incrementing id so two pipes from the same node never collide. Open a second pipe for B→A.
3. Open a pipe with a handshake: sender sends a pipe request with the chosen id; the receiver's accepter callback decides accept/reject; the receiver replies with the id and an accepted flag. A duplicate request for an already-accepted pipe returns accepted immediately (idempotent).
4. Define packet types up front; each type fixes a delivery guarantee. Send by giving the pipe, the type, a payload pointer, and a size — nothing else.
5. Choose the guarantee per type: unreliable (may drop/reorder — for stale-able data), ordered (delivered in send order), or reliable deliver-all (every packet arrives, duplicates allowed). Reserve reliable for data that must land (handshakes, snapshots, terminal events).
6. Drive receipt with callbacks: a receiver callback returns whether it handled the packet, a bootstrapper runs once after node creation, an accepter gates incoming pipe requests. Register multiple interfaces per node as needed.

### Example

```c
// Packet TYPE carries the delivery guarantee, not the send call.
uint64_t pkt = hash_string("player_input");
net_define_packet(net, pkt, NET_DELIVERY_UNRELIABLE); // stale next frame; ok to drop

uint64_t snap = hash_string("state_snapshot");
net_define_packet(net, snap, NET_DELIVERY_RELIABLE_ALL); // must arrive

// A pipe is one-way; opening it runs a request/accept/response handshake.
net_node_o *node = net_create_node(net, 7777);
net_pipe_id pipe = net_open_pipe(net, node, &(net_addr){.ip = "127.0.0.1", .port = 7778});

// Send = pipe + type + pointer + size. Transport applies the type's guarantee.
char input[8]; fill_input(input);
net_send(net, pipe, pkt, input, sizeof input);

// Receipt is callback-driven; return true if handled.
static bool on_packet(net_o *net, void *ctx, net_pipe_id pipe,
                      uint64_t type, const void *data, uint32_t size) {
    return dispatch(type, data, size);
}
```

### Gotchas

- A pipe is unidirectional: forgetting the return pipe means B literally cannot answer A; you need one pipe per direction.
- Picking reliable-ordered for everything recreates TCP's head-of-line blocking over UDP — a dropped reliable packet stalls everything behind it; reserve it for data that truly must arrive in order.
- Unreliable packets may arrive out of order or never; never use them for state a receiver must converge to (snapshots, handshakes) — use a reliable type.
- The handshake must be idempotent: a re-sent pipe request (because the response was lost) has to return accepted, not open a second pipe.
- Pipe ids must be monotonic per source node; reusing an id risks binding new data to a stale pipe on the receiver.
- A receiver callback that returns "not handled" for an unknown type silently drops data — make unhandled types loud in development.
- "Deliver all, allow duplicates" guarantees arrival, not uniqueness; the handler must tolerate seeing the same packet twice.

### Related

[references/delta-and-snapshots.md](./delta-and-snapshots.md), [references/topology-and-authority.md](./topology-and-authority.md), **data-oriented-design-guide**
