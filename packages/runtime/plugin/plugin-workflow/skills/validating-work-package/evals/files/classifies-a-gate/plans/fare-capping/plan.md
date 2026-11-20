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
  - name: shop-app
    resource: https://git.example.org/shop-app
---

# Objective

A rider's daily fares never exceed the daily price, and the app shows the capped total.

# Scope

In: the pricing service and the app total. Out: season tickets.

# Success criteria

- The capped total of three trips in one day equals the daily price (check: `sh checks/pricing-rule.sh`).
- The app shows the capped total on the trip list (check: `sh checks/app-total.sh`).
- The new daily price applies in production only after the tariff authority approves it in the ticketing portal.
