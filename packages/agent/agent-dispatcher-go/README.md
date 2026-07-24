# Agent Dispatcher

The dispatcher is the external workflow boundary for GitHub and GitLab. It verifies
webhooks, stores their raw payloads as untrusted data, normalizes trusted routing
fields, and executes explicitly authorized provider effects through a PostgreSQL
outbox.

It does not interpret webhook text as instructions and does not create `AgentRun`
resources. Agent dispatch and structured result ingestion are phase 4 and remain a
separate, intentionally absent boundary.

## Implemented runtime flow

```text
GitHub or GitLab webhook
  -> signature and replay validation
  -> durable delivery deduplication
  -> provider-neutral workflow event
  -> external trusted coordinator
  -> authenticated effect API
  -> allowlist and precondition validation
  -> preview or durable outbox
  -> provider reconciliation and write
  -> audit, retry, or dead letter
```

Webhook events never create effects directly. Until phase 4 is implemented, a trusted
coordinator submits effects to `POST /v1/effects`. This preserves the security boundary
without pretending that an agent session has run.

## Provider coverage

| Operation                       | GitHub                        | GitLab                                              |
| ------------------------------- | ----------------------------- | --------------------------------------------------- |
| Create/update/open/close ticket | Issues REST API               | Issues REST API                                     |
| Add or update kanban item       | Projects v2 GraphQL           | Issue labels or Work Item Status GraphQL            |
| Archive kanban item             | Projects v2 GraphQL           | Not a GitLab board-card operation                   |
| Publish context                 | Issue or pull-request comment | Issue or merge-request note                         |
| Publish review                  | Pull-request review           | Merge-request note and optional SHA-pinned approval |
| Create/update deployment        | Deployments REST API          | Deployments REST API                                |
| Approve/reject deployment       | Not applicable                | Deployment approval API                             |

GitLab boards are views over issue/work-item attributes. Label-backed status works on
all supported tiers. Native Work Item Status requires a `work_item_id` and status
`option_id`, and depends on the GitLab version and tier. GitHub Project status requires
the Project, item, field, and option node IDs.

The implementation follows the current
[GitHub webhook security guidance](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries),
[GitHub webhook delivery practices](https://docs.github.com/en/webhooks/using-webhooks/best-practices-for-using-webhooks),
[GitHub Projects GraphQL model](https://docs.github.com/en/graphql/reference/projects),
[GitLab Standard Webhooks contract](https://docs.gitlab.com/user/project/integrations/webhooks/),
[GitLab Issues API](https://docs.gitlab.com/api/issues/), and
[GitLab Deployments API](https://docs.gitlab.com/api/deployments/).

## Context guarantees

`context.publish` validates the same canonical record used by the workflow skill:

- the context ID has a stable restricted form;
- versions are positive and successors identify the superseded record;
- the SHA-256 digest covers the canonical semantic lines from `Type` through
  `Visibility`;
- `internal` context cannot be sent to a provider;
- a matching provider marker is a completed retry;
- duplicate matches or the same ID/version with different visible content fail closed;
- historical provider comments and notes are never rewritten or deleted.

Provider comment creation does not expose a uniqueness or compare-and-swap primitive.
The PostgreSQL resource lease gives dispatcher replicas one writer, while external
writers can still race the final provider create. Every retry therefore reconciles the
complete marker set before another write.

## Persistence and recovery

The embedded migration creates:

- deduplicated webhook deliveries and normalized workflow events;
- effects serving as the transactional outbox;
- renewable resource leases for multi-replica single-writer execution;
- append-only context records;
- immutable audit events;
- dead letters with an explicit authenticated requeue operation.

Retryable network, rate-limit, and server failures use bounded exponential retry or
the provider's `Retry-After` value. Unknown create results are reconciled by stable
markers or the pinned deployment tuple. Permanent validation, authorization,
precondition, duplicate, and conflict failures go to the dead-letter queue.

## Configuration

Copy `config.example.json` to a deployment-specific path and provide it through
`DISPATCHER_CONFIG`. Secret values stay in environment variables named by the file;
the JSON file contains only secret variable names.

Required runtime variables for the example:

```text
DISPATCHER_CONFIG=/run/config/dispatcher.json
DISPATCHER_ADMIN_TOKEN=<coordinator bearer token>
DISPATCHER_DATABASE_URL=<PostgreSQL connection URL>
GITHUB_APP_INSTALLATION_TOKEN=<short-lived installation token>
GITHUB_WEBHOOK_SECRET=<webhook HMAC secret>
GITLAB_PROJECT_ACCESS_TOKEN=<least-privilege project or group token>
GITLAB_WEBHOOK_SIGNING_TOKEN=<whsec_ Standard Webhooks signing token>
```

Use the GitHub API base URL for GitHub or GitHub Enterprise. Use the GitLab instance
root URL; the adapter selects `/api/v4` and `/api/graphql` itself. Plain HTTP is
rejected unless a tenant explicitly enables `allow_insecure_http`, which is intended
only for local tests.

Start in preview mode:

```bash
go run ./cmd/agent-dispatcher
```

Set `apply_enabled` only after preview payloads, permissions, webhook signatures,
repository allowlists, and provider conformance fixtures have been reviewed.

## Validation

Run the complete local gate with:

```bash
npx moon run agent-dispatcher-go:ci-check
go test -race ./...
```

The PostgreSQL lifecycle test is skipped unless it receives a disposable test
database:

```bash
DISPATCHER_TEST_DATABASE_URL='postgres://postgres:password@127.0.0.1:5432/dispatcher_test?sslmode=disable' \
  go test -run TestStoreLifecycle -count=1 ./internal/store/postgres
```

The lifecycle test scopes records to a unique tenant and removes them afterward. Use
a dedicated database because it applies the embedded schema migration.

## HTTP API

| Endpoint                             | Authentication                                       | Purpose                                       |
| ------------------------------------ | ---------------------------------------------------- | --------------------------------------------- |
| `POST /webhooks/github/{tenant}`     | GitHub HMAC                                          | Verify, deduplicate, and normalize a delivery |
| `POST /webhooks/gitlab/{tenant}`     | GitLab Standard Webhooks or explicit legacy fallback | Verify, deduplicate, and normalize a delivery |
| `POST /v1/effects`                   | Admin bearer token                                   | Queue one typed preview or apply effect       |
| `GET /v1/effects/{id}`               | Admin bearer token                                   | Read effect state and provider result         |
| `POST /v1/dead-letters/{id}/requeue` | Admin bearer token                                   | Explicitly retry an investigated dead letter  |
| `GET /health/live`                   | None                                                 | Process liveness                              |
| `GET /health/ready`                  | None                                                 | PostgreSQL readiness and queue counts         |
| `GET /metrics`                       | None                                                 | Prometheus counters                           |

Effects require a caller-supplied `idempotency_key`; the service supplies `id` and
`correlation_id` when omitted. Every provider, tenant, repository, and effect kind must
match the static configuration allowlists. Apply requests are rejected while global
apply mode is disabled.

## Phase 4 boundary

Phase 4 will consume normalized workflow events, resolve active context, create an
isolated `AgentRun` for the selected role and operation, wait for its terminal state,
validate a structured `OperationResult`, and submit only its authorized proposed
effects to this runtime.

That phase is the difference between this dispatcher being a durable provider
automation service and being an agent workflow dispatcher. It should be added only
with:

- deterministic event-to-operation policy;
- pinned subject revisions and capability sets;
- blind-first independent review runs;
- a schema-versioned result envelope;
- a strict rule that provider credentials remain in this service, never in the agent
  prompt or pod;
- correlation from delivery through `AgentRun`, effect, provider reference, and
  context successor.
