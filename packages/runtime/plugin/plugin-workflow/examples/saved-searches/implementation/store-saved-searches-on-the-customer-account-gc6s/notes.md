---
type: Workflow Implementation Notes
schema: workflow/implementation-notes/1
title: Notes for storing saved searches
description: Added the saved search store with the sign-in rule and the limit of 20, uncommitted in the worktree, with every criterion met by its own check.
generated: {by: implementing-work-item/0.1.1, at: 2026-10-28T09:05:41Z}
work: work/store-saved-searches-on-the-customer-account-gc6s
worktree: ../web-shop-store-saved-searches
branch: store-saved-searches
tasks:
  - id: saved-search-store
    concern: store-and-limit
    summary: Added createSavedSearches with save and list, the 401 rule and the limit of 20.
  - id: saved-search-tests
    concern: criteria-tests
    summary: Added one test for each criterion of the work item.
decision_notes:
  - decision: decisions/the-accounts-service-stores-saved-searches-vbhx
    location: src/accounts/saved-searches.mjs:4
    note: The limit of 20 comes from the decision that the Accounts service keeps at most 20 for each customer.
  - decision: decisions/saved-searches-keep-the-query-not-the-results-opua
    location: src/accounts/saved-searches.mjs:15
    note: The entry keeps the normalised query and the name, never the products a search returned.
checks:
  - command: node --test test/saved-searches.test.mjs
    result: passed, 3 tests
  - command: node --test
    result: passed, 4 tests
---

# Summary

The Accounts module now saves, lists and limits saved searches, and the changes sit uncommitted in the worktree on branch store-saved-searches. Nothing is left for this work item.

# Self-check against the criteria

| Criterion | Evidence | State |
| --- | --- | --- |
| `c1` | `node --test test/saved-searches.test.mjs`, test "lists a search saved by the signed-in customer" ran green | met |
| `c2` | Same command, test "rejects a save without a signed-in customer and stores nothing" ran green | met |
| `c3` | Same command, test "rejects the 21st saved search of one customer" ran green | met |
