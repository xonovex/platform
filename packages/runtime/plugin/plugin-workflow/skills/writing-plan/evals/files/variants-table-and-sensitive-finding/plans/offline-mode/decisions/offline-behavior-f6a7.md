---
type: Workflow Decision
schema: workflow/decision/1
title: Offline behavior
description: Both clients keep showing the last list when the connection is lost.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T09:00:00+02:00}
verified: {by: human:jdoe, at: 2026-10-21T10:00:00+02:00}
state: settled
owner: Product
---

# Decision

When the connection is lost, the mobile and the web client shall show the last synced list with a notice.

# Reason

Customers prefer stale data with a notice over an error.

# Options considered

| Option | Gains  | Costs  |
| ------ | ------ | ------ |
| A      | Gain a | Cost a |
| B      | Gain b | Cost b |
