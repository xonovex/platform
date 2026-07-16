# Provider Conformance

Test Services and every supported Server release independently against a pinned source snapshot and live tenant/collection where credentials are available.

## Required cases

- Product/edition/version/API detection; unknown and stale baselines fail clearly.
- Boards create/read/revise/relate/version operations, revision conflict, relationship reconstruction after restart, and native work-item references.
- Repos commit/pull-request references, policy layering, bypass inventory, exact-revision build validation, and omitted/renamed/spoofed/stale status failures.
- Reusable pipeline template, protected-resource approval/check, artifact digest/access/retention, and provider-native run/timeline/artifact references.
- Workload-federated service connection with constrained claims, least privilege, temporary credential expiry, forbidden out-of-scope action, and no generated long-lived secret.
- REST pagination, unsupported API version, authorization failure, rate limit, bounded retry, duplicate request, partial apply, outage, and service-hook duplicate/replay/order behavior.
- Preview/authorization separation, idempotent apply, authoritative verify, rollback, uninstall ownership, and deliberate drift.

Store fixture facts, expected outcomes, and opaque example identifiers without credentials or live payload content. Mark documentation-only cases as such; never report them as live passes.

The repository-wide enterprise fixture validates the shared contract and mixed-stack composition. Provider-specific live probes add native evidence references and may be skipped only with an explicit `not-probed` result and reason.
