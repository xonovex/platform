# Commonality and Variability

At each axis, keep the contract and rules true for every variant at the root; keep each differing answer in its own leaf.

```text
destination/
  shared/          # data contract and policy every destination uses
  file/            # only file behavior
  object-store/    # only object-storage behavior
```

Use these placement tests:

- Every variant needs it unchanged → axis `shared/`.
- Only one variant needs it → that variant’s leaf.
- A subset needs similar code → tolerate duplication until the common rule is stable; do not make unrelated variants depend on it.
- The root names a concrete variant or branches on its name → variability leaked upward.
- One leaf imports a sibling → either move a genuinely universal contract to `shared/` or revisit the axis.

When a shared abstraction accumulates ignored flags and special cases, inline it back into the leaves, let them diverge, and re-extract only the rule that remains common.
