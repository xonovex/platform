# Repos, Pull Requests, and Policies

## Source and review references

Preserve project/repository IDs, commit IDs, ref names, pull-request ID, source/target commits, iterations, thread/comment IDs, status contexts, policy configuration/evaluation IDs, and native URLs. Bind validation evidence to the exact commit under review; a mutable branch name is insufficient.

## Branch policy onboarding

1. Discover applicable project/repository permissions and every policy on the target ref, including scope, enabled/blocking state, path filters, build queue behavior, reviewers, comment resolution, statuses, bypass actors, and inherited settings.
2. Map the semantic requirement to separate native controls: minimum reviewers, required reviewers, linked work items, comment resolution, build validation, or external status.
3. Preview the exact policy configuration, build definition, stable status/check identity, queue/expiry behavior, bypass permissions, failure behavior, and rollback target.
4. Apply with the expected configuration revision, then re-read the effective policies.
5. Probe a compliant exact commit and negative cases for omitted, renamed, spoofed, skipped, stale, failed, and bypassed validation.

Do not treat pull-request approval, branch write permission, build validation, and external status as equivalent. Preserve each native decision/evaluation reference independently.

## Pull-request writes

Re-fetch iterations and commit IDs after every push before anchoring comments or evaluating policy. Writes are not automatically idempotent: reconcile by native pull-request/thread/status identity before retrying. A 2xx response does not prove the thread anchored correctly or a blocking policy evaluated the current commit.
