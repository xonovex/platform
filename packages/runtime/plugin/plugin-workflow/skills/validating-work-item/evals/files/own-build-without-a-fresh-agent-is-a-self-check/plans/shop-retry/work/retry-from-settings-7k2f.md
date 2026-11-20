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
    text: Retries stop after the configured limit.
    verification: Run npm test -- retry and read the limit case; saved as checks/retry-test.txt
  - id: c2
    text: The limit is read from the settings.
    verification: Read the settings loader test and the diff of src/retry.js
---

# Objective

Retry from settings.
