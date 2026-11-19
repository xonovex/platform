---
type: Workflow Commit Message
schema: workflow/commit-message/1
title: Commit message for storing saved searches
description: The commit stores saved searches for a signed-in customer with a limit of 20.
generated: {by: preparing-delivery/0.1.1, at: 2026-10-28T09:06:12Z}
work: work/store-saved-searches-on-the-customer-account-gc6s
branch: store-saved-searches
---

feat(accounts): store saved searches for a customer

A signed-in customer can now save a named search and list the saved ones. The store keeps the normalised query so that opening a saved search shows current prices and stock, and it refuses a save without a signed-in customer or beyond the limit of 20.

- decisions/saved-searches-keep-the-query-not-the-results-opua: a saved search stores the normalised query and nothing about the products it returned.
- decisions/the-accounts-service-stores-saved-searches-vbhx: the Accounts service stores the saved searches, at most 20 for each customer.

Plan: web-shop: plans/saved-searches/plan.md
