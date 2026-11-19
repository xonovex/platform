---
type: Workflow Decision
schema: workflow/decision/1
title: Saved searches keep the query, not the results
description: A saved search stores the normalised query and its filters, never the list of products it returned.
generated: {by: settling-questions/0.1.1, at: 2026-10-28T09:02:54Z}
verified: {by: human:a.jansen, at: 2026-10-28T09:02:54Z}
state: settled
owner: Search team
---

# Decision

A saved search stores the normalised query string and nothing about the products it returned.

# Reason

Prices and stock change every day, so a stored result list would be wrong the next time the customer opens it. The query is small, and `normalizeQuery` already gives one string for one search.

# Options considered

| Option | Gains | Costs |
| --- | --- | --- |
| Store the query | Always current results, tiny records | A search is run again on every open |
| Store the results | No second search | Stale prices and stock, larger records |

# Holds either way

- The customer sees a name for the saved search in the account menu.
