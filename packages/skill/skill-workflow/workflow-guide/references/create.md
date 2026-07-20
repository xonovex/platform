# Create

## Goal

Create one new result from the supplied subject and references without changing the source. Return it inline unless persistence is explicitly requested.

## Procedure

1. Resolve the subject and explicit selections. Infer kind or provider only when unambiguous; report the inference and stop on ambiguity.
2. Load only explicitly selected or unambiguous domain, method, and provider capabilities. Stop and name any unavailable explicit capability.
3. Let the selected provider interpret opaque references and revisions.
4. Create a new result that satisfies the criteria without mutating the subject.
5. Return the result inline. Persist only to an explicit destination and return its native locator and revision when supported.

Kind, perspective, method, capability, and provider remain independent. Trigger, executor, identity, and maturity do not change the operation. Do not construct a preset, central resolver, registry, or default provider.

## Error handling

- Stop on a missing subject or criteria required by the selected method.
- Report ambiguous kind or provider choices instead of guessing.
- Identify an unavailable explicit capability without substituting another.
- Keep the result inline when no destination provider can be resolved.
