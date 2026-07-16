# Azure Boards

## Native result boundary

Treat a work item as an Azure Boards record, not a portable workflow document. Preserve organization/collection, project, work-item ID and URL, type, process, revision, fields, state, relations, changed date, and actor. The opaque reference resolves through the selected Azure DevOps provider.

## Procedure

1. Discover the process model, work-item types, required fields, allowed states, area/iteration paths, link types, and actor permissions.
2. Read the current work-item revision and all relationships relevant to the requested operation.
3. Preview a JSON Patch or native CLI mutation with before/after fields, expected revision, relationship endpoints, notifications, permissions, and rollback limitations.
4. Apply conditionally against the observed revision. A conflict triggers a re-read and new preview.
5. Verify the item revision, history entry, fields, and forward/reverse relationship through native reads.

Use WIQL for queries, not as an immutable snapshot. Query results and `System.ChangedDate` express freshness; the work-item revision is the optimistic-concurrency signal. Bind code/build evidence separately to immutable commit, build, or artifact references.

## Relationships

Keep relationship type and direction explicit: parent/child, predecessor/successor, related, commit, pull request, build, artifact, or external link. Do not infer a relationship from title text, tags, or a copied URL. Verify both the work-item relation and the target provider record where assurance requires it.

## Failures

Surface process-field rejection, invalid state transition, stale revision, missing relationship target, authorization failure, partial link creation, rate limit, and provider outage. Never fall back to a local issue file for a selected Azure Boards write.
