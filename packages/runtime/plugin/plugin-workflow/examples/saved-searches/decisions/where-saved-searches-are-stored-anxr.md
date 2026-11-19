---
type: Workflow Decision
schema: workflow/decision/1
title: Where saved searches are stored
description: Which service owns the saved searches, the search code or the Accounts service.
generated: {by: settling-questions/0.1.1, at: 2026-10-28T09:02:55Z}
status: deprecated
state: open
owner: Accounts team
superseded_by: decisions/the-accounts-service-stores-saved-searches-vbhx
---

# Decision

Which service stores the saved searches of a customer?

# Reason

The first work item cannot choose its endpoints, its data and its tests until the home of the store is known, and the home decides who answers for a customer's data.

# Options considered

| Option | Gains | Costs |
| --- | --- | --- |
| Search service | Close to the query code | A second place that holds customer data |
| Accounts service | Customer data and consent stay in one place | The Accounts team adds a new kind of record |

# Holds either way

- Only a signed-in customer can save a search.
