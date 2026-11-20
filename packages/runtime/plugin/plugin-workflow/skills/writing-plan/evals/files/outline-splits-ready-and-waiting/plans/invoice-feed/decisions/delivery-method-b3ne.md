---
type: Workflow Decision
schema: workflow/decision/1
title: Delivery method
description: Invoices reach finance through an event feed.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T09:00:00+02:00}
verified: {by: human:jdoe, at: 2026-10-21T10:00:00+02:00}
state: settled
owner: Billing
---

# Decision

When an invoice is issued, the system shall publish it to the invoice feed.

# Reason

The feed supports retries.

# Options considered

| Option | Gains  | Costs  |
| ------ | ------ | ------ |
| A      | Gain a | Cost a |
| B      | Gain b | Cost b |
