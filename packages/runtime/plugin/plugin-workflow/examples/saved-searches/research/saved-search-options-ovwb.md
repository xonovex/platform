---
type: Workflow Research
schema: workflow/research/1
title: Saved search options
description: What the shop already has for saving a search, and the two questions that are still open.
generated: {by: researching/0.1.1, at: 2026-10-28T09:02:34Z}
subject: What the web shop needs to let a signed-in customer save a product search and open it again.
kind: feature
roles:
  - search-explorer
  - accounts-explorer
questions:
  - id: what-is-stored
    text: Does a saved search keep the query or the result list?
  - id: storage-location
    text: Which service stores the saved searches?
sources:
  - id: normalize-query
    resource: web-shop@a9a7d3e2b712:src/search/normalize-query.mjs
    title: Query normalisation
  - id: session
    resource: web-shop@a9a7d3e2b712:src/accounts/session.mjs
    title: Account session
---

# Bottom line

The shop can already turn a search into one stable string and can tell a signed-in customer from a guest, so saving a search needs a store and two endpoints. The largest gap is the home of the store: no code or document says which service owns it.

# Observed

## Repositories read

- web-shop at a9a7d3e2b712: `git -C web-shop rev-parse HEAD`

## search-explorer

- `normalizeQuery` sorts the query parameters and drops empty ones, so two spellings of one search give one string.[^normalize-query]

## accounts-explorer

- `customerIdOf` returns the customer id of a session, or null for a guest.[^session]

# Inference

- Inference: a saved search can hold the output of `normalizeQuery`, from the fact that it already gives one string for one search.
- Inference: the sign-in rule can reuse `customerIdOf`, from the fact that it is the only place that reads the session for a customer.

# Unknowns

- Whether the search code has any persistent storage of its own.

# Open decisions

| Id | Question | Options | Recommendation |
| --- | --- | --- | --- |
| `what-is-stored` | Does a saved search keep the query or the results? | The query; the result list | Inference: the query, because results change with prices and stock. |
| `storage-location` | Which service stores the saved searches? | Search service; Accounts service | Inference: the Accounts service, because customer data lives there. |

# Not checked

- Nothing was run beyond reading the two files.

[^normalize-query]: Query normalisation, lines 1 to 7

[^session]: Account session, lines 1 to 3
