---
type: Workflow Work Item
schema: workflow/work-item/1
title: Retry transient payment failures
description: The client retries a payment that failed with a transient error.
generated: {by: writing-plan/0.1.0, at: 2026-10-20T09:00:00+02:00}
status: draft
kind: work-item
team: Team Payments
readiness: waiting
waits_on:
  - decisions/retry-limit-a2b3
criteria:
  - id: c1
    text: A transient failure is retried and the retry is visible in the client log.
    verification: Unit test with a stub gateway that fails twice.
---

# Objective

The client retries a payment that failed with a transient error.
