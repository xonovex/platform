# Governed Learning Promotion

## Gather candidates from operational evidence

Extract reusable candidates from the current session and, when provided, lifecycle results, onboarding outcomes, policy denials, incidents, exceptions, break-glass reviews, drift, rollbacks, support cases, and module failures. Preserve the exact source reference/version, scope, observed outcome, discovery, proposed lesson, affected owners, confidence, conflicts, privacy limits, and expiry.

Do not infer a causal lesson from one metric or event without checking alternative explanations, data quality, selection bias, incentives, and whether the observation recurs beyond its original context.

## Route to one owner

Route a small project-specific lesson to the nearest `AGENTS.md`; reusable domain guidance to the existing owning skill; organization policy to its policy owner; executable behavior to its module owner. Create a new skill only when 3–7 strong reusable lessons lack an owner.

Keep the direct-apply default for ordinary reviewable `AGENTS.md` and skill edits. Direct application changes the working tree; it does not publish, install, enable, enforce, or globally promote the result.

## Gate managed or executable promotion

When the target is managed, organization-wide, executable, enforcing, configuration-changing, or privileged:

1. preview exact content/behavior, scope, permissions, data flows, conflicts, affected profiles, evidence, failure behavior, and rollback;
2. require the target owner's review and exact-scope authorization;
3. version the target and preserve candidate → decision → version provenance;
4. validate and canary representative contexts where executable;
5. measure the intended outcome plus false positives, bypasses, user friction, privacy, and rollback criteria;
6. promote gradually, verify authoritative effective state, or roll back/remove.

Never auto-promote one denial, incident, exception, drift event, metric, model inference, or reflection into global instructions or enforcement.

## Output a promotion record

Record candidate/source references, target owner and version, scope, reviewer/authority, decision, conflicts, validation, canary, effectiveness measures, privacy/retention, apply/effective-state references, rollback/removal path, and review/expiry. Rejected and rolled-back candidates remain evidence; they do not silently re-enter the queue.
