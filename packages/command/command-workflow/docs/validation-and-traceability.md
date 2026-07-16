# Validation and Traceability

Release baseline: **6.0.0**. The repository maintaining this plugin keeps the authoritative design-time traceability record — a source registry, decision-to-source matrix, control and obligation crosswalk, platform capability matrix, subplan task traceability, and validation policy — and validates it with a repository-level release validator (the `release-validate` moon task) requiring all **94 sources**, **39 decisions**, **45 controls/obligations**, **14 platform entries**, **11 subplans**, and **150 numbered tasks** to resolve, alongside marketplace/package lockstep, plugin dependencies, migration coverage, diagram semantics, and prohibited public claims. This package validates its own contents — documentation links, manifest dependency consistency, migration coverage, the validation matrix below, and prohibited claims — via `scripts/validate-documentation.mjs`, which runs without repository context.

## Mapping status and caveats

Traceability organizes evidence and review; it does not establish one-to-one equivalence, certification, conformity, vendor conformance, or legal compliance.

- Legal rows require an applicability result, current consolidated text, jurisdiction/role/date analysis, and qualified review.
- Detailed ISO clause claims require licensed full-text verification; public abstracts support only their stated high-level scope.
- Platform rows require the detected product surface, edition/tier, tested version/date, and live or pinned conformance evidence. Documentation-only fixtures remain runtime-unverified.
- Architectural synthesis remains labeled as synthesis even when sources support the concern.
- Unknown freshness, applicability, capability, or evidence blocks a certainty claim or remains an explicit limitation.

## Test and evaluation matrix

| Area                      | Required positive, negative, and fault coverage                                                                                                                                                                                                     | Executable evidence                                                               |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| Workflow contracts        | Profiles, topology, storage-neutral result providers, exact revisions, lifecycle completion, stale evidence, development, delivery, review, QA, acceptance, integration, transition, release, observation, incidents, corrective action, retirement | `skill-workflow:test` fixtures and workflow evals                                 |
| Governance contracts      | Adoption modes, profiles, authority zones, executor classes, modules, policy outcomes, provider ports, evidence origin, exceptions, break glass, data, telemetry, catalogs, learning, operations                                                    | `skill-agent-governance:test` conformance and governance-operation fixtures       |
| Harness adapters          | Claude Code, Codex, Kiro, Copilot CLI/cloud, Pi, OpenCode; version/date, unsupported/experimental features, trust, native scopes, blocking/context, concurrency/order, diagnostics, disable, rollback, drift                                        | Harness matrices, per-harness trigger/output evals, and harness fixture validator |
| Module execution          | Deterministic scripts, bounded model evaluators, bounded agents, human/external execution, schema validation, budgets, maximum recursion, authority attenuation, cancellation, kill behavior                                                        | Module-template and walking-skeleton fixtures                                     |
| Onboarding                | Discover, assess, recommend, preview, consent, idempotent apply, verify, record, diagnose, update, disable, remove, rollback, and drift                                                                                                             | Harness, external-enforcement, enterprise-platform, and walking-skeleton fixtures |
| External enforcement      | GitHub reusable workflows/actions/rulesets/environments; GitLab components/pipeline policies/compliance frameworks; policy decision/enforcement separation; outage and bypass                                                                       | GitHub/GitLab evals and external-enforcement fixtures                             |
| Enterprise providers      | Azure DevOps Services/Server, Bitbucket Cloud/Data Center, Bitrise hosted/self-hosted and OIDC, AWS accounts/regions/policy layers, Datadog site/tier/data controls; mixed-stack references                                                         | Five enterprise skill evals and enterprise-platform fixtures                      |
| Reliability and security  | Concurrency, ordering, idempotency, retries, partial failure, provider outage, secret exposure, provenance, permission expansion, telemetry redaction/retention/access/cost                                                                         | Governance validators plus domain skill evals                                     |
| Documentation and release | IDs, task rows, legal/ISO flags, native capability limitations, link resolution, static claim guards, diagram sources/assets, 11-subplan exception, version/manifest/lockfile alignment, migration/rollback                                         | `command-workflow:test`, diagram build, format/lint/typecheck/build/test          |

Trigger evals use realistic positive and near-miss queries. Output evals use observable binary assertions and preserve documentation-only versus live evidence. Passing an eval does not turn advisory skill content into enforcement proof.

## Release commands

```bash
npx moon run skill-workflow:test skill-agent-governance:test command-workflow:test
npx moon run diagram-agent-workflow:graph-build
npm run fmt:check
npm run lint
npm run typecheck
npm run build
npm run test
```

The static validator rejects claims of cross-harness parity, skill installation as enforcement, required universal workflow YAML, hidden agent launch, or automatic compliance. It accepts explicit statements that a feature is unsupported, partial, experimental, advisory, or requires separate native enforcement.
