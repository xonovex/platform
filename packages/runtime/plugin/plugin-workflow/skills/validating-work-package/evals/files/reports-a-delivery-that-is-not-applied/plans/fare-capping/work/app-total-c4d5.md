---
type: Workflow Work Item
schema: workflow/work-item/1
title: Capped total in the app
description: Capped total in the app.
generated: {by: defining-work-items/0.1.0, at: 2026-10-20T10:00:00+02:00}
kind: work-item
team: Pricing team
readiness: ready
repository: shop-app
depends_on:
  - work/pricing-rule-a2b3
criteria:
  - id: c1
    text: The trip list shows the capped total.
    verification: Run sh checks/app-total.sh and read the total.
---

# Objective

Show the capped total in the app.
