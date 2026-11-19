---
type: Workflow Work Item
schema: workflow/work-item/1
title: Export audit record
description: Write an audit record for every export.
generated: {by: defining-work-items/5.3.0, at: 2026-11-11T10:00:00Z}
kind: work-item
repository: shop-api
readiness: ready
criteria:
  - id: c1
    text: Every export writes one audit record with the user and the period.
    verification: Run npm test -- export-audit.
---

# Objective

Write an audit record for every export.
