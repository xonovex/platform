# Failure and Recovery

## Observability and Response

Verify that telemetry can distinguish user impact, dependency failure, capacity,
deployment change, security signals, and data or queue health. For each actionable
alert, identify condition, noise evidence, severity, owner, notification path,
dashboard, runbook, escalation, and expected response.

Runbooks cover repeatable procedures; playbooks guide diagnosis under uncertainty.
Test both with the actual roles, access, tools, time pressure, and communication paths.
Confirm that audit logs and sensitive telemetry have appropriate access and retention.

## Failure Exercises

Use architecture, threat, incident, and near-miss evidence to select scenarios:

- dependency timeout, malformed response, quota, or contract change;
- overload, resource exhaustion, queue growth, and regional or zone loss;
- bad deployment, configuration, schema, data, or certificate change;
- credential loss, compromise, abuse, and security-control failure;
- operator error, unavailable specialist, and communication failure.

For each, define detection, containment, degraded behavior, decision authority,
recovery procedure, recovery-time and recovery-point objectives, data-integrity
verification, and learning capture.

Backups require protected retention plus periodic restore and integrity exercises.
Record exercise subject, environment, result, duration, gaps, and limitations.
