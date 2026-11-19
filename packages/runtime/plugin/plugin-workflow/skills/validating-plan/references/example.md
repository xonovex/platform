# Validation Report Example

```text
Validation: order-import-backpressure at revision f4e5d6 (2026-01-15)

Checks: app:ci-check green; app:test-integration green; policy n/a
Criterion 1, zero drops under the spike fixture: pass; the counter is
  asserted in batcher.test.ts:88 at this revision
Criterion 2, both producers covered: pass; replay fixture added
Criterion 3, p95 latency under 200 ms: fail; measured 240 ms on the
  spike fixture, bench output attached; the plan is not changed
Criterion 4, staging soak: blocked; the staging feed is still absent
Definition of Done: docs updated pass; changelog entry missing, fail
Result: 2 pass, 1 fail, 1 blocked, with the evidence above
```

Every criterion gets its own verdict with reproducible evidence and the exact revision checked; a failing number is reported as measured.
