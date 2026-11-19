---
type: Workflow Work Item
schema: workflow/work-item/1
title: Retire the nightly export
description: The nightly export stops once the feed is live.
generated: {by: writing-plan/0.1.0, at: 2026-10-22T09:00:00+02:00}
status: draft
kind: work-item
team: Billing
repository: billing-service
readiness: waiting
waits_on:
  - decisions/nightly-retention-7k2f
depends_on:
  - work/publish-invoice-events-m2qd
criteria:
  - id: c1
    text: The nightly export job does not run.
    verification: Check the scheduler for the job.
---

# Objective

The nightly export no longer runs after the feed is live.
