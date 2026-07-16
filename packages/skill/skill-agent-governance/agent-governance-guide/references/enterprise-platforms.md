# Enterprise Platform Composition

Enterprise platforms are optional adapters around the governance and workflow ports. Azure DevOps, Bitbucket, Bitrise, AWS, and Datadog each keep ownership of native discovery, configuration, mutations, references, permissions, editions, and limitations. The shared contract owns only how independently selected providers compose.

## Ownership boundaries

| Owner                    | Owns                                                                                                                                            | Does not own                                                                 |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `azure-devops-guide`     | Boards work items, Repos and pull requests, branch policies, Pipelines, service connections, REST, service hooks, Services/Server detection     | Bitbucket source semantics, Bitrise execution, AWS policy, Datadog telemetry |
| `bitbucket-guide`        | Cloud/Data Center detection, repositories, pull requests, Pipelines where supported, deployments, checks, permissions, REST, webhooks           | Azure Boards, Bitrise, AWS, Datadog                                          |
| `bitrise-guide`          | Workflows, Pipelines, Steps, triggers, secrets, artifacts, statuses, runners, API/webhooks, Bitrise-to-AWS OIDC                                 | Repository-provider or AWS-policy semantics                                  |
| `aws-guide`              | IAM and federation, Organizations/SCPs, CloudTrail, Config, Security Hub, AWS-native evidence and break-glass-safe setup                        | CI-provider identity claims or Datadog configuration                         |
| `datadog-guide`          | CI/CD and DORA visibility, OpenTelemetry, LLM observability, audit, catalog, AWS integration, cloud security, telemetry privacy                 | Workflow identity or authoritative source/build/cloud state                  |
| `agent-governance-guide` | Capability negotiation, transactional configuration, evidence/provider ports, authority, failure policy, conformance, and effective composition | Provider-native commands, schemas, editions, or fallback behavior            |

A cross-platform flow lives with the platform that initiates it and cross-references the receiving platform's skill by name for its side of the trust relationship: Bitrise-to-AWS federation belongs to `bitrise-guide` with `aws-guide` owning the role-trust semantics, and Datadog's AWS integration belongs to `datadog-guide` with `aws-guide` owning the account-side permissions. No skill defines another platform's native behavior.

## Capability negotiation

1. Detect product, deployment model, edition, account or tier, native version, API version, region or residency, installed extensions, and effective actor before selecting a capability.
2. Compare detected facts with a dated and version-pinned conformance baseline.
3. Return supported, unsupported, tier-restricted, extension-dependent, stale, or unknown per capability. Similar product names never establish parity.
4. Keep native references opaque. Consumers may resolve a reference through its selected provider but never parse it to infer identity or state.
5. Fail an explicitly selected unavailable provider. Never redirect a mutation or evidence publication to another platform or a local sidecar.

Cloud and self-managed products are separate variants. Azure DevOps Services does not prove Azure DevOps Server behavior. Bitbucket Cloud Pipelines, OIDC, shared configuration, and custom checks do not imply Bitbucket Data Center support.

## Transactional onboarding

Every configuration helper follows `discover → assess → propose → preview → authorize → apply → verify → record → operate`.

- **Preview** shows the exact native subjects and before/after mutations; requested actor, token, role, secret, network, and data access; external destinations; cost or ingestion effects; expected native evidence; failure and partial-apply behavior; verification probes; rollback; and drift ownership.
- **Authorize** binds approval to the preview digest, provider, subject, version, authority, and expiry. A changed preview needs new authorization.
- **Apply** is idempotent against an observed native version and idempotency key. Drift, permission expansion, and partial application stop the transaction.
- **Verify** re-reads authoritative provider state and runs an allow probe, deny or unsupported probe, evidence resolution, outage behavior, and rollback-availability check.
- **Operate** diagnoses, detects drift, updates explicit pins, rotates credentials, disables, removes, and rolls back without deleting retained evidence.

Helpers default to read-only discovery or dry-run. A state-changing operation requires an explicit apply action and authorization reference. They use provider-native APIs and configuration; fixture JSON tests the adapter contract and is not configuration to apply.

## Migration, coexistence, and uninstall

- **Migrate** by discovering authoritative native state, pinning the source and target variants, mapping only supported capabilities, previewing ownership and evidence changes, running old and new paths in a bounded verification window, and retaining a rollback point. Do not infer edition parity or rewrite opaque references.
- **Coexist** with existing policies, identities, pipelines, hooks, integrations, and telemetry by assigning one owner per mutation and evidence kind. Adopt a compatible native resource only after explicit ownership transfer; otherwise leave it foreign and compose through its native reference.
- **Uninstall** by disabling new work, verifying no dependent workflow remains, removing only owned resources in dependency order, revoking owned credentials and trust, and rerunning negative probes. Preserve foreign resources, audit history, retained evidence, and references required by retention policy.

## Credentials and data

- Prefer workload identity, OIDC, managed identity, or another provider-supported federation path that issues temporary credentials. Trust policies constrain issuer, audience, subject, repository/project/workspace, environment, branch/tag, and role session where claims exist.
- Never create a long-lived access key or other static credential by default on any platform. If federation is unavailable, return unsupported with remediation instead of silently generating one; a static credential is only an explicit, labeled, expiring exception.
- Separate discovery credentials from configuration credentials and runtime credentials. Request only the native scopes used by the selected operation.
- Keep secrets out of command arguments, previews, logs, artifacts, native references, and fixtures. Report secret identifiers and access paths, not values.
- Telemetry collection starts with metadata and identifiers. Content capture, prompts, source, tool output, personal data, and model input/output require a declared purpose, authorization, redaction, sampling, retention, residency, access, deletion, and cost policy.

## Mixed-stack composition

The enterprise fixture composes these independent coordinates:

```text
work item  = Azure Boards opaque work-item reference
source/PR  = Bitbucket Cloud or Data Center opaque commit and pull-request references
mobile CI  = Bitrise opaque build, artifact, and status references
runtime    = AWS opaque role session, CloudTrail, Config, and Security Hub references
telemetry  = Datadog opaque pipeline, deployment, trace, audit, and finding references
```

Each provider remains the source of truth for its own record. Relationships carry opaque references plus subject, immutable revision where available, origin, freshness, authority, and limitations. A correlation or trace identifier may connect evidence but never becomes workflow identity. Recovery starts in a fresh process from the provider references; it does not need the originating conversation, a YAML sidecar, or a central result store.

## Failure and conformance

- Provider outage, rate limit, stale version, missing tier, unsupported edition, invalid OIDC claim, unauthorized access, duplicate request, partial apply, drift, secret exposure, broken artifact/status linkage, and telemetry-redaction failure are explicit outcomes.
- Mandatory privileged actions fail closed on missing or invalid authority/evidence unless a scoped, authorized, expiring exception or break-glass record applies. Advisory telemetry fails visibly and states that no enforcement occurred.
- Retries are bounded and cancellable; side effects use provider idempotency or reconcile by native identity before retrying.
- Conformance pins the product or edition, tested version/API, source snapshot, capabilities, limitations, and native reference kinds. Documentation conformance does not claim a live tenant probe passed.
- Every platform claim traces to official documentation pinned with a retrieval date in the owning skill's `SOURCES.md`; a detailed capability claim additionally requires an edition/version-pinned probe before a release relies on it. A claim without a pinned source or probe is a release blocker for the owning skill.

Run `node scripts/validate-enterprise-platform-fixtures.mjs` from the guide directory or `npx moon run skill-agent-governance:test` from the repository root.
