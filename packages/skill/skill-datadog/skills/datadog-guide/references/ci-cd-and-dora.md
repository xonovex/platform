# CI/CD Visibility and DORA

## Correlation contract

Define canonical service, environment, version, repository, commit SHA, pipeline/provider, pipeline/run/job IDs, artifact digest, deployment ID/time/status, and source/provider URLs before ingestion. Preserve the source-host, CI, artifact, deployment, and Datadog references separately.

A Datadog pipeline or deployment event correlates evidence; it is not the workflow result identity and does not replace the native build/deployment record. Resolve exact source and target revisions at their owners.

## CI Visibility

Discover supported CI provider/instrumentation, test visibility, agentless/agent path, API/app keys, tags, logs/artifacts, retries, queue/execution timing, repository metadata, and data volume. Keep untrusted pull-request jobs separate from credentials and target-changing work.

Preview fields/content, tags, status mapping, network destination, redaction, sampling, retention, access, and cost. Test duplicate/retried jobs, cancelled/skipped/timed-out states, missing source metadata, and provider outage without coercing them to success.

## CD Visibility and DORA

Define what constitutes a deployment, production environment, service, change, failure, and recovery for the organization. Preserve mapping/version and limitations for deployment frequency, lead time, change failure rate, and recovery time.

Test late, duplicate, rolled-back, failed, partial, multi-service, and out-of-band deployments. An incomplete source/CI/deployment integration makes metrics partial; report that limitation instead of imputing events.

Use exact native references to investigate a metric. Aggregates and dashboards are views, not immutable evidence.
