---
type: Workflow Decision
schema: workflow/decision/1
title: Cache store
description: Product data is cached in a shared in-memory store.
generated: {by: settling-questions/0.1.0, at: 2026-10-20T09:00:00+02:00}
verified: {by: human:jdoe, at: 2026-10-21T10:00:00+02:00}
state: settled
owner: Platform
---

# Decision

When a product page is requested, the system shall serve product data from the shared in-memory store before it reads the database.

# Reason

A local cache per server was rejected because stale copies of the price diverged between servers during the trial.

# Options considered

| Option | Gains  | Costs  |
| ------ | ------ | ------ |
| A      | Gain a | Cost a |
| B      | Gain b | Cost b |
