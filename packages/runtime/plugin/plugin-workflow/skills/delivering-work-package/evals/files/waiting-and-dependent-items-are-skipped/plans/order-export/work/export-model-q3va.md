---
type: Workflow Work Item
schema: workflow/work-item/1
title: Export model
description: Build the export of the orders of one period.
generated: {by: defining-work-items/5.3.0, at: 2026-11-11T10:00:00Z}
kind: work-item
parent: work/order-export-package-k2mz
repository: shop-api
readiness: ready
criteria:
  - id: c1
    text: An export holds every order of the period and no other order.
    verification: Run npm test -- export-model.
  - id: c2
    text: Building the export of 10000 orders takes under 200 ms.
    verification: Run npm run bench -- export-model and read the p95.
---

# Objective

Build the export of the orders of one period.
