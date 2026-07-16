# plan-accept: Approve an Exact Planning Revision

Record the accountable actor's approval against one exact Planning revision. This is a terminal authority action for the revision; do not revise, expand, or implement in the same operation.

## Core workflow

1. Resolve the Planning provider context, opaque native reference, exact revision, critiques, source evidence, and profile requirements.
2. Sanity-check non-empty `skills_to_consult`, reconciled dependencies, validation and success criteria, unresolved feedback, evidence freshness, and provider capabilities.
3. Verify the actor and authority reference required by the profile. A model may summarize but cannot impersonate or infer approval.
4. Obtain accept, reject, or conditional status with conditions, expiry/review triggers, and rationale.
5. Publish the status decision against the exact revision and return its native reference, authority/evidence references, and next capability.

Apply [early-lifecycle-contracts.md](early-lifecycle-contracts.md). Approval of revision N does not transfer to revision N+1.
