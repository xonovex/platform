# Multi-Role SDLC Composition

Compose a software-delivery lifecycle as explicit handoffs between atomic operations.
The playbook orders work; it does not make any command imply the next one.

| Phase              | Primary roles                                   | Durable subject or evidence                               | Typical capability and operation                                                              |
| ------------------ | ----------------------------------------------- | --------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| Opportunity        | Product manager, stakeholders                   | Problem, outcome, baseline, constraints                   | **product-discovery-guide** with Create, Review, Decide                                       |
| User discovery     | Product owner, UX researcher                    | Research plan, observations, findings                     | **ux-research-guide** with Create, Review                                                     |
| Definition         | Product owner, developer, QA                    | User needs, vertical stories, acceptance examples         | **user-stories-guide**, **bdd-guide**                                                         |
| Design             | UX, accessibility, engineering                  | Tested journeys, design reference, accessibility criteria | **ux-research-guide**, **figma-guide**, **accessibility-guide**                               |
| Architecture       | Architect, technical lead, specialists          | ADRs, boundaries, non-functional criteria                 | Subject-specific architecture capabilities with Review, Decide                                |
| Planning           | Delivery lead, developers, QA                   | Plan, child plans, Definition of Done                     | **plan-guide** operations, then Publish when persistence is needed                            |
| Work setup         | Developer                                       | Exact source revision and isolated workspace              | Workspace create                                                                              |
| Implementation     | Developer                                       | Code, tests, commits, implementation evidence             | Plan Continue or Execute with language, framework, TDD, and testing capabilities              |
| Review             | Reviewer, maintainer                            | Revision-pinned findings and CI evidence                  | **code-review-guide**, **pull-request-guide**, Review                                         |
| Quality validation | QA, product owner, UX                           | Criterion-level release-candidate evidence                | **exploratory-testing-guide**, **bdd-guide**, accessibility and platform checks with Validate |
| Security design    | Architect, security specialist                  | Threats, mitigations, verification evidence               | **threat-modeling-guide** with Review, Revise, Validate                                       |
| Release            | Release manager, platform, accountable approver | Version, protected-gate evidence, deployment reference    | Versioning and provider capabilities; Decide remains descriptive                              |
| Operations         | Incident commander, operations, communications  | Incident timeline, effects, recovery evidence             | **incident-response-guide**, observability capabilities, Execute                              |
| Learning           | Product, UX, engineering, QA                    | Outcome evidence and next opportunity                     | Review production evidence, then start a new explicit operation                               |

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

## Example Vertical Slice

1. Product discovery frames a measurable checkout problem without prescribing a
   solution.
2. UX research tests the riskiest user assumptions and returns evidence-linked
   findings.
3. Product, development, and QA derive a vertical story plus concrete acceptance
   examples.
4. Architecture and threat-model reviews inspect the same definition revision.
5. Planning creates and independently critiques an implementation plan.
6. Workspace create isolates the exact source revision.
7. Plan Continue or Execute applies one bounded target using the selected
   implementation and test capabilities.
8. Code review and QA validation return separate evidence entries against the same
   candidate revision.
9. Decide records a release recommendation without approving deployment.
10. The protected provider gate authorizes release; concrete provider capabilities
    deploy and verify it.
11. Operations evidence feeds a new product or engineering operation instead of
    mutating the completed workflow retrospectively.

Assignments, access control, approvals, status transitions, notifications, scheduling,
and SLAs remain owned by the selected provider. Preserve those native references
without recreating them inside this skill.
