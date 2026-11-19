---
type: Workflow Plan
schema: workflow/plan/1
title: Invoice feed
description: Issued invoices reach the portal through an event feed.
generated: {by: writing-plan/0.1.0, at: 2026-10-22T09:00:00+02:00}
status: draft
repositories:
  - {
      name: billing-service,
      resource: https://git.example.org/billing/billing-service,
    }
  - {name: portal-app, resource: https://git.example.org/portal/portal-app}
---

# Objective

Issued invoices reach the portal through an event feed.

# Scope

In scope:

- The feed and its portal view.

Out of scope:

- Payment.

# Success criteria

- An invoice issued in the test environment is visible in the portal within 60 seconds.

<!-- generated:begin -->

## Decisions

- [Delivery method](/decisions/delivery-method-b3ne.md)
- [Nightly retention](/decisions/nightly-retention-7k2f.md)

## Work

- [Publish invoices as events](/work/publish-invoice-events-m2qd.md)
- [Show the feed in the portal](/work/portal-invoice-feed-c5rt.md)
- [Retire the nightly export](/work/retire-nightly-export-q7wd.md)

<!-- generated:end -->
