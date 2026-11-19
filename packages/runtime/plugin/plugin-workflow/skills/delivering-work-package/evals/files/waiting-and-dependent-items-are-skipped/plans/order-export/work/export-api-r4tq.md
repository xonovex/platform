---
type: Workflow Work Item
schema: workflow/work-item/1
title: Export API
description: Start an export through the API.
generated: {by: defining-work-items/5.3.0, at: 2026-11-11T10:00:00Z}
kind: work-item
parent: work/order-export-package-k2mz
repository: shop-api
readiness: ready
depends_on:
  - work/export-model-q3va
criteria:
  - id: c1
    text: POST /exports returns 202 with the id of the export.
    verification: Run npm test -- export-api.
---

# Objective

Start an export through the API.
