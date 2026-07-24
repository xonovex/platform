# Multi-Role SDLC Composition

Compose a software-delivery lifecycle as explicit handoffs between atomic operations.
The playbook orders work; it does not make any command imply the next one.

| Phase                   | Primary roles                                                  | Durable subject or evidence                                     | Typical capability and operation                                                                                  |
| ----------------------- | -------------------------------------------------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Opportunity             | Product manager, product owner, stakeholders                   | Problem, outcome, baseline, constraints                         | **product-discovery-guide** with Create, Review, Decide                                                           |
| User discovery          | UX researcher, product designer, product owner                 | Research plan, observations, findings                           | **ux-research-guide** with Create, Review                                                                         |
| Measurement design      | Product analyst, product manager, privacy and data specialists | Measurement model, metric contracts, guardrails                 | **product-analytics-guide** with Create, Review                                                                   |
| Definition              | Product owner, business analyst, developer, QA, UX             | User needs, vertical stories, acceptance examples               | **user-stories-guide**, **bdd-guide**                                                                             |
| Experience design       | Product designer, content designer, accessibility, engineering | Tested journeys, states, rationale, accessibility criteria      | **ux-design-guide**, **ux-research-guide**, **accessibility-guide**; **figma-guide** only for provider operations |
| Architecture            | Architect, technical lead, security, operations, data          | Quality scenarios, tradeoffs, ADRs, boundaries                  | **architecture-evaluation-guide**, **adr-guide**, subject-specific architecture capabilities with Review, Decide  |
| Security design         | Architect, security specialist, developers, operations         | Threats, mitigations, verification requirements                 | **threat-modeling-guide** with Review, Revise, Validate                                                           |
| Planning                | Delivery lead, developers, QA, product, design                 | Plan, child plans, Definition of Done                           | **plan-guide** operations, then Publish when persistence is needed                                                |
| Work setup              | Developer                                                      | Exact source revision and isolated workspace                    | Workspace create                                                                                                  |
| Implementation          | Developers, data and platform engineers, content               | Code, tests, configuration, content, commits, evidence          | Plan Continue or Execute with language, framework, TDD, testing, infrastructure, and content capabilities         |
| Review                  | Reviewer, maintainer, specialists                              | Revision-pinned findings and CI evidence                        | **code-review-guide**, **pull-request-guide**, Review                                                             |
| Test strategy           | Test lead, QA, developers, product, specialists                | Risk model, coverage allocation, environments, completion rules | **test-strategy-guide** with Create, Review, Revise                                                               |
| Quality validation      | QA, product owner, UX, accessibility, performance specialists  | Criterion-level release-candidate evidence                      | **exploratory-testing-guide**, **bdd-guide**, accessibility, performance, compatibility, and platform validation  |
| Security verification   | Security tester, developers, QA                                | Authorized assessment, findings, remediation and retest         | **security-testing-guide** with Create, Validate, Revise                                                          |
| Operational readiness   | Service owner, SRE, operations, support, security              | SLOs, ownership, runbooks, exercises, readiness findings        | **operational-readiness-guide** with Review, Validate, Revise                                                     |
| Release readiness       | Release manager, product, QA, security, operations             | Pinned candidate, evidence matrix, rollout and recovery         | **release-readiness-guide** with Review, Decide                                                                   |
| Approval and deployment | Accountable approver, release operator, platform               | Protected-gate evidence, deployment and verification references | Versioning and provider capabilities; Decide remains descriptive and provider approval remains external           |
| Operations              | Service owner, SRE, operations, support                        | Service objectives, telemetry, support and change evidence      | **operational-readiness-guide** plus concrete observability and provider capabilities                             |
| Incident response       | Incident commander, operations, engineering, communications    | Timeline, effects, mitigation, recovery and learning            | **incident-response-guide**, observability capabilities, Execute                                                  |
| Outcome learning        | Product analyst, product, UX, engineering, QA                  | Outcome evidence, experiment result, next opportunity           | **product-analytics-guide**, **product-discovery-guide**, **ux-research-guide**, then a new explicit operation    |

## Handoff Invariants

1. Pin every provider-native artifact to its exact revision when the provider exposes
   one.
2. Preserve parent, supersedes, related, and evidence references in the Markdown
   handoff.
3. Run organizationally independent reviews or validations as separate invocations
   against the same subject revision.
4. Keep descriptive decisions separate from provider-enforced approvals.
5. Preview protected effects before apply; require concurrency and retry protection
   when supported.
6. Publish inline results only through a separate explicit operation.
7. Keep integration and cleanup separate.
8. Keep role accountability in the provider while every handoff retains the same
   subject revision, evidence references, limitations, and unresolved risk.

## Role Continuity

- Product management and ownership carry the problem, outcome, scope, and accountable
  product decision from opportunity through outcome learning.
- Research, design, content, and accessibility carry user evidence and whole-journey
  behavior into definition, implementation review, and validation.
- Architecture, development, data, and platform roles carry quality scenarios,
  decisions, source revisions, implementation evidence, and operational consequences.
- QA and specialist testers carry the risk model and independent criterion-level
  evidence into the exact release candidate.
- Security carries threats into verification requirements, findings, remediation,
  retest, and residual-risk ownership.
- Service ownership, SRE, operations, support, and communications carry production
  objectives, readiness, change evidence, incidents, and learning.
- Delivery and release roles coordinate sequence and evidence but do not absorb
  product, technical, quality, security, operational, or provider authority.

One person may perform several roles and one role may be distributed across teams. The
workflow shape does not change: select capabilities for the operation, pin the subject
and revision, produce traceable evidence, and preserve accountable provider decisions.

## Example Vertical Slice

1. Product discovery frames a measurable checkout problem without prescribing a
   solution.
2. UX research tests the riskiest user assumptions and returns evidence-linked
   findings.
3. UX design compares and tests interaction alternatives; product analytics defines
   the outcome measures and guardrails.
4. Product, design, development, and QA derive a vertical story plus concrete
   acceptance examples.
5. Architecture evaluation and threat modeling inspect the same definition revision.
6. Planning creates and independently critiques an implementation plan.
7. Workspace create isolates the exact source revision.
8. Plan Continue or Execute applies one bounded target using the selected
   implementation and test capabilities.
9. Test strategy allocates risk coverage; code review, QA, accessibility, performance,
   and security testing return separate evidence against the same candidate.
10. Operational readiness verifies that the service and owning team can operate and
    recover it.
11. Release readiness integrates the pinned candidate, evidence, staged rollout,
    guardrails, and recovery into a recommendation.
12. Decide records that recommendation without approving deployment.
13. The protected provider gate authorizes release; concrete provider capabilities
    deploy and verify the exact artifact.
14. Operations and product analytics collect service and outcome evidence; incidents
    follow incident response.
15. Learning starts a new product or engineering operation instead of mutating the
    completed workflow retrospectively.

Assignments, access control, approvals, status transitions, notifications, scheduling,
and SLAs remain owned by the selected provider. Preserve those native references
without recreating them inside this skill.
