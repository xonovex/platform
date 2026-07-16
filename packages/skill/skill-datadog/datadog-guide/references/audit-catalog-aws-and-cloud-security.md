# Audit, Catalog, AWS, and Cloud Security

## Audit Trail

Discover product/tier, event scope, actors, configuration/admin actions, retention, access roles, archive/export, notifications, and gaps. Preserve event ID/time, actor, action, target, organization/site, result, configuration version, and native reference. Audit events support accountability but do not replace the authoritative target's state or AWS CloudTrail.

## Software Catalog

Declare the source of truth for service identity, owner, lifecycle, contacts, repositories, dependencies, environments, and metadata schema. Preview synchronization direction, conflict behavior, permissions, data exposure, and deletion. Do not let an observed telemetry tag silently overwrite authoritative ownership.

## AWS integration

Preview Datadog site/org, AWS accounts/regions, IAM role/trust/external ID, permissions, resource collection, metrics/logs/traces/findings, filters/tags, agentless/agent paths, network/data volume, cost, and rollback. Prefer a dedicated least-privilege role and verify another account/action cannot use it.

Keep Datadog AWS integration evidence separate from IAM/CloudTrail/Config/Security Hub source references. Cross-link by account/region/resource and opaque native references.

## Cloud Security

Detect supported products/resources, agentless/agent/runtime mode, enabled rules/standards, finding ownership, severity/workflow, suppression, retention, access, and remediation integration. Preserve finding identity, resource/account/region, rule/version, evidence time, state, limitations, and source-cloud reference.

Findings are provider observations, not automatic compliance or permission decisions. Verify source coverage, false-positive/suppression workflow, stale/resolved state, duplicate findings, outage/delay, unauthorized access, and rollback.
