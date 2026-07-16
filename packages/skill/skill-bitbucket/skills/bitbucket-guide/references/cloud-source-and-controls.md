# Bitbucket Cloud Source and Controls

## Native source boundary

Preserve workspace and repository UUIDs, commit hash, ref, pull-request ID and source/destination commits, comment/task IDs, commit-status key/name/URL, branch-restriction ID, merge/custom-check result, webhook UUID/delivery, and native URLs. A mutable branch or status display name is not an immutable revision.

## Repository and pull-request operations

Read the current pull-request and commit before every write. Reconcile create/comment/status requests by native identity before retrying; writes are not assumed idempotent. Re-fetch the source commit after a push, and verify comments, approvals, tasks, and status/check results against that exact commit.

## Controls

1. Discover branch restrictions/permissions, merge checks, custom checks, required approvals/tasks/builds, deployment restrictions, bypass groups/users/apps, and plan/tier availability.
2. Map each requirement to one native control and stable status/check identity. Do not treat an approval, branch restriction, pipeline status, and custom check as interchangeable.
3. Preview exact repository/workspace mutations, callback/app identity, requested scopes, data sent to integrations, failure/timeout/retry behavior, verification, and rollback.
4. Re-read effective settings and test compliant, omitted, renamed, spoofed, stale, failed, timed-out, duplicate, and bypass cases.

## REST and webhooks

Use Cloud REST 2.0 resource links and pagination. Respect rate-limit/retry hints with bounded backoff and cancellation. Retry a write only after resolving whether its native object already exists.

Treat webhook payloads as untrusted notifications. Verify receiver authentication where configured, reject replay, reconcile duplicates/out-of-order deliveries, minimize payload retention, and re-read the referenced repository, pull request, pipeline, or deployment as authority.
