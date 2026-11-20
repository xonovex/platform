---
type: Workflow Decision
schema: workflow/decision/1
title: Retry limit
description: Payments are retried up to three times.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T09:00:00+02:00}
state: settled
owner: Team Payments
sources:
  - id: gateway-limits
    resource: /research/gateway-limits-q4w2.md
    title: Research on the gateway limits
---

# Decision

When a payment fails with a transient error, the system shall retry it up to three times.

# Reason

The gateway sheds load after the fourth request in a minute.[^gateway-limits]

# Options considered

| Option         | Gains                  | Costs                           |
| -------------- | ---------------------- | ------------------------------- |
| Three attempts | Bounded load           | Some transient failures surface |
| Five attempts  | Fewer visible failures | More load on the gateway        |

[^gateway-limits]: Research on the gateway limits
