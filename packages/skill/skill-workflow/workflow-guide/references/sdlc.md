# Composing Operations Across Roles

Order work as explicit operations with handoffs at the cold boundaries between them.
The ordering is a playbook, not a lifecycle: no command implies the next one. The
scenarios are the frozen families in [contract.md](contract.md).

## Xonovex platform

A catalog change runs Create or Revise with the authoring guides, then Review before
it lands. A multi-step change runs Decide for the open questions and **plan-guide**
for the plan itself. Delivery is Review, then Publish through **pull-request-guide**
and **versioning-guide**.

## Drodan and CruiseReviews

An increment starts at Decide and **user-stories-guide**, with **bdd-guide** turning
the agreed examples into checks. Build runs Execute against the framework guides;
Validate evaluates the acceptance criteria independently. Editorial work runs Create,
Revise, then Publish.

## Native and game engine

A subsystem or renderer feature runs Execute with the runtime guides. A reported
crash runs Execute to reproduce, then Abandon with a retry boundary when the repro
does not hold. Portability work runs Execute against the platform layer.

## Infrastructure and operations

A deployable service runs Create for its manifests and Publish to the GitOps
repository. Automation runs Execute and Validate. Parallel work uses the workspace
operations: create to isolate, merge to integrate, cleanup as a separate step.

## Invariants

1. Pin every provider-native artifact to its exact revision when the provider
   exposes one.
2. Run independent reviews or validations as separate invocations against the same
   subject revision.
3. Preview protected effects before apply.
4. Keep integration and cleanup separate.
5. Carry only the context the receiving role needs, in the shape
   [handoffs.md](handoffs.md) defines.

Authority, evidence, and fetched-content rules live in [governance.md](governance.md).

## Provider Mapping

Use one provider skill explicitly; do not infer it from a reference's shape.

| Concern        | GitHub                                      | GitLab                                               |
| -------------- | ------------------------------------------- | ---------------------------------------------------- |
| Ticket         | Repository issue: repo + number/node ID/URL | Project issue/work item: project + IID/global ID/URL |
| Kanban card    | ProjectV2 item with its own item ID         | Issue card rendered from the issue                   |
| Workflow stage | Project field/option ID, usually Status     | List-defining issue attribute                        |
| Change         | Pull request + HEAD SHA                     | Merge request + HEAD SHA/diff refs                   |
| Context        | Append-only comment on issue or PR          | Append-only Note on issue or MR                      |

Create or update tickets and move kanban state through Execute with `--effect
preview`, then the same operation with `--effect apply`. Assignments, access control,
approvals, status transitions, notifications, and scheduling stay owned by the
provider.
