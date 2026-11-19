---
type: Workflow Work Item
schema: workflow/work-item/1
title: Capping rule in the pricing service
description: Capping rule in the pricing service.
generated: {by: defining-work-items/0.1.0, at: 2026-10-20T10:00:00+02:00}
kind: work-item
team: Pricing team
readiness: ready
repository: pricing-service
criteria:
  - id: c1
    text: Three trips in one day total the daily price.
    verification: Run sh checks/pricing-rule.sh and read the total.
---

# Objective

Cap the daily total in the pricing service.
