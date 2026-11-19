---
type: Workflow Decision
schema: workflow/decision/1
title: The Accounts service stores saved searches
description: The Accounts service owns the saved searches and keeps at most 20 for each customer.
generated: {by: revising-plan/0.1.1, at: 2026-10-28T09:03:54Z}
verified: {by: human:a.jansen, at: 2026-10-28T09:03:54Z}
state: settled
owner: Accounts team
supersedes:
  - decisions/where-saved-searches-are-stored-anxr
sources:
  - id: search-options
    resource: /research/saved-search-options-ovwb.md
    title: Saved search options
---

# Decision

The Accounts service stores the saved searches, at most 20 for each customer.

# Reason

Customer data and its consent rules already live in the Accounts service, and the research found no storage of its own in the search code.[^search-options] The Accounts team asked for the limit of 20 when the storage question was settled, and the critique asked for a limit to protect the account page.

# Options considered

| Option | Gains | Costs |
| --- | --- | --- |
| Accounts service | Customer data and consent stay in one place | The Accounts team adds a new kind of record |
| Search service | Close to the query code | A second place that holds customer data |

# Holds either way

- Only a signed-in customer can save a search.

[^search-options]: Saved search options, section Unknowns
