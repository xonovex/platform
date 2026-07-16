# plan-update: Publish Planning Status and Validation

Refresh one Planning result from current implementation and evidence without changing the planned scope.

## Core workflow

1. Resolve the provider context, opaque Planning reference, exact native revision, child relationships, and Development/validation evidence.
2. Compare every task and success criterion with exact subject/workspace revisions and evidence; do not infer completion from conversation or one green check.
3. Set completed, pending, blocked, and validation states with concrete evidence references and limitations. Preserve unavailable validation categories explicitly.
4. Reconcile child status and cumulative completion without erasing independently published results.
5. Publish a new Planning native revision with updated status/evidence and supersession; do not transfer approval when the update changes plan meaning.
6. Return the new opaque reference/revision and remaining work.

Apply [early-lifecycle-contracts.md](early-lifecycle-contracts.md). Local frontmatter edits are one provider representation, not a universal requirement.
