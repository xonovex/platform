# Integration and coverage

## Contents

- Checking the integration between items
- Plan criteria that no work item covers

## Checking the integration between items

Items that pass alone can still fail together. For a package with more than one work item:

1. List the seams from the work files: each `depends_on` edge, each item that produces something another consumes (a contract, a schema, a message, a build artifact, a configuration value), and each criterion of the plan that spans items.
2. Check each seam against the code of both sides at the revisions in hand: producer and consumer agree on names, shapes, versions and order of deployment.
3. Run the end-to-end check that crosses the seams when one exists, and say which one. Where none exists, the seam is checked by reading and the entry says so.
4. A seam that fails or cannot be checked becomes a blocker or an open point with an owner. It also keeps the criteria it supports from being met.

A package with one work item has no seams. Say that in `# Integration`.

## Plan criteria that no work item covers

Map each success criterion to the work item criteria that support it. A plan criterion that no work item criterion covers is open, owned by the plan's owners, and the gap is reported. It points to a missing work item, a plan the work no longer matches, or a criterion that was never checkable; the close-out says which and leaves the choice to **revising-plan**.
