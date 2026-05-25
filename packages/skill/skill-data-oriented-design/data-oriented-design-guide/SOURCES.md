# Sources

## Mike Acton — "Data-Oriented Design and C++" (CppCon 2014)

- **URL:** https://www.youtube.com/watch?v=rX0ItVEVjHc
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Processing
  - The "where there is one, there are many", transforms, and "the problem is data movement" framing
- **Aspects extracted:**
  - The memory wall / latency premise → `references/cache-behavior.md`
  - Data as bulk transforms, design from the data → `references/data-as-transforms.md`
  - Existence-based processing, branches→loops → `references/existence-based-processing.md`

## Richard Fabian — "Data-Oriented Design" (book)

- **URL:** https://www.dataorienteddesign.com/dodbook/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Layout, Memory
  - Tables/streams mental model, existence-based processing, "the absence of a thing is a thing"
- **Aspects extracted:**
  - AoS / SoA / AoSoA layouts → `references/soa-aos-aosoa.md`
  - Hot/cold field splitting → `references/hot-cold-splitting.md`
  - Existence-based processing, condition tables → `references/existence-based-processing.md`
  - Tables, streams, schemas → `references/data-as-transforms.md`
  - Handles / relational references → `references/handles-and-indices.md`

## Noel Llopis — "Data-Oriented Design (Or Why You Might Be Shooting Yourself in the Foot With OOP)" and related articles

- **URL:** https://gamesfromwithin.com/data-oriented-design
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Essentials, Layout
  - The practical OOP-vs-DOD contrast and cache-locality motivation
- **Aspects extracted:**
  - Cache lines, locality, miss types → `references/cache-behavior.md`
  - SoA for hot loops, layout selection → `references/soa-aos-aosoa.md`
  - Hot/cold splitting → `references/hot-cold-splitting.md`

## Supporting hardware / technique references

- **URLs:**
  - Ulrich Drepper, "What Every Programmer Should Know About Memory" — https://people.freebsd.org/~lstewart/articles/cpumemory.pdf
  - Agner Fog, optimization manuals — https://www.agner.org/optimize/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Memory, Gotchas
- **Aspects extracted:**
  - Cache hierarchy latencies, prefetching, miss taxonomy → `references/cache-behavior.md`, `references/access-patterns.md`
  - Alignment, SIMD loads, AoSoA → `references/simd-friendly-layout.md`
  - Profiling, hardware counters, microbenchmarking → `references/measurement-and-profiling.md`
  - Indices/handles, swap-remove, free lists → `references/handles-and-indices.md`

## Game-engine development blog (archive)

- **URL:** https://archive-host.github.io/blog_archive/
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Layout, Memory
  - Practical engine DOD storage and allocator techniques
- **Aspects extracted:**
  - "Data Structures Part 1–2: Bulk Data / Indices" — bulk-data store + handle↔index indirection (stable handles AND dense data) → `references/handles-and-indices.md`
  - "It's All About The Data" — data-first framing → `references/data-as-transforms.md`
  - "Minimalist container library in C" — contiguous growable containers (noted in c99-opinionated)

## Refresh Workflow

1. Re-watch/re-read the upstream source(s) above
2. Diff against the prior pull (or scan for newly added sections)
3. For each changed area, update the corresponding `references/<topic>.md`
4. Bump **Last reviewed** date above
