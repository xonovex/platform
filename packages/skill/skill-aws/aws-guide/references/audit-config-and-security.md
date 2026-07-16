# CloudTrail, Config, and Security Hub

## CloudTrail

Discover organization/account trails, Lake event data stores, regions, management/data/network activity events, selectors, destinations, encryption, log-file validation, retention, access, alerting, and gaps. Preserve account, region, event ID/time, event source/name, actor/session, resource, request/result classification, trail/event-store reference, and policy/config version.

Minimize copied event data. Secrets, request parameters, resource contents, personal data, and cross-account information remain governed by access, redaction, retention, residency, and deletion rules.

## AWS Config conformance packs

Pin the pack template/version, parameters, delivery channel/aggregator, accounts/regions/resources, managed/custom rules, remediation ownership, evaluation time, and result references. Preview rule/remediation permissions, recording cost, unsupported resources/regions, and rollback.

A pack is a collection of rules and remediation guidance, not a compliance certificate. Preserve compliant/noncompliant/not-applicable/insufficient-data states without coercion.

## Security Hub

Discover enabled regions, organization/admin configuration, standards, controls, integrations, finding aggregation, automation rules, workflow/status ownership, suppression/archive policy, retention, and downstream notifications. Preserve finding ARN/ID, product, account/region/resource, standard/control version, severity, workflow state, timestamps, and native URL/reference.

Security Hub findings do not replace CloudTrail events or authoritative target state. Correlate them by opaque references and resource identity while keeping each source independent.

## Verification

Re-read authoritative service configuration; generate safe positive/negative evidence where practical; test region/account gaps, delayed/unavailable service, duplicate findings/events, retention/access, and provider outage; verify rollback without deleting retained evidence.
