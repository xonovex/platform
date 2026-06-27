# command-handoff: Lock-Free Command Handoff to the Audio Thread

## Guideline

The game thread never touches voice state directly; it posts immutable commands (play, stop, set-gain, set-pitch, set-position) into a single-producer/single-consumer lock-free queue that the audio thread drains at the top of each render block.

## Rationale

The audio thread cannot take a lock the game thread might hold (it would risk a deadline-missing stall), and two threads mutating the same voice struct is a data race. A lock-free SPSC handoff resolves both: the game thread is the sole producer, the audio thread the sole consumer, ownership of voice state stays entirely on the audio thread, and the only cross-thread contract is the queue's. Commands are self-contained value messages, so no shared object is mutated from two sides. Draining the queue once per block — at a known point, before mixing — keeps command application deterministic and bounded. The queue mechanics (indices, acquire/release ordering, full/empty handling) are not re-derived here; that machinery is owned by lock-free-guide.

## How to Apply

1. Define a small tagged-union command type carrying everything the audio thread needs to act without dereferencing game-side memory (voice handle, target gain matrix, pitch, source pointer into a preallocated/immutable asset).
2. Use an SPSC ring sized for the worst-case burst of commands per block; the game thread pushes, the audio thread pops. (See lock-free-guide for the ring, indices, and memory ordering.)
3. At the start of each render block, drain all pending commands and apply them to the voice pool, then mix. Never apply commands mid-mix.
4. If the queue is full, drop or coalesce on the producer side (e.g. keep only the latest set-gain for a voice) rather than blocking — blocking the game thread is acceptable, blocking the audio thread never is, and an unbounded queue reintroduces allocation.
5. For data too large to copy into a command (a decoded clip, a long buffer), pass a pointer to immutable, preallocated memory that outlives the voice; the audio thread reads it but never frees it.

## Example

```c
// Self-contained commands; no pointer the audio thread must coordinate ownership of,
// except into immutable, preallocated asset memory.
typedef enum { CMD_PLAY, CMD_STOP, CMD_SET_GAIN, CMD_SET_PITCH } cmd_kind_t;

typedef struct {
  cmd_kind_t kind;
  voice_handle_t voice;
  union {
    struct { const source_t *source; float gain[MAX_IN][MAX_OUT]; uint16_t prio; } play;
    struct { float gain[MAX_IN][MAX_OUT]; } set_gain; // becomes the ramp target
    struct { float pitch; } set_pitch;
  };
} audio_cmd_t;

// Audio thread: drain, apply, then mix. spsc_pop is the SPSC ring from lock-free-guide.
static void apply_commands(voice_pool_t *pool, spsc_ring_t *cmds) {
  audio_cmd_t c;
  while (spsc_pop(cmds, &c)) {
    switch (c.kind) {
      case CMD_PLAY:     voice_start(pool, &c); break;
      case CMD_STOP:     voice_request_stop(pool, c.voice); break; // ramp to zero
      case CMD_SET_GAIN: voice_set_gain_target(pool, c.voice, c.set_gain.gain); break;
      case CMD_SET_PITCH:voice_set_pitch(pool, c.voice, c.set_pitch.pitch); break;
    }
  }
}
```

## Gotchas

- Commands must be value-complete; a command that points at game-thread-mutable state reintroduces the data race the queue was meant to remove.
- A full queue must not block the consumer (audio thread); decide the producer-side policy up front (drop, coalesce, or grow a preallocated bound) — never `malloc` a bigger queue from either thread.
- Applying commands mid-block makes a voice's parameters change partway through the samples, smearing the ramp; drain to a clean point, then mix.
- A "set parameter" command should set the ramp target, not the live value — applying it as an instant change clicks (see mixing-and-buffers).
- This is strictly one producer and one consumer; if multiple game threads can post audio commands, you need an MPSC queue, not SPSC (see lock-free-guide), or funnel them through one thread first.
- A stale voice handle (the slot was recycled) must be detected by generation and ignored; otherwise a late set-gain hits an unrelated new sound.

## Related

[references/voice-management.md](./voice-management.md), [references/audio-callback-thread.md](./audio-callback-thread.md), **lock-free-guide**, **memory-management-guide**
