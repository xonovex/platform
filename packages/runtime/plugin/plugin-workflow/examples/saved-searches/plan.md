---
type: Workflow Plan
schema: workflow/plan/1
title: Saved searches
description: A signed-in customer saves a product search and opens it again later from the account menu.
generated: {by: revising-plan/0.1.1, at: 2026-10-28T09:04:49Z}
verified: {by: human:a.jansen, at: 2026-10-28T09:04:49Z}
status: stable
owners:
  - Search team
  - Accounts team
repositories:
  - name: web-shop
    resource: https://git.example.test/shop/web-shop.git
sources:
  - id: search-options
    resource: /research/saved-search-options-ovwb.md
    title: Saved search options
tags:
  - search
  - accounts
---

# Objective

A signed-in customer saves the product search they are looking at, gives it a name and opens it again later from the account menu. The shop keeps the query and its filters, so the results are always current, and the Accounts service keeps up to 20 saved searches for each customer.

![A signed-in customer saves and opens a search](diagrams/saved-searches-lead.svg)

# Scope

In scope:

- Saving a named search for a signed-in customer, with a limit of 20 for each customer
- Listing the saved searches and opening one from the account menu

Out of scope:

- Saved searches for guests
- A badge in the account menu with the number of saved searches, raised by the Search team for a later change

# Current state

The shop turns a search into one normalised string, and the account session names the signed-in customer or a guest.[^search-options] Nothing stores a search today.

# Success criteria

- A signed-in customer who saves a search sees it in the saved searches list of the account menu, checked on the staging shop with a test customer.
- Opening a saved search shows the product list for the stored query and filters with current prices and stock, checked by comparing the page with a fresh search for the same query.
- A request to save a search without a signed-in customer returns status 401 and stores nothing, checked by the Accounts service tests.
- The 21st saved search of one customer is rejected with status 409, checked by the Accounts service tests.

<!-- generated:begin -->

## Decisions

- [Saved searches keep the query, not the results](/decisions/saved-searches-keep-the-query-not-the-results-opua.md)
- [The Accounts service stores saved searches](/decisions/the-accounts-service-stores-saved-searches-vbhx.md)

## Work

- [Saved searches for signed-in customers](/work/saved-searches-for-signed-in-customers-4kk7.md)
- [Store saved searches on the customer account](/work/store-saved-searches-on-the-customer-account-gc6s.md)
- [Open a saved search from the account menu](/work/open-a-saved-search-from-the-account-menu-66en.md)

<!-- generated:end -->

[^search-options]: Saved search options, section Observed
