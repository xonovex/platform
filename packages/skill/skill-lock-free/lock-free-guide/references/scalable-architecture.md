# scalable-architecture: Scalable Architecture

**Guideline:** Scale by _removing sharing_, not by making shared access faster — the fastest synchronization is none. Partition state per-thread/per-core, and only synchronize on the rare path.

**Rationale:** Any memory written by multiple cores costs a coherence round-trip, and that cost is fixed no matter how clever the lock-free trick. Past a few cores, the only thing that scales is data that isn't shared. Algorithmic structure beats micro-optimized atomics: an architecture that avoids contention with a plain mutex outperforms a lock-free one that contends.

**Techniques:**

- **Per-thread / per-core shards:** give each thread its own counter, allocator arena, or queue; aggregate lazily when a total is actually needed (sum the per-thread counters at frame end). Turns a contended atomic into a contention-free local increment.
- **Partitioning / sharding:** split a shared map or array into N independent partitions keyed by hash; a thread touches only its partition, so writers to different keys never contend.
- **Combining / flat combining:** under heavy contention, instead of every thread CAS-ing the structure, threads post their operations to a list and one thread (the current "combiner") applies a batch. Converts N cache-line bounces into one, and can be faster than lock-free at high core counts.
- **Work-stealing deques:** each worker owns a double-ended queue — it pushes/pops its _own_ end with no synchronization (the common case), and only uses atomics when _stealing_ from another worker's far end (the rare case). This is the standard scalable job-system scheduler.
- **Fiber-based jobs:** run each job on a fiber (a stackful cooperative coroutine). When a job must wait on a dependency, it _yields its fiber_ and the worker thread immediately picks up another runnable job instead of blocking — cores never idle on a stall, and there is no thread-pool oversubscription. Pairs naturally with work-stealing (steal fibers, not just tasks). This is the high-end engine scheduler design.
- **Asymmetric synchronization:** make the common path cheap and the rare path expensive — e.g. RCU/seqlock (free reads, costly writes), or a fast lock-free try followed by a slow blocking fallback. Optimize for the operation you do a million times, not the one you do once.
- **Amortization:** batch many operations behind one synchronization (drain a whole local buffer into a shared queue under one lock acquire instead of one lock per item).
- **NUMA-aware placement:** pin threads and allocate their hot data on the local NUMA node; cross-socket coherence is dramatically slower, so keep a worker's shard on its own socket.

**Example:**

```c
#define CACHE_LINE_SIZE 64 // platform cache line: 64 B x86-64, 128 B Apple silicon

// Sharded counter: hot path is a contention-free local increment.
typedef struct {
    _Alignas(CACHE_LINE_SIZE) atomic_uint_fast64_t value; // padded: no false sharing
    char pad[CACHE_LINE_SIZE - sizeof(atomic_uint_fast64_t)];
} shard_t;

shard_t shards[MAX_WORKERS];

void bump(int worker_id) {
    // relaxed: each shard is effectively thread-local; only the rare total reads it.
    atomic_fetch_add_explicit(&shards[worker_id].value, 1, memory_order_relaxed);
}

uint64_t total(void) {                 // cold path: pay the cross-core reads once
    uint64_t sum = 0;
    for (int i = 0; i < MAX_WORKERS; i++)
        sum += atomic_load_explicit(&shards[i].value, memory_order_relaxed);
    return sum;
}
```

**Gotchas:**

- Sharding trades exact-at-all-times reads for cheap writes; the aggregated total is only a snapshot, not a consistent instant. Fine for stats, wrong for invariants.
- Padding shards to a cache line is mandatory — an unpadded shard array recreates the very false sharing you sharded to avoid.
- Work-stealing's correctness lives entirely in the _steal_ path's atomics (the owner's pop and a thief's steal can race on the last element); use a verified deque, don't hand-roll it.

**Related:** [references/false-sharing.md](./false-sharing.md), [references/progress-guarantees.md](./progress-guarantees.md), [references/locks-and-backoff.md](./locks-and-backoff.md)
