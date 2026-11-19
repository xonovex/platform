---
type: Workflow Work Item
schema: workflow/work-item/1
title: Retry from settings
description: Retry from settings.
generated: {by: writing-plan/0.1.0, at: 2026-10-22T10:00:00+02:00}
status: stable
kind: work-item
team: platform
repository: shop-app
readiness: ready
criteria:
  - id: c1
    text: Retries stop within 200 ms of the limit.
    verification: Run npm run bench:retry and read the p95; saved as checks/bench.txt
---

# Objective

Retry from settings.
