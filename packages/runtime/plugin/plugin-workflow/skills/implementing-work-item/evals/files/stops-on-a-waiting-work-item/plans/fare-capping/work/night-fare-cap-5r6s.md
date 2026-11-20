---
type: Workflow Work Item
schema: workflow/work-item/1
title: Night fare cap
description: The cap treats night fares as the owners decide.
generated: {by: defining-work-items/0.1.0, at: 2026-10-22T10:00:00+02:00}
verified: {by: human:pricing-owner, at: 2026-10-22T10:30:00+02:00}
kind: work-item
team: Pricing
repository: pricing-service
readiness: waiting
waits_on:
  - decisions/cap-scope-6p3d
criteria:
  - id: c1
    text: Night fares count toward the cap as the settled decision states.
    verification: Run `python3 -m unittest` in the worktree and see the night fare test pass.
---

# Objective

The cap applies to night fares by the owners' decision.
