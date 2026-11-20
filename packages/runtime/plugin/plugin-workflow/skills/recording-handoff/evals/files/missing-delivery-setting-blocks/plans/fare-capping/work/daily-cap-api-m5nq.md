---
type: Workflow Work Item
schema: workflow/work-item/1
title: Daily cap API
description: Expose the daily cap in the pricing API.
generated: {by: defining-work-items/0.1.0, at: 2026-10-02T10:00:00+01:00}
kind: work-item
team: pricing
repository: shop-app
readiness: ready
criteria:
  - id: c1
    text: The API returns the cap for a rider.
    verification: Contract test against the pricing API.
---

# Objective

Expose the daily cap so the app can show it.
