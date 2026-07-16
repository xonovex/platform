# deliver-publish: Publish a Reviewable Candidate

## Core workflow

1. Resolve the exact Development or consolidated Development source revisions and the selected deliverable provider's capabilities.
2. Verify required source validation, target intent, publication authority, provider authentication, duplicate/idempotency behavior, and profile-required governance evidence.
3. Ask the provider adapter to publish the reviewable candidate using its native resource, relationships, and side effects. A repository commit, pull request, merge request, artifact record, database record, or non-file work item may each be valid when selected.
4. Resolve the created native reference again and obtain the strongest immutable revision or digest the provider supports. Report weaker freshness tokens and their limitations; never fabricate immutability.
5. Publish or return a Deliverable Publication result with the native candidate reference, immutable revision, intended target, source Development revisions, publication status, actor, provider evidence, and follow-up capabilities.

An explicitly selected unavailable provider fails visibly. Do not fall back to a local file, repository, or another host. Publication makes a candidate reviewable; it does not accept, integrate, release, or deploy it.
