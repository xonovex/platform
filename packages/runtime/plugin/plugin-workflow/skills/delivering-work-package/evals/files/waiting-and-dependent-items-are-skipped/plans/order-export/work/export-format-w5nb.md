---
type: Workflow Work Item
schema: workflow/work-item/1
title: Export file format
description: Write the export in the agreed file format.
generated: {by: defining-work-items/5.3.0, at: 2026-11-11T10:00:00Z}
kind: work-item
parent: work/order-export-package-k2mz
repository: shop-api
readiness: waiting
waits_on:
  - decisions/file-format-h6tc
depends_on:
  - work/export-model-q3va
criteria:
  - id: c1
    text: The export file opens in the agreed format.
    verification: Run npm test -- export-format.
---

# Objective

Write the export in the agreed file format.
