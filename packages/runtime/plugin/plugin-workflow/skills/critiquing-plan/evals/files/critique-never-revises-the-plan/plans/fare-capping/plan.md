---
type: Workflow Plan
schema: workflow/plan/1
title: Fare capping
description: Cap a rider's daily fares so no rider pays more than the daily price.
generated: {by: writing-plan/0.1.0, at: 2026-10-20T10:00:00+02:00}
owners:
  - Pricing team
repositories:
  - name: pricing-service
    resource: https://git.example.org/pricing-service
---

# Objective

A rider's daily fares never exceed the daily price, and the app shows the capped total.

# Scope

In: the pricing service and the app total. Out: season tickets.

# Success criteria

- The capped total is correct.
- Performance is acceptable.
