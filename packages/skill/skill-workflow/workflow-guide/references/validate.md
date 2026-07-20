# Validate

## Goal

Evaluate one exact subject against explicit criteria, return reproducible evidence for every criterion, and remain read-only with respect to the subject.

## Procedure

1. Resolve the exact subject, revision, criteria, supporting evidence, and explicit selections.
2. Load only explicitly selected or unambiguous validation, domain, method, and provider capabilities.
3. Let the selected provider interpret opaque references and revisions.
4. Report each criterion as pass, fail, or blocked with reproducible evidence and freshness.
5. Return evidence inline or persist it only to an explicit destination.

Validation does not revise, accept, publish, or otherwise mutate the subject. Trigger, executor, identity, and maturity do not change its contract.

## Error handling

- Stop on missing criteria or an unpinned mutable subject.
- Mark an unevaluable criterion blocked and identify the missing evidence.
- Report ambiguous providers and unavailable explicit capabilities without fallback.
