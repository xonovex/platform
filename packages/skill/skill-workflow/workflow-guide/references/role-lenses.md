# Role Lenses

A role lens illustrates how the same operations may be composed. It is not a permission model, mandatory handoff, stage sequence, or role-specific command set.

Examples:

- Product or UX work may use `create → review → revise → decide` with a product, research, interaction, content, or accessibility perspective.
- Development may use `create → review → revise → execute → validate → publish` with independently selected implementation, testing, and provider capabilities.
- QA may create or review test material, execute checks, validate behavior, and publish evidence without owning source-review approval.
- A developer reviewer may use `review → publish` to evaluate an exact change and publish a disposition through the selected source-control provider.

Role names never imply authority, required order, hidden criteria, provider effects, or a different operation contract.
