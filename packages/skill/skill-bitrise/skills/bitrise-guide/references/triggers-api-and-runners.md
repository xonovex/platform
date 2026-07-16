# Triggers, API, Webhooks, and Runners

## Trigger selection

Discover push/tag/pull-request trigger rules, target/source patterns, priorities, Pipeline versus Workflow selection, manual/scheduled/API triggers, and existing duplicate paths. Preview the exact event-to-execution mapping and prove near-miss branches/tags do not launch a privileged workflow.

Reconcile concurrent duplicate events by app, commit, trigger, and intended execution identity. Do not assume webhook order or exactly-once delivery. A retry either resolves the existing build or uses an explicit new attempt linked to it.

## API and webhooks

Use the documented API surface with scoped tokens, pagination, bounded rate-limit retries, cancellation, and native request/build identifiers. Keep tokens out of arguments, logs, URLs, fixtures, and error output.

Treat incoming/outgoing webhooks as untrusted notifications. Authenticate receivers where supported, reject replay, minimize retained payloads, and resolve authoritative app/build/artifact state from native references.

## Hosted and self-hosted runners

For hosted execution, pin the stack and record image/tool versions and documented isolation limits. For self-hosted execution, inventory owner, host/cluster, updates, isolation, workspace cleanup, caches, network, secrets, credentials, concurrency, logs, and incident/rollback controls.

Changing runner type is a new trust decision. Verify the effective build used the intended runner and that untrusted jobs cannot read another build's workspace, cache, or credentials.
