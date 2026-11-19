---
type: Workflow Work Item
schema: workflow/work-item/1
title: Capped total API
description: The pricing service caps daily totals.
generated: {by: defining-work-items/0.1.0, at: 2026-10-22T10:00:00+02:00}
verified: {by: human:pricing-owner, at: 2026-10-22T10:30:00+02:00}
kind: work-item
team: Pricing
repository: pricing-service
readiness: ready
depends_on:
  - work/fare-rules-model-3m4q
decisions:
  - decisions/round-reduced-down-4h2k
criteria:
  - id: c1
    text: cap_daily_total returns the sum of the fares below the daily price and the daily price otherwise.
    verification: Run `python3 -m unittest` in the worktree and see the cap tests pass.
  - id: c2
    text: apply_reduction returns the reduced total rounded down to whole cents.
    verification: Run `python3 -m unittest` in the worktree and see the reduction test pass.
  - id: c3
    text: export_capped_totals returns a header line and one line for each day with its capped total.
    verification: Run `python3 -m unittest` in the worktree and see the export test pass.
  - id: c4
    text: The app shows the capped total on a physical device.
    verification: Open the app on a device and capture a screenshot of the total.
---

# Objective

Riders see a daily total that never exceeds the daily price.
