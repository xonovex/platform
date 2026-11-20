---
type: Workflow Work Item
schema: workflow/work-item/1
title: Open a saved search from the account menu
description: The account menu lists the saved searches of the customer and opens one with its stored query.
generated: {by: defining-work-items/0.1.1, at: 2026-10-28T09:05:16Z}
verified: {by: human:a.jansen, at: 2026-10-28T09:05:16Z}
status: stable
kind: work-item
parent: work/saved-searches-for-signed-in-customers-4kk7
team: Search team
repository: web-shop
readiness: ready
depends_on:
  - work/store-saved-searches-on-the-customer-account-gc6s
decisions:
  - decisions/saved-searches-keep-the-query-not-the-results-opua
  - decisions/the-accounts-service-stores-saved-searches-vbhx
criteria:
  - id: c1
    text: Selecting a saved search in the account menu opens the product list with the stored query and filters.
    verification: Open the saved search of a test customer on the staging shop and compare the URL parameters with the stored query.
  - id: c2
    text: The opened list shows the current prices and stock, not those from the time of saving.
    verification: Change the price of a product on staging after saving, open the saved search again and read the new price.
---

# Objective

The account menu shows the saved searches of the signed-in customer, and selecting one runs the stored query again so that the customer sees current results.
