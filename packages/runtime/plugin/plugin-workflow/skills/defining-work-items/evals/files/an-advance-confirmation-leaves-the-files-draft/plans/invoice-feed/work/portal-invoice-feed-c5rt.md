---
type: Workflow Work Item
schema: workflow/work-item/1
title: Show the feed in the portal
description: The portal lists invoices from the feed.
generated: {by: writing-plan/0.1.0, at: 2026-10-22T09:00:00+02:00}
status: draft
kind: work-item
team: Portal
repository: portal-app
readiness: ready
depends_on:
  - work/publish-invoice-events-m2qd
criteria:
  - id: c1
    text: A published invoice is listed on the portal invoices page.
    verification: Open the invoices page for a test customer.
---

# Objective

Customers see their invoices from the feed in the portal.
