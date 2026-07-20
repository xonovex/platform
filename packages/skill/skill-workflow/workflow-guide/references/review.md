# Review

## Goal

Evaluate one exact subject against explicit criteria without modifying it. Produce evidence-linked findings and return them inline unless persistence is explicitly requested.

## Procedure

1. Resolve the exact subject, revision, criteria, and explicit selections. Infer kind or provider only when unambiguous.
2. Load only explicitly selected or unambiguous review, domain, method, and provider capabilities.
3. Let the selected provider read opaque references and revisions.
4. Report findings, evidence, severity when relevant, and uncertainty. Do not revise the subject.
5. Return findings inline or persist them only to an explicit destination.

Review produces judgment and findings; it does not provide criterion-by-criterion validation, revise the subject, or grant authority.

## Error handling

- Stop before reviewing a mutable referenced subject without an exact revision.
- Request criteria or an unambiguous method when no review standard is available.
- Report ambiguous providers and unavailable explicit capabilities without fallback.
