---
type: Workflow Work Item
schema: workflow/work-item/1
title: Fare rules model
description: The fare rules model the cap reads.
generated: {by: defining-work-items/0.1.0, at: 2026-10-22T10:00:00+02:00}
verified: {by: human:pricing-owner, at: 2026-10-22T10:30:00+02:00}
kind: work-item
team: Pricing
repository: pricing-service
readiness: ready
criteria:
  - id: c1
    text: The fare rules load from the rules file.
    verification: Run `python3 -m unittest` in the worktree and see the rules test pass.
---

# Objective

The cap reads its daily price from a rules model.
