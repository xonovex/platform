# measurement-and-profiling: Measurement and Profiling

**Guideline:** You cannot optimize what you do not measure — drive every layout change with a profiler and hardware counters, comparing before/after on representative data.

**Rationale:** DOD claims are about the hardware (cache misses, prefetch, vectorization), and intuition about the hardware is unreliable: the bottleneck is frequently not where it "should" be, and a layout that looks faster can be slower. Worse, asymptotic complexity hides constant factors and memory behavior — an O(n log n) tree can lose to an O(n²) array scan at realistic n because the array streams while the tree chases pointers. Only measurement on real inputs distinguishes a real win from a plausible-sounding regression.

**Techniques:**

- **Profilers** - Sampling profilers (perf, VTune, Instruments) show where time goes; look for memory-stall-dominated functions, not just high self-time.
- **Hardware performance counters** - Read `cache-misses`, `LLC-load-misses`, `cycles`, `instructions` (IPC), `stalled-cycles-backend`, branch mispredicts. Falling cache-misses + rising IPC confirms a layout win.
- **Microbenchmarks** - Isolate the hot loop, run it on realistic sizes and data distributions, warm the cache or measure cold deliberately, and report a stable statistic (median of many runs).
- **Before/after discipline** - Measure the baseline, change one thing, measure again on the same input. Keep the harness fixed so the delta is attributable.
- **Big-O vs constants and memory** - Treat complexity as a guide, not a verdict; for cache-bound work the constant factor and access pattern dominate at realistic n.

**How to Apply:**

1. Establish a repeatable benchmark with representative data sizes and distributions.
2. Record baseline counters: time, cache-misses, IPC, branch-misses.
3. Make a single layout/access change (e.g. AoS→SoA, hot/cold split).
4. Re-measure; accept only if the target counters improved on the real workload — revert otherwise.

**Example:**

```sh
# Compare cache behavior of two builds on the same input (Linux perf).
perf stat -e cycles,instructions,cache-references,cache-misses,branch-misses \
  ./sim_aos  level.dat   # baseline: Array-of-Structs
perf stat -e cycles,instructions,cache-references,cache-misses,branch-misses \
  ./sim_soa  level.dat   # candidate: Struct-of-Arrays
# Win = fewer cache-misses AND higher instructions/cycle on sim_soa.
```

```c
// Microbenchmark skeleton: warm up, time many iterations, report median.
double bench(void (*run)(void *), void *ctx, int reps) {
  run(ctx); /* warm cache */
  double best = 1e30;
  for (int r = 0; r < reps; r++) {
    double t0 = now_seconds();
    run(ctx);
    double dt = now_seconds() - t0;
    if (dt < best) best = dt; // min = least noise
  }
  return best;
}
```

**Gotchas:**

- Benchmarking on tiny data that fits in L1 hides the very misses you are optimizing — use realistic working-set sizes.
- Compiler optimizations can elide a benchmark whose result is unused; consume the output (e.g. accumulate and print) to prevent dead-code elimination.
- A single run is noise; report a stable statistic and pin frequency scaling where possible.

**Related:** [references/cache-behavior.md](./cache-behavior.md), [references/soa-aos-aosoa.md](./soa-aos-aosoa.md), [references/simd-friendly-layout.md](./simd-friendly-layout.md), [references/data-as-transforms.md](./data-as-transforms.md)
