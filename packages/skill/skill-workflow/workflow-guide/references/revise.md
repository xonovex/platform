# Revise

## Goal

Produce one traceable new revision from explicit feedback while preserving the source, feedback provenance, and unresolved items.

## Procedure

1. Resolve the subject, exact source revision, feedback, and explicit selections. Stop on ambiguous kind or provider inference.
2. Load only explicitly selected or unambiguous revision, domain, method, and provider capabilities.
3. Let the selected provider interpret opaque subjects, feedback, references, and revisions.
4. Apply feedback to a new result. Record applied, deferred, and conflicting feedback without rewriting source history.
5. Return the revision inline or persist it only to an explicit destination.

Kind, perspective, method, capability, and provider remain independent. Trigger, executor, identity, and maturity do not alter revision semantics.

## Error handling

- Stop on missing feedback or an unpinned mutable source.
- Report conflicting feedback and request a decision instead of guessing.
- Report ambiguous providers and unavailable explicit capabilities without substitution.
