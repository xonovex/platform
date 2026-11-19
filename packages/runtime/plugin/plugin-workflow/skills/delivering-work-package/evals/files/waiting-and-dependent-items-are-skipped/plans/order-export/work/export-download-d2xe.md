---
type: Workflow Work Item
schema: workflow/work-item/1
title: Export download
description: Give the shop owner a link to download the export.
generated: {by: defining-work-items/5.3.0, at: 2026-11-11T10:00:00Z}
kind: work-item
parent: work/order-export-package-k2mz
repository: shop-api
readiness: ready
depends_on:
  - work/export-format-w5nb
criteria:
  - id: c1
    text: GET /exports/{id}/file returns the file of a finished export.
    verification: Run npm test -- export-download.
---

# Objective

Give the shop owner a link to download the export.
