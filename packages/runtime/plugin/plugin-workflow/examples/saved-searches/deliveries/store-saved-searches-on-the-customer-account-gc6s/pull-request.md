---
type: Workflow Pull Request Description
schema: workflow/pull-request/1
title: 'feat(accounts): store saved searches for a customer'
description: The pull request adds the saved search store for a signed-in customer with a limit of 20.
generated: {by: preparing-delivery/0.1.1, at: 2026-10-28T09:06:13Z}
work: work/store-saved-searches-on-the-customer-account-gc6s
base: main
head: store-saved-searches
verdict: verdicts/store-saved-searches-on-the-customer-account-gc6s
verdict_matches: true
---

## Work item

Store saved searches on the customer account (`work/store-saved-searches-on-the-customer-account-gc6s`).

## Acceptance criteria

- c1 (met, self-check): A signed-in customer who saves a search gets it back from the list with the query normalised.
- c2 (met, self-check): A save request without a signed-in customer returns status 401 and stores nothing.
- c3 (met, self-check): The 21st saved search of one customer is rejected with status 409.

## Plan

web-shop: plans/saved-searches/plan.md

## Decisions

- `decisions/saved-searches-keep-the-query-not-the-results-opua`: a saved search stores the normalised query and nothing about the products it returned.
- `decisions/the-accounts-service-stores-saved-searches-vbhx`: the Accounts service stores the saved searches, at most 20 for each customer.

## How it was tested

- Verdict `verdicts/store-saved-searches-on-the-customer-account-gc6s`, a self-check at tree 498f2cd02f7a, which matches this branch: 3 of 3 criteria pass.
- Self-check: `node --test` passed with 4 tests.
