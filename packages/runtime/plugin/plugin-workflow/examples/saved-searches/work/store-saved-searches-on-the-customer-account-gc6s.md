---
type: Workflow Work Item
schema: workflow/work-item/1
title: Store saved searches on the customer account
description: The Accounts service saves, lists and limits the saved searches of a signed-in customer.
generated: {by: defining-work-items/0.1.1, at: 2026-10-28T09:05:18Z}
verified: {by: human:a.jansen, at: 2026-10-28T09:05:17Z}
status: stable
kind: work-item
parent: work/saved-searches-for-signed-in-customers-4kk7
team: Accounts team
repository: web-shop
readiness: ready
decisions:
  - decisions/saved-searches-keep-the-query-not-the-results-opua
  - decisions/the-accounts-service-stores-saved-searches-vbhx
criteria:
  - id: c1
    text: A signed-in customer who saves a search gets it back from the list with the query normalised.
    verification: 'Run `node --test test/saved-searches.test.mjs` and read the result of the test "lists a search saved by the signed-in customer".'
  - id: c2
    text: A save request without a signed-in customer returns status 401 and stores nothing.
    verification: 'Run `node --test test/saved-searches.test.mjs` and read the result of the test "rejects a save without a signed-in customer and stores nothing".'
  - id: c3
    text: The 21st saved search of one customer is rejected with status 409.
    verification: 'Run `node --test test/saved-searches.test.mjs` and read the result of the test "rejects the 21st saved search of one customer".'
---

# Objective

The Accounts service saves a named search for a signed-in customer, lists the searches of that customer and refuses a 21st one.

# Notes

- The stored query is the output of `normalizeQuery`, as the decision on what a saved search keeps requires.
