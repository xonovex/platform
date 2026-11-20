---
type: Workflow Decision
schema: workflow/decision/1
title: Cap is applied before discounts
description: The daily cap applies to the undiscounted fares.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T11:00:00+02:00}
verified: {by: human:alex, at: 2026-10-20T11:05:00+02:00}
state: settled
---

# Decision

When a rider's undiscounted fares exceed the daily price, the system shall charge the daily price before applying any discount.

# Reason

The pricing rules apply the cap first.

# Options considered

| Option         | Gains             | Costs                                |
| -------------- | ----------------- | ------------------------------------ |
| Cap first      | Matches the rules | Discount has no effect above the cap |
| Discount first | Rewards riders    | Contradicts the rules                |
