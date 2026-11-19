---
type: Workflow Decision
schema: workflow/decision/1
title: Retry limit
description: The number of attempts for a transient payment failure is still to be chosen.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T09:00:00+02:00}
state: open
owner: Team Payments
---

# Decision

When a payment fails with a transient error, the system shall retry it up to a limit that is still to be chosen.

# Reason

The limit decides how much load a failing gateway receives.

# Options considered

| Option         | Gains                  | Costs                           |
| -------------- | ---------------------- | ------------------------------- |
| Three attempts | Bounded load           | Some transient failures surface |
| Five attempts  | Fewer visible failures | More load on the gateway        |
