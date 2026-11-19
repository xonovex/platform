# Implementation Result Example

```text
Continue (apply): plans/order-import-backpressure/01-reader-pull.md

Target: subplan 01, the first with unfinished tasks (02 untouched)
Baseline: app:test 214/214 before edits; lint clean
Implemented: tasks 1 to 5; revision f4e5d6
Verification: app:test 219/219; app:ci-check green; criterion 2's
  counter asserts zero drops under the spike fixture
Criteria: 3 of 4 met; criterion 4 (staging soak) needs the staging
  feed, recorded unmet with its blocker named
Remaining: subplan 02, not started, per the one-target rule
```

The report separates baseline from new results, ties each criterion to its evidence, and stops at one target.
