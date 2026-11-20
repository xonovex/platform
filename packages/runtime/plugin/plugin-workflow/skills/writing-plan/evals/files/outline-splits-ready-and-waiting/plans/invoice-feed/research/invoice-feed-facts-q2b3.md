---
type: Workflow Research
schema: workflow/research/1
title: Invoice feed facts
description: Billing issues invoices in a nightly export that finance reads from a file drop
generated: {by: researching-change/0.1.0, at: 2026-10-20T09:00:00+02:00}
subject: Billing issues invoices in a nightly export that finance reads from a file drop
questions:
  - id: nightly-retention
    text: How long is the nightly file kept after the feed starts?
sources:
  - id: billing
    resource: billing-service@3fa91c2:src/export.ts
    title: Export job
---

# Bottom line

Billing writes one export file each night.[^billing]

# Observed

- The export job runs at 02:00.[^billing]

# Inference

None.

# Unknowns

None.

# Open decisions

| Id                  | Question                   | Options       | Recommendation |
| ------------------- | -------------------------- | ------------- | -------------- |
| `nightly-retention` | How long is the file kept? | 30 days; none | None           |

# Not checked

Runtime behavior was not run.

[^billing]: Export job, lines 10 to 40
