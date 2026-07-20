# Abandon

## Goal

Stop work on one subject while preserving its reason, partial results, evidence, cleanup state, and retry boundary without deleting or rewriting the subject.

## Procedure

1. Resolve the exact subject, revision, reason, and explicit selections.
2. Load only explicitly selected or unambiguous domain, abandonment, cleanup, and provider capabilities.
3. Let the selected provider interpret opaque references and revisions.
4. Record the reason, partial result, evidence, cleanup state, and retry boundary. Preserve the source by default.
5. Return the record inline or persist it only to an explicit destination. Apply cleanup only when explicitly scoped and authorized.

Abandonment is descriptive unless a selected provider performs an explicitly authorized state update.

## Error handling

- Stop on a missing subject or reason.
- Report ambiguous providers and unavailable explicit capabilities without substitution.
- Preserve state and return a preview when destructive cleanup lacks exact scope or authorization.
