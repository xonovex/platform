# testing-evidence: Show How You Verified the Change

**Guideline:** Document the evidence that the change works - environment, scenarios, edge cases, and how a reviewer can reproduce it - not just "tested locally".

**Rationale:** "Tested locally" is unverifiable and tells the reviewer nothing. Concrete evidence lets them trust the change without pulling the branch, and reproduction steps catch the bugs unit tests miss by exercising the code the way a user or caller would.

**How to Apply:**

1. List the automated tests added or updated, and what they cover.
2. Describe manual verification: the environment, the exact scenario, and the observed result.
3. Call out the edge cases you checked (empty, error, concurrency, large input).
4. Give the reviewer a way to reproduce it end to end (a command, a request, a deep link).
5. For a fix, state how you confirmed the bug before and the fix after.

**Example:**

```text
// Bad
Tested locally, works.

// Good
- Unit: added FlagCache hit/miss/expiry tests.
- Manual: ran the booking flow against the local stack, past + planned tabs load (gateway log shows 200s).
- Edge: a device with zero bookings renders the empty state.
```

**Counter-Example:** None - even a docs-only change states "no behaviour change, docs only".

**Related:** [self-review.md](./self-review.md)
