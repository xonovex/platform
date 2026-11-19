---
type: Workflow Verdict
schema: workflow/verdict/1
title: Verdict for storing saved searches
description: All 3 criteria passed, checked as a self-check by the context that implemented the work.
generated: {by: validating-work-item/0.1.1, at: 2026-10-28T09:05:56Z}
work: work/store-saved-searches-on-the-customer-account-gc6s
worktree_hash: 498f2cd02f7ac717e9f2c1c5cffa4f509e34a8e4
source: self-check
results:
  - criterion: c1
    result: pass
    evidence: 'node --test test/saved-searches.test.mjs ran the test "lists a search saved by the signed-in customer" and it passed.'
  - criterion: c2
    result: pass
    evidence: 'The same command ran the test "rejects a save without a signed-in customer and stores nothing" and it passed.'
  - criterion: c3
    result: pass
    evidence: 'The same command ran the test "rejects the 21st saved search of one customer" and it passed.'
---

# Result

3 pass, 0 fail, 0 blocked, tree 498f2cd02f7a, self-check.

Diff checked:

```text
 src/accounts/saved-searches.mjs | 26 ++++++++++++++++++++++++++
 test/saved-searches.test.mjs    | 25 +++++++++++++++++++++++++
 2 files changed, 51 insertions(+)
```

# Criteria

## c1 A signed-in customer who saves a search gets it back from the list with the query normalised.

- Command: `node --test --test-reporter=spec test/saved-searches.test.mjs`
- Observed: the first test passed and compares the list with the normalised query `color=red&q=kettle`.
- Limits: runs against the in-memory store, not a deployed Accounts service.

## c2 A save request without a signed-in customer returns status 401 and stores nothing.

- Command: the same command.
- Observed: the second test passed, with status 401 and an empty list afterwards.
- Limits: the session is a plain object, so no real sign-in was exercised.

## c3 The 21st saved search of one customer is rejected with status 409.

- Command: the same command.
- Observed: the third test saved 20 searches with status 201 each and got status 409 for the 21st.
- Limits: the limit is checked for one customer only.

# Could not check

- The checks ran in the context that also wrote the code, so this is a self-check and not an independent validation.
