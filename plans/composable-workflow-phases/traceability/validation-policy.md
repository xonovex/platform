# Traceability Validation Policy

Last verified: **2026-07-14**

The final documentation and validation plan must fail when any of the following is true:

1. A settled decision has no decision ID, classification or source mapping.
2. A mandatory control has no control ID, applicability rule, evidence requirement or enforcement point.
3. A legal obligation lacks an exact regulation/article candidate mapping, current-text review flag or qualified-review requirement.
4. A detailed ISO claim is based only on a public abstract without a licensed-text verification flag.
5. A platform capability lacks official documentation, a pinned tested version or a conformance probe.
6. Architectural synthesis is presented as a legal, standards or vendor requirement.
7. A crosswalk claims one-to-one equivalence, certification or compliance without an authoritative assessment.
8. A subplan task has no decision/control/source traceability.
9. A cited source ID is missing, superseded without review or outside its stated claim scope.
10. A mandatory enforcement module cannot name the policy version, decision, evidence and native enforcement point it applies.
11. A skill installation is represented as enforcement evidence.
12. A workflow result or governance module is required to use a universal YAML/JSON/file representation.
13. An exception or break-glass path lacks scope, owner, expiry, compensating controls and review.
14. Model or agent execution can replace authoritative deterministic/external evidence without an explicit rule.
15. Source freshness, platform compatibility or legal applicability is unknown but the plan claims certainty.
16. Azure DevOps Services/Server, Bitbucket Cloud/Data Center, Bitrise runner/workspace, AWS organization/account, or Datadog product/tier differences are ignored.
17. CI-to-AWS onboarding creates long-lived static credentials by default or lacks trust-claim and least-privilege validation.
18. Observability onboarding lacks data collection, redaction, retention, residency, access and cost declarations.

## Required mapping statuses

- `verified-primary-public`
- `licensed-text-required`
- `qualified-legal-review-required`
- `version-pinned-conformance-required`
- `architectural-synthesis`
- `partial-contextual-mapping`

## Release evidence

The release candidate must include generated reports proving:

- all IDs resolve;
- all numbered subplan tasks are mapped;
- no orphan decision/control/source IDs remain;
- source versions and retrieval dates are recorded;
- platform claims have conformance results;
- legal and licensed-standard caveats remain visible;
- stale sources or changed platform behavior block release or are explicitly waived.
