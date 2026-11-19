---
type: Workflow Inline Comment
schema: workflow/inline-comment/1
title: Limit of saved searches
description: The limit of 20 comes from a decision of the Accounts team, not from a technical constraint.
generated: {by: preparing-delivery/0.1.1, at: 2026-10-28T09:06:14Z}
work: work/store-saved-searches-on-the-customer-account-gc6s
path: src/accounts/saved-searches.mjs
line: 4
side: right
marker: limit-per-customer
label: decision
---

The Accounts team asked for at most 20 saved searches for each customer, see `decisions/the-accounts-service-stores-saved-searches-vbhx`. Change the number only through a new decision.
