---
type: Workflow Close-out Record
schema: workflow/close-out/1
title: Close-out of the saved searches package
description: Two of four success criteria are met, and the menu work item stands between the plan and done.
generated: {by: validating-work-package/0.1.1, at: 2026-10-28T09:06:49Z}
plan: plan
package: work/saved-searches-for-signed-in-customers-4kk7
result: partly-met
criteria:
  - criterion: "A signed-in customer who saves a search sees it in the saved searches list of the account menu, checked on the staging shop with a test customer."
    status: open
    evidence: "The Search team owns the menu item, which is not built yet; it is tracked as work/open-a-saved-search-from-the-account-menu-66en."
  - criterion: "Opening a saved search shows the product list for the stored query and filters with current prices and stock, checked by comparing the page with a fresh search for the same query."
    status: open
    evidence: "The Search team owns the menu item, which is not built yet; it is tracked as work/open-a-saved-search-from-the-account-menu-66en."
  - criterion: "A request to save a search without a signed-in customer returns status 401 and stores nothing, checked by the Accounts service tests."
    status: met
    evidence: "node --test test/saved-searches.test.mjs passed 3 tests in this run at tree 498f2cd02f7a, including the test for status 401."
  - criterion: "The 21st saved search of one customer is rejected with status 409, checked by the Accounts service tests."
    status: met
    evidence: "The same run passed the test that saves 20 searches and gets status 409 for the 21st."
sources:
  - id: saved-search-store
    resource: web-shop@tree-498f2cd02f7ac717e9f2c1c5cffa4f509e34a8e4:src/accounts/saved-searches.mjs
    title: Saved search store
  - id: saved-search-tests
    resource: web-shop@tree-498f2cd02f7ac717e9f2c1c5cffa4f509e34a8e4:test/saved-searches.test.mjs
    title: Saved search tests
---

# Independence

This assessment ran in the context that implemented the work, so it is not an independent check.

# Status

Partly met: the storage item meets two criteria and the menu item, which carries the other two, is not built. The schedule depends on the Search team starting the menu item, because the storage item it depends on is built.

## Met and evidenced

| Criterion | Evidence |
| --- | --- |
| A save without a signed-in customer returns 401 | `node --test test/saved-searches.test.mjs` passed 3 tests in this run, tree 498f2cd02f7a.[^saved-search-tests] |
| The 21st saved search is rejected with status 409 | The limit is the constant at line 4 and the check at line 14.[^saved-search-store] |

## Gated

None.

## Open

| Criterion | Owner | Where |
| --- | --- | --- |
| The saved search list in the account menu | Search team | work/open-a-saved-search-from-the-account-menu-66en |
| Opening a saved search shows current results | Search team | work/open-a-saved-search-from-the-account-menu-66en |

# Integration

The storage item is the only item built, so the integration with the menu item cannot be checked yet. The package has no other work items to compare it with.

# Blockers and gates

None.

# Open points and residue

- The commit message and the pull request description of the storage item are prepared and not applied. a.jansen decides when to commit and to open the pull request.
- The verdict for the storage item is a self-check, so the criteria it marks are not independently validated.
- The Search team note on a count badge is outside the scope of this plan, and the follow-up seed below holds it.

# Follow-up seeds

## Saved search count badge

- Objective: The account menu shows how many saved searches a customer has.
- Why it exists: The Search team raised it while the storage question was settled, and it is outside the scope of this plan.
- Start: after the menu work item is delivered

### Current state, evidenced

The store lists the saved searches of a customer as an array, so the count is the length of the list.[^saved-search-store]

### What is open

- Does the badge show the count or only a dot when there is at least one saved search?

### Known traps

- The list endpoint returns a 401 for a guest, so the badge must not request it for guests.

### Success criteria sketch

- The account menu of a test customer with 3 saved searches shows the number 3, checked on the staging shop.

### Validation

A test on the list response length and a screenshot of the menu on staging.

### Gated on

Nothing.

### Skills to consult

None.

# Not checked

- The menu item was not read or run, because it does not exist yet.
- Nothing was run on the staging shop.

[^saved-search-store]: Saved search store

[^saved-search-tests]: Saved search tests
