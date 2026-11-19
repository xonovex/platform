# testing-evidence: Show How You Verified the Change

Replace "tested locally" with concrete evidence: the automated tests added and what they cover, the manual environment + exact scenario + observed result, the edge cases checked (empty, error, concurrency, large input), and a way to reproduce end to end (a command, request, or deep link). For a fix, state how you confirmed the bug before and the fix after. Even a docs-only change states "no behaviour change, docs only".

```text
- Unit: added FlagCache hit/miss/expiry tests.
- Manual: ran the booking flow against the local stack, past + planned tabs load (gateway log shows 200s).
- Edge: a device with zero bookings renders the empty state.
```

## Related

[self-review.md](./self-review.md)
