---
type: Workflow Work Item
schema: workflow/work-item/1
title: Publish invoices as events
description: The billing service publishes each issued invoice to the feed.
generated: {by: defining-work-items/0.1.0, at: 2026-10-22T09:00:00+02:00}
status: draft
kind: work-item
team: Billing
repository: billing-service
readiness: ready
decisions:
  - decisions/delivery-method-b3ne
criteria:
  - id: c1
    text: An invoice issued in the test environment appears on the feed within 60 seconds.
    verification: Issue a test invoice and read the feed.
  - id: c2
    text: An invoice the billing service rejects is not published.
    verification: Issue an invalid invoice and read the feed.
---

# Objective

The billing service publishes every issued invoice to the feed.
