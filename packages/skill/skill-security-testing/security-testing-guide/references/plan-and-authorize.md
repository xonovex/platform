# Plan and Authorize

## Rules of Engagement

Before active testing, record:

```text
Authorizing owner and evidence:
In-scope targets, accounts, data, environments, and revisions:
Explicit exclusions:
Time window and source identities:
Allowed and prohibited techniques:
Rate, load, and data limits:
Monitoring, contacts, escalation, and stop conditions:
Backup, cleanup, and recovery:
Evidence protection, retention, and disclosure route:
```

Resolve target ownership and dependencies. Use isolated environments and synthetic
data where they can answer the question. For production-only risks, narrow the probe
and obtain explicit production authorization.

## Verification Basis

Trace coverage to the current threat model, assets, trust boundaries, abuse scenarios,
architecture and configuration changes, historical vulnerabilities, applicable
regulation, and a versioned verification standard suited to the product.

Select requirements by applicability and risk. Record not-applicable rationale and
standard versions because identifiers and coverage evolve. Security testing should
also feed secure-development and test-strategy evidence across the lifecycle.

Pause when the target, authorization, safe limits, sensitive-data handling, or
incident contact is ambiguous. Discovery of an active incident follows the incident
response path.
