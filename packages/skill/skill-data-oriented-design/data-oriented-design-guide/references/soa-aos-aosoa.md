# soa-aos-aosoa: AoS vs SoA vs AoSoA

**Guideline:** Choose the layout by how the hot loop accesses fields — AoS when you touch most fields of one record at a time, SoA when you stream one or few fields across many records, AoSoA when you need both cache and SIMD efficiency.

**Rationale:** The three layouts store identical logical data but produce different memory traffic. A cache line is wasted on fields a loop does not read; the right layout maximizes useful bytes per loaded line and feeds the prefetcher and SIMD units. The choice is workload-specific, not universal.

**The three layouts:**

- **AoS (Array of Structs)** - `record[i].field`. Natural, one allocation, great when a loop reads many fields of the same record (random single-entity access, gameplay logic). Bad for streaming one field: each line drags the other fields.
- **SoA (Struct of Arrays)** - `field[i]`, one array per field. Each hot field is contiguous, so a loop touching a subset of fields wastes no line and the prefetcher streams perfectly. Best for bulk transforms over a few fields (physics integration, culling). Cost: full-record access touches N arrays; insert/remove must keep arrays in sync.
- **AoSoA (Array of Structs of Arrays, "tiled"/hybrid)** - Array of small blocks, each block being SoA over a fixed tile width W (often the SIMD width). `data[block].field[lane]`. Keeps a whole tile in a few lines (cache-friendly) while exposing contiguous lanes per field (SIMD-friendly). Best when you want both and can tolerate the extra index math.

**How to Apply:**

1. Identify the hot loop and the exact fields it reads/writes.
2. If it touches one/few fields across many elements → SoA. If it touches most fields of one element → AoS. If it does both, or you vectorize → AoSoA with W = SIMD lane count.
3. For SoA/AoSoA, pad arrays to the tile/SIMD width so the tail iteration stays branch-free.
4. Watch struct padding: in AoS a `char` between two `double`s can cost 7 bytes of padding per record; SoA columns have none.

**Example:**

```c
#define N 4096
#define W 8 // tile width = SIMD lanes

// AoS: good for "do everything to entity i"
typedef struct { float x, y, z; float vx, vy, vz; int hp; } entity_t;
entity_t aos[N];

// SoA: good for "integrate every position" (loop touches x..z, vx..vz only)
typedef struct { float x[N], y[N], z[N], vx[N], vy[N], vz[N]; int hp[N]; } soa_t;
soa_t soa;
for (size_t i = 0; i < N; i++) soa.x[i] += soa.vx[i]; // pure stream, no waste

// AoSoA: tiles of W lanes — cache-local block, SIMD-contiguous lanes
typedef struct { float x[W], y[W], z[W], vx[W], vy[W], vz[W]; } block_t;
block_t aosoa[N / W];
for (size_t b = 0; b < N / W; b++)
  for (size_t l = 0; l < W; l++)          // inner loop vectorizes
    aosoa[b].x[l] += aosoa[b].vx[l];
```

**Gotchas:**

- SoA is not free: more arrays to allocate, parallel indices to keep consistent, worse for whole-record random access.
- Do not convert to SoA on faith — verify the loop actually reads a subset of fields, or it can be a pessimization.

**Related:** [references/hot-cold-splitting.md](./hot-cold-splitting.md), [references/simd-friendly-layout.md](./simd-friendly-layout.md), [references/cache-behavior.md](./cache-behavior.md), [references/access-patterns.md](./access-patterns.md)
