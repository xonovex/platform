# Heuristics and Oracles

Use risk-relevant dimensions rather than applying every checklist:

- structure: components, interfaces, dependencies, boundaries;
- function: inputs, outputs, transformations, calculations, permissions;
- data: empty, typical, extreme, malformed, stale, duplicated, migrated, sensitive;
- sequence and state: first use, repeat, retry, undo, refresh, timeout, resume;
- environment: platform, locale, time, network, resource pressure, configuration;
- interaction: concurrency, ordering, partial failure, external-service behavior;
- operations: observability, recovery, rollback, supportability, cleanup.

Judge behavior using named oracles:

- explicit requirement or acceptance example;
- domain invariant, policy, contract, or safety rule;
- user expectation and recoverability;
- consistency within the product or across equivalent paths;
- comparison with a prior trusted revision;
- inverse, round-trip, idempotence, conservation, or monotonic property;
- independent calculation or alternate implementation.

An oracle can be fallible. Record what it is, why it is relevant, and any disagreement
instead of presenting surprise as proof of a defect.
