# voice-management: Voice Pools, Lifetime, and Stealing

**Guideline:** Play every sound through a voice drawn from a fixed-size, preallocated pool; cap concurrency at the pool size, steal the least important voice when the pool is full, and ramp a stopped voice to silence before returning its slot.

**Rationale:** A fixed pool is what makes the render loop's worst case bounded and the audio thread allocation-free: the inner mix loop iterates at most `MAX_VOICES` times no matter how many plays the game requests, so the per-block cost has a hard ceiling. Unbounded voices would mean unbounded mix time (deadline misses) and runtime allocation (forbidden on the audio path). When the game asks for one more sound than the pool holds, you must choose what loses — silently dropping the new request or cutting an old voice — and "least important / quietest / oldest" is a better default than refusing to play. Because cutting a voice instantly clicks, the slot is not free the moment you stop it; it frees once its gain has ramped to zero.

**How to Apply:**

1. Allocate `MAX_VOICES` voices once at init. A "play" returns a handle (slot index + generation counter), not a pointer, so a stale handle to a recycled slot is detectable.
2. Store voice state struct-of-arrays for the mix loop's hot fields (read cursor, gain matrix, source pointer, flags), so the inner loop streams cache-friendly columns rather than chasing fat structs (see data-oriented-design-guide).
3. On play: find a free slot; if none, pick a steal victim by priority (lowest priority, then quietest, then oldest) and reuse it.
4. On stop: set the voice's target gain matrix to zero and let it ramp down; mark the slot reclaimable only after the gain reaches zero, then bump its generation so old handles go stale.
5. Each block, retire voices that finished their source or completed a stop-ramp; return their slots to the free set.

**Example:**

```c
#define MAX_VOICES 256

// Hot fields the mix loop touches every block, laid out struct-of-arrays.
typedef struct {
  double   cursor[MAX_VOICES];        // fractional read position
  float    gain[MAX_VOICES][MAX_IN][MAX_OUT]; // live gain matrix (ramped)
  float    gain_target[MAX_VOICES][MAX_IN][MAX_OUT];
  uint16_t priority[MAX_VOICES];
  uint8_t  active[MAX_VOICES];        // 0 = free, 1 = playing, 2 = stopping
  uint32_t generation[MAX_VOICES];    // bumped on free; encoded into handles
  const source_t *source[MAX_VOICES];
} voice_pool_t;

typedef struct { uint16_t slot; uint16_t gen; } voice_handle_t;

// Allocate a slot, stealing the lowest-priority active voice if the pool is full.
static int32_t voice_alloc(voice_pool_t *p, uint16_t prio) {
  int32_t victim = -1;
  for (int32_t i = 0; i < MAX_VOICES; ++i) {
    if (p->active[i] == 0) return i;                 // free slot wins
    if (victim < 0 || p->priority[i] < p->priority[victim]) victim = i;
  }
  return (p->priority[victim] <= prio) ? victim : -1; // steal only if not more important
}
```

**Gotchas:**

- Stealing a voice by overwriting it instantly clicks; ramp the victim's gain to zero (a quick fade) before repurposing the slot, or accept the pop.
- A "stopped" voice still occupies its slot until its ramp completes, so the effective free count lags the logical one; size the pool for the ramp overlap, not just steady state.
- Returning a pointer into the pool invites use-after-recycle; hand out an opaque handle with a generation counter and validate it before every operation.
- Stealing the wrong voice (e.g. always the oldest) cuts important long sounds like music under a flurry of footsteps; rank by explicit priority first, loudness/age only as tiebreakers.
- Off-thread "play" requests must arrive via the command queue, not by mutating the pool directly — the audio thread owns voice state (see command-handoff).

**Related:** [references/command-handoff.md](./command-handoff.md), [references/mixing-and-buffers.md](./mixing-and-buffers.md), **data-oriented-design-guide**, **memory-management-guide**
